// Pod 日志与日志分析
// 每条日志行以 [mock] 前缀标记，单测断言包含该标记。
// analyzeLogs 是真正的解析器：从日志文本提取 level 计数、stack traces、慢请求、安全事件。

const TAG = '[mock]'

const line = (level: string, content: string) => `${TAG} ${level} ${content}`

const LOG_MAP: Record<string, string[]> = {
  nginx: [
    line('INFO', '127.0.0.1 - - [01/Apr/2026:12:00:00 +0000] "GET / HTTP/1.1" 200 612 "-" "kube-probe/1.29"'),
    line('INFO', '10.244.1.4 - - [01/Apr/2026:12:00:01 +0000] "GET /healthz HTTP/1.1" 200 12 "-" "kubelet"'),
    line('WARN', 'upstream response time exceeded threshold 1280ms path=/api/products'),
    line('INFO', '10.244.2.10 - - [01/Apr/2026:12:00:02 +0000] "GET /static/main.css HTTP/1.1" 304 0'),
    line('INFO', '10.244.1.7 - - [01/Apr/2026:12:00:03 +0000] "GET /api/listings HTTP/1.1" 200 8420'),
    line('ERROR', '502 Bad Gateway: upstream sent invalid header line'),
    line('INFO', '10.244.1.4 - - [01/Apr/2026:12:00:04 +0000] "GET /favicon.ico HTTP/1.1" 200 1410'),
    line('INFO', 'worker reloaded successfully in 128ms'),
  ],
  httpbin: [
    line('INFO', 'GET /status/200 200 1.2ms'),
    line('INFO', 'GET /uuid 200 0.8ms'),
    line('INFO', 'POST /anything 200 2.1ms'),
    line('INFO', 'GET /headers 200 0.6ms'),
    line('INFO', 'GET /delay/1 200 1004ms'),
    line('DEBUG', 'incoming request headers parsed (8 headers)'),
  ],
  // frontend（klaw-test）常规日志
  frontend: [
    line('INFO', '[http] GET / 200 24ms'),
    line('INFO', '[http] GET /products 200 38ms'),
    line('INFO', '[http] GET /api/cart 502 1820ms (upstream=mall-prod/cart-service)'),
    line('WARN', '[http] slow request /api/recommend 3214ms'),
    line('INFO', '[http] GET /static/app.js 304 0'),
  ],
  // mall-prod/frontend mall-frontend-7d9c5f8b4-z8x3c — 故事线：JS heap OOM 导致 CrashLoopBackOff
  'mall-frontend-oom': [
    line('INFO', 'application boot, env=production, commit=v2.4.1'),
    line('INFO', 'connected to mall-gateway upstream'),
    line('INFO', 'cache warm-up completed (124 entries)'),
    line('WARN', 'event loop lag 412ms'),
    line('ERROR', 'FATAL JavaScript heap out of memory'),
    '    at Object.<anonymous> (/srv/app/dist/server.bundle.js:14823:42)',
    '    at processImmediate (node:internal/timers:491:21)',
    '    at v8::Isolate::FatalProcessOutOfMemory (/srv/app/dist/server.bundle.js:14872:12)',
    line('FATAL', 'process exited with code 134'),
    line('INFO', 'kubelet restarting container mall-frontend'),
  ],
  'mall-gateway': [
    line('INFO', 'upstream order-service responded 200 18ms'),
    line('INFO', 'upstream payment-service responded 200 42ms'),
    line('WARN', 'upstream cart-service responded 502, retry 1/3'),
    line('INFO', 'rate limit window reset'),
    line('INFO', 'circuit-breaker closed for payment-service'),
  ],
  'order-service': [
    line('INFO', 'mysql query took 18ms queryId=42'),
    line('WARN', 'mysql query slow 1240ms SELECT * FROM order_items WHERE order_id=?'),
    line('INFO', 'rabbitmq ack order_id=1024'),
    line('INFO', 'mysql query took 22ms'),
  ],
  'payment-service': [
    line('INFO', 'stripe charge succeeded amount=120.00 order_id=1024'),
    line('WARN', '401 Unauthorized from 198.51.100.66 path=/api/v1/charge token=invalid'),
    line('INFO', 'fraud-check passed order_id=1025'),
    line('ERROR', 'HTTP 503 from upstream acquirer, retry scheduled in 30s'),
  ],
  'cart-service': [
    line('INFO', 'cart updated user_id=8c1f items=3'),
    line('INFO', 'cache miss for sku=SKU-9182, fetched from product-service'),
    line('INFO', 'cart synced to redis cluster'),
  ],
  'inventory-service': [
    line('INFO', 'stock adjusted sku=SKU-9182 delta=-1 new=8'),
    line('INFO', 'low-stock alert sku=SKU-2204 stock=2'),
  ],
  redis: [
    line('INFO', 'redis cluster ready, master=10.244.2.45:6379'),
    line('INFO', 'BGSAVE scheduled, last save 240s ago'),
    line('INFO', 'client connected from 10.244.1.33'),
  ],
  mysql: [
    line('INFO', 'mysqld: ready for connections, version 8.0.39'),
    line('INFO', 'InnoDB: Buffer pool: 4096 MB allocated'),
    line('WARN', 'Slow query recorded (1230ms) SELECT * FROM order_items WHERE created_at > ?'),
  ],
  // 预发/数据平台
  'mall-frontend-staging': [
    line('INFO', 'staging build v2.5.0-rc.1'),
    line('INFO', 'GET /api/canary 200 28ms'),
    line('INFO', 'canary traffic weight=10%'),
  ],
  'order-service-staging': [
    line('INFO', 'staging bootstrap, mysql test schema ready'),
    line('WARN', 'kafka producer retry attempt=2'),
  ],
  spark: [
    line('INFO', 'spark.driver: starting driver at spark://spark-driver.data-platform:7077'),
    line('INFO', 'spark.executor: registered executor 0'),
    line('WARN', 'spark.scheduler: stage 3 took 12400ms (expected <5000ms)'),
  ],
  kafka: [
    line('INFO', 'kafka server started, broker.id=1'),
    line('INFO', 'topic created: orders.events partitions=12 replication=2'),
  ],
  flink: [
    line('INFO', 'flink jobmanager started, rest address = http://flink-jobmanager:8081'),
    line('INFO', 'submitted job: orders-streaming-aggregator (streaming)'),
  ],
  coredns: [
    line('INFO', 'plugin/forward: forwarding query'),
    line('INFO', 'plugin/cache: HIT for klaw.local/A'),
  ],
  ingress: [
    line('INFO', 'controller started, watching ingressclasses'),
    line('INFO', 'updated Ingress mall-frontend/ingress-v2'),
  ],
  'mall-frontend-staging-error': [
    line('INFO', 'OK endpoint /healthz 200'),
    line('ERROR', 'Error fetching upstream: ECONNREFUSED 10.96.20.3:8081'),
    line('ERROR', '    at TCPConnectWrap.afterConnect [as oncomplete] (net.js:1138:16)'),
  ],
}

const DEFAULT = [
  line('INFO', 'application started, listening on :8080'),
  line('INFO', 'metrics endpoint enabled, scraping every 15s'),
  line('DEBUG', 'serving static files from /srv/public'),
]

// 按 Pod 名字匹配最合适的日志模板
export function getLogsForPod(podName: string): string {
  const n = podName.toLowerCase()
  if (n.includes('mall-frontend')) {
    if (n.endsWith('z8x3c')) return LOG_MAP['mall-frontend-oom'].join('\n')
    return LOG_MAP['mall-frontend-staging-error'].join('\n')
  }
  if (n.includes('frontend')) return LOG_MAP.frontend.join('\n')
  if (n.includes('nginx')) return LOG_MAP.nginx.join('\n')
  if (n.includes('httpbin')) return LOG_MAP.httpbin.join('\n')
  if (n.includes('mall-gateway') || n.includes('gateway')) return LOG_MAP['mall-gateway'].join('\n')
  if (n.includes('order-service-staging')) return LOG_MAP['order-service-staging'].join('\n')
  if (n.includes('order-service')) return LOG_MAP['order-service'].join('\n')
  if (n.includes('cart-service')) return LOG_MAP['cart-service'].join('\n')
  if (n.includes('payment-service')) return LOG_MAP['payment-service'].join('\n')
  if (n.includes('inventory-service')) return LOG_MAP['inventory-service'].join('\n')
  if (n.includes('redis')) return LOG_MAP.redis.join('\n')
  if (n.includes('mysql')) return LOG_MAP.mysql.join('\n')
  if (n.includes('spark')) return LOG_MAP.spark.join('\n')
  if (n.includes('kafka')) return LOG_MAP.kafka.join('\n')
  if (n.includes('flink')) return LOG_MAP.flink.join('\n')
  if (n.includes('coredns')) return LOG_MAP.coredns.join('\n')
  if (n.includes('ingress')) return LOG_MAP.ingress.join('\n')
  return DEFAULT.join('\n')
}

// 从任意日志文本派生 LogAnalysis（真实解析）
export function analyzeLogs(logs: string) {
  const lines = logs.split('\n')
  const errors: { line: number; content: string; timestamp?: string; level: string }[] = []
  const warnings: { line: number; content: string; timestamp?: string; level: string }[] = []
  const stackTraces: string[] = []
  const slowRequests: { url: string; responseTime: string }[] = []
  const securityEvents: { type: string; message: string; severity: string }[] = []
  const logLevels: Record<string, number> = {}
  const patternStats: Record<string, number> = {}

  let errorCount = 0, warningCount = 0, infoCount = 0, debugCount = 0

  lines.forEach((raw, i) => {
    // 栈帧：以 "    at" 起头（任意位置的 4 空格 + at），不属于日志行
    if (/^\s+at /.test(raw)) {
      stackTraces.push(raw.trim())
      return
    }
    const lvlMatch = raw.match(/^(?:\[mock\]\s+)?(INFO|WARN|ERROR|FATAL|DEBUG)\s/)
    if (!lvlMatch) return
    const lvl = lvlMatch[1]
    logLevels[lvl] = (logLevels[lvl] || 0) + 1
    if (lvl === 'ERROR' || lvl === 'FATAL') {
      errorCount++
      errors.push({ line: i + 1, content: raw, level: lvl })
    } else if (lvl === 'WARN') {
      warningCount++
      warnings.push({ line: i + 1, content: raw, level: lvl })
    } else if (lvl === 'INFO') {
      infoCount++
    } else if (lvl === 'DEBUG') {
      debugCount++
    }

    // pattern: HTTP 状态码
    const httpMatch = raw.match(/\b([1-5]\d\d)\s/)
    if (httpMatch) {
      const code = httpMatch[1]
      patternStats[`http_${code}`] = (patternStats[`http_${code}`] || 0) + 1
    }

    // slow requests: >500ms duration
    const slowMatch = raw.match(/(\d{4,})ms/)
    if (slowMatch && parseInt(slowMatch[1],) > 500) {
      const urlMatch = raw.match(/(?:GET|POST|PUT|DELETE)\s(\S+)/)
      slowRequests.push({
        url: urlMatch ? urlMatch[1] : '/unknown',
        responseTime: `${slowMatch[1]}ms`,
      })
    }

    // security: 401/403/未授权
    if (/\b(401|403)\b/.test(raw) || /Unauthorized/.test(raw)) {
      const ipMatch = raw.match(/from\s+(\d+\.\d+\.\d+\.\d+)/)
      securityEvents.push({
        type: 'authentication_failed',
        message: raw.replace(/^\[mock\]\s+/, ''),
        severity: 'warning',
      })
      if (ipMatch) {
        patternStats[`security_auth_failed_${ipMatch[1]}`] = (patternStats[`security_auth_failed_${ipMatch[1]}`] || 0) + 1
      }
    }
  })

  patternStats['http_total'] = lines.filter((l) => /\b[1-5]\d\d\b/.test(l)).length

  return {
    totalLines: lines.length,
    errorCount,
    warningCount,
    infoCount,
    debugCount,
    errors,
    warnings,
    stackTraces,
    logLevels,
    patternStats,
    performanceMetrics: { slowRequests },
    securityEvents,
  }
}