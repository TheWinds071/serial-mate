<script setup lang="ts">
import { ref, onMounted, nextTick, watch, computed, reactive } from 'vue';
// 引入后端方法
import { GetSerialPorts, OpenSerial, OpenTcpClient, OpenTcpServer, OpenUdp, Close as CloseConnection, SendData } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// --- 1. 核心状态 ---
const portList = ref<string[]>([]);
const selectedPort = ref('');
const isConnected = ref(false);

// 模式选择
const mode = ref<'SERIAL' | 'TCP_CLIENT' | 'TCP_SERVER' | 'UDP'>('SERIAL');

// Serial 参数
const baudRate = ref(115200);
const dataBits = ref(8);
const stopBits = ref(1);
const parity = ref('None');
const baudOptions = [9600, 19200, 38400, 57600, 115200, 921600];

// Network 参数
const netIp = ref('127.0.0.1');
const netPort = ref('43211');
const udpLocalPort = ref('8081');

// --- 2. 数据处理 ---
const receivedData = ref<string>('');
const rawDataBuffer = ref<number[]>([]);
const sendInput = ref('');
const showHex = ref(true);
// 新增：控制是否追加换行符
const appendNewline = ref(false);
const autoScroll = ref(true);
const logWindowRef = ref<HTMLElement | null>(null);
const rxCount = ref(0);
const txCount = ref(0);

// --- 3. 主题 (简化) ---
const showThemePanel = ref(false);
const defaultTheme = {
  bgMain: '#F2F1ED', bgSide: '#EBEAE6', primary: '#7A8B99', textMain: '#5C5C5C', textSub: '#888888',
};
const theme = reactive({ ...defaultTheme });
const cssVars = computed(() => ({
  '--bg-main': theme.bgMain, '--bg-side': theme.bgSide, '--col-primary': theme.primary, '--text-main': theme.textMain, '--text-sub': theme.textSub,
}));
const resetTheme = () => Object.assign(theme, defaultTheme);

// --- 4. 生命周期 ---
onMounted(async () => {
  await refreshPorts();

  // 数据接收监听
  EventsOn("serial-data", (data: any) => {
    let bytes: number[] = [];

    // Wails 的 []byte 传递过来通常是 Base64 字符串
    if (typeof data === 'string') {
      try {
        bytes = base64ToBytes(data);
      } catch (e) {
        console.error("Base64 decode error:", e);
        return;
      }
    } else if (Array.isArray(data)) {
      bytes = data;
    }

    if (bytes && bytes.length > 0) {
      // [调试] 如果能在浏览器控制台看到这行，说明数据肯定到了前端
      console.log(`RX: ${bytes.length} bytes`, bytes);

      rawDataBuffer.value.push(...bytes);
      rxCount.value += bytes.length;
      receivedData.value += formatData(bytes, showHex.value);
      scrollToBottom();
    }
  });

  EventsOn("serial-error", (err) => {
    console.error("Connection error:", err);
    isConnected.value = false;
    alert("连接已断开: " + err);
  });

  EventsOn("sys-msg", (msg) => {
    console.log("Sys Msg:", msg);
  });
});

const base64ToBytes = (base64: string): number[] => {
  const binaryString = window.atob(base64);
  const len = binaryString.length;
  const bytes = new Array(len);
  for (let i = 0; i < len; i++) bytes[i] = binaryString.charCodeAt(i);
  return bytes;
};

const refreshPorts = async () => {
  try {
    portList.value = await GetSerialPorts();
    if (portList.value.length > 0 && !selectedPort.value) selectedPort.value = portList.value[0];
  } catch (e) { console.error(e); }
};

const toggleConnection = async () => {
  if (isConnected.value) {
    await CloseConnection();
    isConnected.value = false;
  } else {
    let res = "";
    if (mode.value === 'SERIAL') {
      if (!selectedPort.value) return;
      res = await OpenSerial(selectedPort.value, Number(baudRate.value), Number(dataBits.value), Number(stopBits.value), parity.value);
    } else if (mode.value === 'TCP_CLIENT') {
      if (!netIp.value || !netPort.value) return;
      res = await OpenTcpClient(netIp.value, netPort.value);
    } else if (mode.value === 'TCP_SERVER') {
      if (!netPort.value) return;
      res = await OpenTcpServer(netPort.value);
    } else if (mode.value === 'UDP') {
      if (!udpLocalPort.value) return;
      res = await OpenUdp(udpLocalPort.value, netIp.value, netPort.value);
    }

    if (res === "Success") {
      isConnected.value = true;
    } else {
      alert("连接失败: " + res);
    }
  }
};

const handleSend = async () => {
  if (!sendInput.value) return;

  // [修改] 根据复选框决定是否添加换行符
  let dataToSend = sendInput.value;
  if (appendNewline.value) {
    dataToSend += "\n";
  }

  const res = await SendData(dataToSend);

  if(res === 'Sent') {
    txCount.value += dataToSend.length;
    // 发送成功后保留输入框内容，方便重复发送
  } else {
    alert("发送失败: " + res);
  }
};

const clearReceive = () => {
  receivedData.value = '';
  rawDataBuffer.value = [];
  rxCount.value = 0;
};

const decoder = new TextDecoder('utf-8');
const formatData = (bytes: number[], isHex: boolean): string => {
  if (isHex) {
    return bytes.map(b => b.toString(16).padStart(2, '0').toUpperCase()).join(' ') + ' ';
  } else {
    // 使用 stream: true 可以处理被截断的 UTF-8 字符
    return decoder.decode(new Uint8Array(bytes), { stream: true });
  }
};

watch(showHex, () => {
  receivedData.value = formatData(rawDataBuffer.value, showHex.value);
});

const scrollToBottom = () => {
  if (!autoScroll.value || !logWindowRef.value) return;
  nextTick(() => {
    if(logWindowRef.value) logWindowRef.value.scrollTop = logWindowRef.value.scrollHeight;
  });
};
</script>

<template>
  <div :style="cssVars" class="flex h-screen w-screen bg-[var(--bg-main)] text-[var(--text-main)] font-sans overflow-hidden select-none transition-colors duration-300">

    <!-- 侧边栏 -->
    <div class="w-72 bg-[var(--bg-side)] flex flex-col shrink-0 border-r border-black/5 transition-colors duration-300 relative">
      <div class="h-14 flex items-center justify-between px-4 border-b border-black/5">
        <span class="font-bold text-lg tracking-widest text-[var(--col-primary)]">SERIAL MATE</span>
        <button @click="showThemePanel = !showThemePanel" class="p-1.5 rounded-md hover:bg-black/5 text-[var(--text-sub)] transition-colors">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="13.5" cy="6.5" r=".5"></circle><circle cx="17.5" cy="10.5" r=".5"></circle><circle cx="8.5" cy="7.5" r=".5"></circle><circle cx="6.5" cy="12.5" r=".5"></circle><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z"></path></svg>
        </button>
      </div>

      <!-- 主题面板 -->
      <div v-if="showThemePanel" class="absolute top-14 left-0 w-full bg-white/90 backdrop-blur-md p-4 shadow-lg border-b border-black/5 z-20 space-y-3">
        <!-- 省略具体的颜色选择器代码以保持简洁，逻辑与之前相同 -->
        <div class="flex justify-between items-center text-xs font-bold text-[var(--text-sub)]">
          <span>自定义配色</span> <button @click="resetTheme">重置</button>
        </div>
        <!-- ... 颜色选择器 ... -->
      </div>

      <div class="flex-1 overflow-y-auto p-5 space-y-5 custom-scrollbar">
        <!-- 模式切换 -->
        <div class="bg-white/40 p-1 rounded-lg shadow-sm border border-black/5 flex text-[10px] font-bold">
          <button @click="mode='SERIAL'" :class="{'bg-white text-[var(--col-primary)] shadow-sm': mode==='SERIAL', 'text-[var(--text-sub)]': mode!=='SERIAL'}" class="flex-1 py-1.5 rounded transition-all" :disabled="isConnected">SERIAL</button>
          <button @click="mode='TCP_CLIENT'" :class="{'bg-white text-[var(--col-primary)] shadow-sm': mode==='TCP_CLIENT', 'text-[var(--text-sub)]': mode!=='TCP_CLIENT'}" class="flex-1 py-1.5 rounded transition-all" :disabled="isConnected">TCP-C</button>
          <button @click="mode='TCP_SERVER'" :class="{'bg-white text-[var(--col-primary)] shadow-sm': mode==='TCP_SERVER', 'text-[var(--text-sub)]': mode!=='TCP_SERVER'}" class="flex-1 py-1.5 rounded transition-all" :disabled="isConnected">TCP-S</button>
          <button @click="mode='UDP'" :class="{'bg-white text-[var(--col-primary)] shadow-sm': mode==='UDP', 'text-[var(--text-sub)]': mode!=='UDP'}" class="flex-1 py-1.5 rounded transition-all" :disabled="isConnected">UDP</button>
        </div>

        <div class="bg-white/40 p-3 rounded-lg shadow-sm border border-black/5 space-y-3">
          <div class="text-xs font-bold text-[var(--text-sub)] opacity-70 uppercase tracking-wider mb-1">
            {{ mode.replace('_', ' ') }} Settings
          </div>

          <template v-if="mode === 'SERIAL'">
            <!-- Serial Inputs ... -->
            <div class="control-group"><label>端口</label><div class="relative flex-1"><select v-model="selectedPort" @click="refreshPorts" class="morandi-input" :disabled="isConnected"><option v-for="p in portList" :key="p" :value="p">{{ p }}</option></select></div></div>
            <div class="control-group"><label>波特率</label><div class="relative flex-1"><input type="number" v-model="baudRate" list="baud-list" class="morandi-input" placeholder="Custom" :disabled="isConnected"><datalist id="baud-list"><option v-for="b in baudOptions" :key="b" :value="b"></option></datalist></div></div>
          </template>

          <template v-if="mode === 'TCP_CLIENT'">
            <div class="control-group"><label>IP</label><input type="text" v-model="netIp" class="morandi-input" placeholder="127.0.0.1" :disabled="isConnected"></div>
            <div class="control-group"><label>Port</label><input type="text" v-model="netPort" class="morandi-input" placeholder="43211" :disabled="isConnected"></div>
          </template>

          <template v-if="mode === 'TCP_SERVER'">
            <div class="control-group"><label>Local Port</label><input type="text" v-model="netPort" class="morandi-input" placeholder="8080" :disabled="isConnected"></div>
          </template>

          <template v-if="mode === 'UDP'">
            <div class="control-group"><label>Local Port</label><input type="text" v-model="udpLocalPort" class="morandi-input" placeholder="8081" :disabled="isConnected"></div>
            <div class="my-2 border-t border-black/5"></div>
            <div class="control-group"><label>Target IP</label><input type="text" v-model="netIp" class="morandi-input" placeholder="127.0.0.1" :disabled="isConnected"></div>
            <div class="control-group"><label>Target Port</label><input type="text" v-model="netPort" class="morandi-input" placeholder="8080" :disabled="isConnected"></div>
          </template>
        </div>

        <button @click="toggleConnection" class="w-full py-2.5 rounded-lg font-medium text-white transition-all duration-300 transform active:scale-[0.98] shadow-sm flex items-center justify-center space-x-2 bg-[var(--col-primary)] hover:opacity-90">
          <div v-if="!isConnected" class="w-2 h-2 rounded-full bg-white animate-pulse"></div>
          <span>{{ isConnected ? '断开' : '连接' }}</span>
        </button>

        <div class="space-y-2 pt-2">
          <div class="text-xs font-bold text-[var(--text-sub)] opacity-70 uppercase tracking-wider">Display</div>
          <label class="flex items-center space-x-2 cursor-pointer hover:text-[var(--col-primary)] transition-colors">
            <input type="checkbox" v-model="showHex" class="accent-[var(--col-primary)] w-4 h-4">
            <span class="text-sm">Hex 显示</span>
          </label>
          <label class="flex items-center space-x-2 cursor-pointer hover:text-[var(--col-primary)] transition-colors">
            <input type="checkbox" v-model="autoScroll" class="accent-[var(--col-primary)] w-4 h-4">
            <span class="text-sm">自动滚屏</span>
          </label>
        </div>
      </div>
    </div>

    <!-- 右侧主区域 -->
    <div class="flex-1 flex flex-col min-w-0 p-4 gap-4 transition-colors duration-300">
      <div class="flex-1 bg-white/60 rounded-xl shadow-[0_2px_12px_-4px_rgba(0,0,0,0.08)] border border-black/5 flex flex-col overflow-hidden relative backdrop-blur-sm">
        <div class="h-10 px-4 flex items-center justify-between bg-black/[0.02] border-b border-black/5">
          <div class="flex items-center space-x-2">
            <span class="text-xs font-bold text-[var(--col-primary)] tracking-wider">RX MONITOR</span>
            <span class="text-[10px] text-[var(--text-sub)] bg-black/5 px-1.5 py-0.5 rounded-md">{{ rxCount }} Bytes</span>
          </div>
          <button @click="clearReceive" title="清空" class="group flex items-center justify-center w-7 h-7 rounded hover:bg-white hover:shadow-sm text-[var(--text-sub)] hover:text-[var(--col-primary)] transition-all">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6L7.5 16.5"></path><path d="M19.5 4.5L16.5 7.5"></path><path d="M2 22L4.5 19.5"></path><path d="M9.5 12.5C7.5 14.5 6 15 5 16C4 17 3 17 3 17C3 17 3 18 4 19C5 20 5 20 5 20C5 20 6 20 7 19C8 18 8.5 16.5 10.5 14.5L18 7"></path></svg>
          </button>
        </div>
        <textarea ref="logWindowRef" readonly class="flex-1 w-full p-4 font-mono text-sm bg-transparent resize-none outline-none custom-scrollbar leading-relaxed text-[var(--text-main)]" :value="receivedData"></textarea>
      </div>

      <div class="h-40 bg-white/60 rounded-xl shadow-[0_2px_12px_-4px_rgba(0,0,0,0.08)] border border-black/5 flex flex-col overflow-hidden backdrop-blur-sm">
        <div class="h-9 px-4 flex items-center justify-between bg-black/[0.02] border-b border-black/5">
          <div class="flex items-center space-x-4">
            <span class="text-xs font-bold text-[var(--text-sub)] tracking-wider">TX EDITOR</span>

            <!-- 新增：Add Newline 复选框 -->
            <label class="flex items-center space-x-1 cursor-pointer hover:text-[var(--col-primary)]" title="发送时自动追加换行符">
              <input type="checkbox" v-model="appendNewline" class="accent-[var(--col-primary)] w-3 h-3">
              <span class="text-[11px] text-[var(--text-sub)]">Add Newline (\n)</span>
            </label>
          </div>
        </div>

        <div class="flex-1 flex p-3 gap-3">
          <textarea v-model="sendInput" class="flex-1 bg-white/50 border border-transparent focus:border-[var(--col-primary)]/30 rounded-lg p-3 font-mono text-sm text-[var(--text-main)] focus:bg-white transition-all outline-none resize-none placeholder-[var(--text-sub)]/50" placeholder="Input data to send..." @keydown.enter.ctrl.prevent="handleSend"></textarea>
          <div class="flex flex-col gap-2 w-20">
            <button @click="handleSend" class="flex-1 bg-[var(--col-primary)] hover:opacity-90 text-white rounded-lg shadow-sm transition-all flex flex-col items-center justify-center active:scale-95"><span class="text-xs font-bold tracking-widest">SEND</span></button>
            <button @click="sendInput=''" class="h-8 bg-black/5 text-[var(--text-sub)] hover:bg-black/10 rounded-lg text-xs">CLR</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.control-group { display: flex; align-items: center; gap: 0.5rem; }
.control-group label { width: 60px; text-align: right; font-size: 0.75rem; color: var(--text-sub); }
.morandi-input { width: 100%; background-color: rgba(255, 255, 255, 0.6); border: 1px solid rgba(0, 0, 0, 0.1); color: var(--text-main); padding: 0.25rem 0.5rem; font-size: 0.8rem; border-radius: 0.375rem; outline: none; transition: all 0.2s; }
.morandi-input:focus { background-color: #fff; border-color: var(--col-primary); }
.morandi-input:disabled { opacity: 0.6; cursor: not-allowed; }
.custom-scrollbar::-webkit-scrollbar { width: 6px; height: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.15); border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: var(--col-primary); }
</style>