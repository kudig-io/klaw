// 相对时间生成：mock 数据的时间戳始终相对"现在"，保证演示数据新鲜
const iso = (ms: number) => new Date(Date.now() - ms).toISOString()

export const secondsAgo = (s: number) => iso(s * 1000)
export const minutesAgo = (m: number) => iso(m * 60 * 1000)
export const hoursAgo = (h: number) => iso(h * 3600 * 1000)
export const daysAgo = (d: number) => iso(d * 86400 * 1000)
