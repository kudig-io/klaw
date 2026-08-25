// AudioWorklet 运行时全局量（TS 标准 lib 未声明，本地补齐）
declare const sampleRate: number
declare class AudioWorkletProcessor {
  readonly port: MessagePort
}
declare function registerProcessor(name: string, processorCtor: unknown): void

// AudioWorklet：采集麦克风帧（上下文采样率），线性插值下采样到 16k，输出 Int16 PCM
class PcmProcessor extends AudioWorkletProcessor {
  private readonly ratio = sampleRate / 16000
  private pos = 0 // 下采样浮点游标（跨帧累计的小数部分）

  process(inputs: Float32Array[][]): boolean {
    const ch = inputs[0]?.[0]
    if (!ch) return true
    const outLen = Math.floor((ch.length - 1 - this.pos) / this.ratio) + 1
    if (outLen <= 0) {
      this.pos -= ch.length
      return true
    }
    const pcm = new Int16Array(outLen)
    let n = 0
    for (let i = this.pos; i < ch.length - 1 && n < outLen; i += this.ratio) {
      const idx = Math.floor(i)
      const frac = i - idx
      const v = ch[idx] * (1 - frac) + ch[idx + 1] * frac
      const s = Math.max(-1, Math.min(1, v))
      pcm[n++] = s < 0 ? s * 0x8000 : s * 0x7fff
    }
    this.pos = this.pos + n * this.ratio - ch.length
    if (n > 0) {
      const buf = pcm.buffer.slice(0, n * 2)
      this.port.postMessage(buf, [buf])
    }
    return true
  }
}

registerProcessor('pcm-processor', PcmProcessor)
