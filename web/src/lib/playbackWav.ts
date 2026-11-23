// Slice PCM samples before playback rather than relying on timeupdate (which
// can fire hundreds of milliseconds late and play the next sentence).
// Read only RIFF headers and keep the selected PCM bytes as a Blob slice.
async function readBytes(blob: Blob, start: number, end: number): Promise<Uint8Array> {
  return new Uint8Array(await blob.slice(start, end).arrayBuffer())
}

function readTag(bytes: Uint8Array, offset: number): string {
  return String.fromCharCode(...bytes.subarray(offset, offset + 4))
}

export async function slicePlaybackWav(blob: Blob, startMs: number, endMs?: number): Promise<Blob> {
  if (!Number.isFinite(startMs) || startMs < 0 || (endMs !== undefined && (!Number.isFinite(endMs) || endMs <= startMs))) {
    throw new Error('音频片段时间范围无效')
  }
  const riff = await readBytes(blob, 0, 12)
  if (riff.byteLength < 12 || readTag(riff, 0) !== 'RIFF' || readTag(riff, 8) !== 'WAVE') throw new Error('回放音频不是 WAV')
  let format: Uint8Array | undefined
  let offset = 0
  let size = 0
  for (let pos = 12; pos + 8 <= blob.size;) {
    const chunk = await readBytes(blob, pos, pos + 8)
    const length = new DataView(chunk.buffer, chunk.byteOffset, chunk.byteLength).getUint32(4, true)
    const dataStart = pos + 8
    const chunkEnd = dataStart + length
    if (chunkEnd > blob.size) throw new Error('回放音频不完整')
    const tag = readTag(chunk, 0)
    if (tag === 'fmt ' && length >= 16) format = await readBytes(blob, dataStart, dataStart + 16)
    if (tag === 'data') {
      offset = dataStart
      size = length
    }
    pos = chunkEnd + (length % 2)
  }
  if (!format || !offset) throw new Error('回放音频缺少 PCM 数据')
  const fmt = new DataView(format.buffer, format.byteOffset, format.byteLength)
  const rate = fmt.getUint32(4, true)
  const align = fmt.getUint16(12, true)
  if (fmt.getUint16(0, true) !== 1 || fmt.getUint16(14, true) !== 16 || !rate || !align) throw new Error('不支持的回放音频格式')
  const frames = Math.floor(size / align)
  const start = Math.min(frames, Math.floor(startMs * rate / 1000))
  const end = endMs === undefined ? frames : Math.min(frames, Math.ceil(endMs * rate / 1000))
  if (end <= start) throw new Error('该片段没有可播放的音频')
  const length = (end - start) * align
  const header = new ArrayBuffer(44)
  const output = new DataView(header)
  const write = (pos: number, value: string) => [...value].forEach((char, i) => output.setUint8(pos + i, char.charCodeAt(0)))
  write(0, 'RIFF'); output.setUint32(4, 36 + length, true); write(8, 'WAVE'); write(12, 'fmt ')
  output.setUint32(16, 16, true); new Uint8Array(header, 20, 16).set(format)
  write(36, 'data'); output.setUint32(40, length, true)
  const pcm = blob.slice(offset + start * align, offset + end * align)
  return new Blob([header, pcm], { type: 'audio/wav' })
}
