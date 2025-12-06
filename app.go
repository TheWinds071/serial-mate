package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.bug.st/serial"
)

// ConnectionType 定义连接类型
type ConnectionType string

const (
	TypeSerial    ConnectionType = "SERIAL"
	TypeTcpClient ConnectionType = "TCP_CLIENT"
	TypeTcpServer ConnectionType = "TCP_SERVER"
	TypeUdp       ConnectionType = "UDP"
)

// App struct
type App struct {
	ctx          context.Context
	mutex        sync.Mutex
	connType     ConnectionType
	isConnected  bool
	readStopChan chan struct{}

	// 串口资源
	serialPort serial.Port

	// 网络资源
	netConn     net.Conn       // 用于 TCP Client, active TCP Server conn
	netListener net.Listener   // 用于 TCP Server
	udpConn     net.PacketConn // 用于 UDP
	udpRemote   net.Addr       // UDP 远程地址 (用于发送)
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// 1. 获取串口列表
func (a *App) GetSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return []string{}, nil
	}
	return ports, nil
}

// --- 连接逻辑封装 ---

// OpenSerial 打开串口
func (a *App) OpenSerial(portName string, baudRate int, dataBits int, stopBits int, parityName string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isConnected {
		return "Already connected"
	}

	var parity serial.Parity
	switch parityName {
	case "None":
		parity = serial.NoParity
	case "Odd":
		parity = serial.OddParity
	case "Even":
		parity = serial.EvenParity
	case "Mark":
		parity = serial.MarkParity
	case "Space":
		parity = serial.SpaceParity
	default:
		parity = serial.NoParity
	}

	var stop serial.StopBits
	switch stopBits {
	case 1:
		stop = serial.OneStopBit
	case 15:
		stop = serial.OnePointFiveStopBits
	case 2:
		stop = serial.TwoStopBits
	default:
		stop = serial.OneStopBit
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: dataBits,
		Parity:   parity,
		StopBits: stop,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	port.SetMode(mode)
	port.SetDTR(true)
	port.SetRTS(true)

	a.serialPort = port
	a.connType = TypeSerial
	a.startReadLoop(port) // 启动通用读取循环

	return "Success"
}

// OpenTcpClient 连接 TCP 服务端
func (a *App) OpenTcpClient(ip string, port string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isConnected {
		return "Already connected"
	}

	address := net.JoinHostPort(ip, port)
	// 设置超时，防止界面卡死
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return fmt.Sprintf("Connect error: %v", err)
	}

	a.netConn = conn
	a.connType = TypeTcpClient
	a.startReadLoop(conn)

	return "Success"
}

// OpenTcpServer 开启 TCP 服务端
func (a *App) OpenTcpServer(port string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isConnected {
		return "Already connected"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Sprintf("Listen error: %v", err)
	}

	a.netListener = listener
	a.connType = TypeTcpServer
	a.isConnected = true
	a.readStopChan = make(chan struct{})

	// 启动监听协程
	go func() {
		for {
			select {
			case <-a.readStopChan:
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					return
				}

				a.mutex.Lock()
				if a.netConn != nil {
					a.netConn.Close() // 关闭旧连接
				}
				a.netConn = conn
				a.mutex.Unlock()

				runtime.EventsEmit(a.ctx, "sys-msg", fmt.Sprintf("Client connected: %s", conn.RemoteAddr().String()))

				// 针对这个连接启动读取
				go a.handleTcpConnection(conn)
			}
		}
	}()

	return "Success"
}

func (a *App) handleTcpConnection(conn net.Conn) {
	buff := make([]byte, 4096)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			a.mutex.Lock()
			if a.netConn == conn {
				a.netConn = nil // 清理引用
			}
			a.mutex.Unlock()
			return
		}
		if n > 0 {
			// 数据拷贝，防止并发冲突
			dataToSend := make([]byte, n)
			copy(dataToSend, buff[:n])
			runtime.EventsEmit(a.ctx, "serial-data", dataToSend)
		}
	}
}

// OpenUdp 开启 UDP
func (a *App) OpenUdp(localPort string, remoteIp string, remotePort string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isConnected {
		return "Already connected"
	}

	lAddrStr := ":" + localPort
	conn, err := net.ListenPacket("udp", lAddrStr)
	if err != nil {
		return fmt.Sprintf("UDP Listen error: %v", err)
	}

	var rAddr net.Addr
	if remoteIp != "" && remotePort != "" {
		rAddr, err = net.ResolveUDPAddr("udp", net.JoinHostPort(remoteIp, remotePort))
		if err != nil {
			conn.Close()
			return fmt.Sprintf("Remote Addr error: %v", err)
		}
	}

	a.udpConn = conn
	a.udpRemote = rAddr
	a.connType = TypeUdp
	a.isConnected = true
	a.readStopChan = make(chan struct{})

	go func() {
		buff := make([]byte, 4096)
		for {
			select {
			case <-a.readStopChan:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
				n, addr, err := conn.ReadFrom(buff)
				if err != nil {
					// 忽略超时错误，继续循环
					if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
						continue
					}
					// 忽略关闭连接时的错误
					if a.isConnected {
						runtime.EventsEmit(a.ctx, "serial-error", err.Error())
					}
					return
				}

				a.mutex.Lock()
				if a.udpRemote == nil {
					a.udpRemote = addr
					runtime.EventsEmit(a.ctx, "sys-msg", fmt.Sprintf("Remote set to: %s", addr.String()))
				}
				a.mutex.Unlock()

				if n > 0 {
					dataToSend := make([]byte, n)
					copy(dataToSend, buff[:n])
					runtime.EventsEmit(a.ctx, "serial-data", dataToSend)
				}
			}
		}
	}()

	return "Success"
}

// --- 通用方法 ---

// 启动通用的 io.Reader 读取循环 (Serial / TCP Client)
func (a *App) startReadLoop(reader io.Reader) {
	a.isConnected = true
	a.readStopChan = make(chan struct{})

	go func() {
		buff := make([]byte, 4096)
		for {
			select {
			case <-a.readStopChan:
				return
			default:
				n, err := reader.Read(buff)
				if err != nil {
					if a.isConnected {
						// 只有在人为认为连接还开着的时候报错，才通知前端
						fmt.Printf("Read Error: %v\n", err)
						runtime.EventsEmit(a.ctx, "serial-error", err.Error())
						a.Close()
					}
					return
				}
				if n == 0 {
					continue
				}

				// [DEBUG] 在终端打印收到的数据长度，用于排查接收问题
				fmt.Printf("[DEBUG] Recv %d bytes\n", n)

				// 必须拷贝数据，防止 Wails 发送过程中 buff 被覆盖
				dataToSend := make([]byte, n)
				copy(dataToSend, buff[:n])
				runtime.EventsEmit(a.ctx, "serial-data", dataToSend)
			}
		}
	}()
}

// Close 关闭连接 (通用)
func (a *App) Close() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isConnected {
		return "Not connected"
	}

	a.isConnected = false
	if a.readStopChan != nil {
		close(a.readStopChan)
	}

	var err error

	switch a.connType {
	case TypeSerial:
		if a.serialPort != nil {
			err = a.serialPort.Close()
			a.serialPort = nil
		}
	case TypeTcpClient:
		if a.netConn != nil {
			err = a.netConn.Close()
			a.netConn = nil
		}
	case TypeTcpServer:
		if a.netListener != nil {
			err = a.netListener.Close()
			a.netListener = nil
		}
		if a.netConn != nil {
			a.netConn.Close()
			a.netConn = nil
		}
	case TypeUdp:
		if a.udpConn != nil {
			err = a.udpConn.Close()
			a.udpConn = nil
			a.udpRemote = nil
		}
	}

	if err != nil {
		return fmt.Sprintf("Error closing: %v", err)
	}
	return "Success"
}

// SendData 发送数据 (通用)
func (a *App) SendData(data string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isConnected {
		return "Error: Not connected"
	}

	// [修改] 移除了自动添加 \n 的逻辑，现在完全由前端控制

	payload := []byte(data)
	var err error

	switch a.connType {
	case TypeSerial:
		if a.serialPort != nil {
			_, err = a.serialPort.Write(payload)
		}
	case TypeTcpClient, TypeTcpServer:
		if a.netConn != nil {
			_, err = a.netConn.Write(payload)
		} else if a.connType == TypeTcpServer {
			return "Error: No client connected"
		}
	case TypeUdp:
		if a.udpConn != nil && a.udpRemote != nil {
			_, err = a.udpConn.WriteTo(payload, a.udpRemote)
		} else {
			return "Error: No remote address set"
		}
	}

	if err != nil {
		return fmt.Sprintf("Send error: %v", err)
	}
	return "Sent"
}
