package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.bug.st/serial"
)

// App struct
type App struct {
	ctx          context.Context
	port         serial.Port
	isConnected  bool
	mutex        sync.Mutex
	readStopChan chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
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

// OpenSerial 打开串口 (支持完整参数)
func (a *App) OpenSerial(portName string, baudRate int, dataBits int, stopBits int, parityName string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isConnected {
		return "Port already open"
	}

	// 1. 映射校验位
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

	// 2. 映射停止位 (前端传 1, 15(代表1.5), 2)
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

	// 3. 配置 Mode
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

	a.port = port
	a.isConnected = true
	a.readStopChan = make(chan struct{})

	go a.readLoop()

	return "Success"
}

// 3. 读取循环 (将数据推送给前端)
func (a *App) readLoop() {
	buff := make([]byte, 128) // 稍微加大一点缓冲
	for {
		select {
		case <-a.readStopChan:
			return
		default:
			n, err := a.port.Read(buff)
			if err != nil {
				// 关键修改：只有当连接状态显示为 true 时，才认为是异常错误
				// 如果 isConnected 已经是 false，说明是我们主动调用的 CloseSerial，直接退出即可
				if a.isConnected {
					runtime.EventsEmit(a.ctx, "serial-error", err.Error())
					a.CloseSerial() // 触发清理逻辑
				}
				return
			}
			if n == 0 {
				continue
			}
			runtime.EventsEmit(a.ctx, "serial-data", buff[:n])
		}
	}
}

// 4. 关闭串口
func (a *App) CloseSerial() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isConnected {
		return "Port not open"
	}

	// 关键修改：先修改状态，再关闭物理资源
	// 这样 readLoop 里的 err != nil 发生时，会看到 isConnected 已经是 false 了，就不会报错
	a.isConnected = false

	close(a.readStopChan)
	err := a.port.Close()
	a.port = nil

	if err != nil {
		return fmt.Sprintf("Error closing: %v", err)
	}
	return "Success"
}

// 5. 发送数据
func (a *App) SendData(data string) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isConnected {
		return "Error: Port not connected"
	}

	// 这里简化处理，直接发送字符串。如果是Hex发送，前端需先解析为字节数组传过来，
	// 或者在这里将 HexString 转为 []byte
	_, err := a.port.Write([]byte(data))
	if err != nil {
		return fmt.Sprintf("Send error: %v", err)
	}
	return "Sent"
}
