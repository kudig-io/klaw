// Ingress 与 NetworkPolicy mock 数据（多命名空间、覆盖 TLS/多 host/通配符/defaultBackend/ipBlock/endPort 等形态）

import { daysAgo } from '../time'

// ── Ingress ───────────────────────────────────────────────

interface IngressPathSeed {
  path: string
  pathType?: 'Exact' | 'Prefix' | 'ImplementationSpecific'
  service: string
  port?: number
  portName?: string
}

interface IngressRuleSeed {
  host?: string
  paths?: IngressPathSeed[]
}

interface IngressSeed {
  name: string
  namespace: string
  createdAt: string
  className?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  tls?: Array<{ hosts?: string[]; secretName?: string }>
  rules?: IngressRuleSeed[]
  defaultBackendService?: { name: string; port?: number }
  lbIP?: string
}

const toBackendPort = (p: IngressPathSeed) =>
  p.portName ? { name: p.portName } : { number: p.port ?? 80 }

const ing = (s: IngressSeed) => ({
  metadata: {
    name: s.name,
    namespace: s.namespace,
    creationTimestamp: s.createdAt,
    labels: s.labels,
    annotations: s.annotations,
  },
  spec: {
    ingressClassName: s.className,
    defaultBackend: s.defaultBackendService
      ? {
          service: {
            name: s.defaultBackendService.name,
            port: s.defaultBackendService.port ? { number: s.defaultBackendService.port } : undefined,
          },
        }
      : undefined,
    tls: s.tls,
    rules: (s.rules ?? []).map((r) => ({
      host: r.host,
      http: r.paths
        ? {
            paths: r.paths.map((p) => ({
              path: p.path,
              pathType: p.pathType ?? 'Prefix',
              backend: {
                service: { name: p.service, port: toBackendPort(p) },
              },
            })),
          }
        : undefined,
    })),
  },
  status: {
    loadBalancer: { ingress: s.lbIP ? [{ ip: s.lbIP }] : [] },
  },
})

export const mockIngresses = [
  // klaw-test：无 TLS、多路径后端
  ing({
    name: 'frontend-ingress', namespace: 'klaw-test', className: 'nginx', createdAt: daysAgo(3),
    labels: { app: 'frontend' },
    annotations: {
      'nginx.ingress.kubernetes.io/rewrite-target': '/',
      'nginx.ingress.kubernetes.io/proxy-read-timeout': '60',
    },
    rules: [{
      host: 'frontend.klaw-test.example.com',
      paths: [
        { path: '/', service: 'frontend', portName: 'http' },
        { path: '/api', service: 'httpbin', port: 8000 },
      ],
    }],
    lbIP: '172.18.0.101',
  }),
  // klaw-test：路径限速注解
  ing({
    name: 'httpbin-ingress', namespace: 'klaw-test', className: 'nginx', createdAt: daysAgo(3),
    annotations: { 'nginx.ingress.kubernetes.io/limit-rps': '10' },
    rules: [{
      host: 'httpbin.klaw-test.example.com',
      paths: [{ path: '/', pathType: 'ImplementationSpecific', service: 'httpbin', port: 8000 }],
    }],
    lbIP: '172.18.0.101',
  }),
  // mall-prod：TLS 双证书、多 host、多路径
  ing({
    name: 'mall-ingress', namespace: 'mall-prod', className: 'nginx', createdAt: daysAgo(5),
    labels: { app: 'mall', env: 'prod' },
    annotations: {
      'nginx.ingress.kubernetes.io/ssl-redirect': 'true',
      'nginx.ingress.kubernetes.io/proxy-body-size': '8m',
      'cert-manager.io/cluster-issuer': 'letsencrypt-prod',
    },
    tls: [{ hosts: ['mall.example.com', 'api.mall.example.com'], secretName: 'mall-tls-cert' }],
    rules: [
      {
        host: 'mall.example.com',
        paths: [
          { path: '/', service: 'mall-frontend', port: 80 },
          { path: '/cart', service: 'mall-frontend', port: 80 },
        ],
      },
      {
        host: 'api.mall.example.com',
        paths: [
          { path: '/orders', service: 'order-service', port: 8081 },
          { path: '/payments', service: 'payment-service', port: 8083 },
          { path: '/gateway', service: 'mall-gateway', port: 8080 },
        ],
      },
    ],
    lbIP: '172.18.0.101',
  }),
  // mall-staging：通配符 host + defaultBackend
  ing({
    name: 'mall-staging-ingress', namespace: 'mall-staging', className: 'nginx', createdAt: daysAgo(2),
    annotations: { 'nginx.ingress.kubernetes.io/canary': 'false' },
    tls: [{ hosts: ['*.staging.mall.example.com'], secretName: 'staging-wildcard-cert' }],
    rules: [{
      host: '*.staging.mall.example.com',
      paths: [{ path: '/', service: 'mall-frontend-staging', port: 80 }],
    }],
    defaultBackendService: { name: 'mall-frontend-staging', port: 80 },
    lbIP: '172.18.0.101',
  }),
  // ingress-nginx：控制器自身入口
  ing({
    name: 'ingress-nginx-dashboard', namespace: 'ingress-nginx', className: 'nginx', createdAt: daysAgo(180),
    labels: { 'app.kubernetes.io/name': 'ingress-nginx' },
    rules: [{
      host: 'nginx-dashboard.example.com',
      paths: [{ path: '/dashboard', service: 'ingress-nginx-controller', port: 80 }],
    }],
    lbIP: '172.18.0.101',
  }),
  // default：未分配 LB 地址的遗留 Ingress
  ing({
    name: 'legacy-web-ingress', namespace: 'default', className: 'nginx', createdAt: daysAgo(150),
    annotations: { 'kubernetes.io/ingress.class': 'nginx' },
    rules: [{
      host: 'legacy.example.com',
      paths: [{ path: '/', pathType: 'Exact', service: 'kubernetes', port: 443 }],
    }],
  }),
]

// ── NetworkPolicy ─────────────────────────────────────────

interface NPSelectorSeed {
  matchLabels?: Record<string, string>
  matchExpressions?: Array<{ key: string; operator: string; values: string[] }>
}

interface NPPeerSeed {
  podSelector?: NPSelectorSeed
  namespaceSelector?: NPSelectorSeed
  ipBlock?: { cidr: string; except?: string[] }
}

interface NPPortSeed {
  protocol: string
  port: number
  endPort?: number
}

interface NPPolicySeed {
  name: string
  namespace: string
  createdAt: string
  labels?: Record<string, string>
  podSelector: NPSelectorSeed
  policyTypes: Array<'Ingress' | 'Egress'>
  ingress?: Array<{ from?: NPPeerSeed[]; ports?: NPPortSeed[] }>
  egress?: Array<{ to?: NPPeerSeed[]; ports?: NPPortSeed[] }>
}

const np = (s: NPPolicySeed) => ({
  metadata: {
    name: s.name,
    namespace: s.namespace,
    creationTimestamp: s.createdAt,
    labels: s.labels,
  },
  spec: {
    podSelector: s.podSelector,
    policyTypes: s.policyTypes,
    ingress: s.ingress,
    egress: s.egress,
  },
})

export const mockNetworkPolicies = [
  // mall-prod：全命名空间默认拒绝（双向）
  np({
    name: 'default-deny-all', namespace: 'mall-prod', createdAt: daysAgo(30),
    labels: { policy: 'baseline' },
    podSelector: {},
    policyTypes: ['Ingress', 'Egress'],
    ingress: [],
    egress: [],
  }),
  // mall-prod：frontend → gateway 白名单
  np({
    name: 'allow-frontend-to-gateway', namespace: 'mall-prod', createdAt: daysAgo(30),
    podSelector: { matchLabels: { app: 'mall-gateway' } },
    policyTypes: ['Ingress'],
    ingress: [{
      from: [{ podSelector: { matchLabels: { app: 'mall-frontend' } } }],
      ports: [{ protocol: 'TCP', port: 8080 }],
    }],
  }),
  // mall-prod：DNS 出站放行（namespace + pod 双 selector）
  np({
    name: 'allow-dns-egress', namespace: 'mall-prod', createdAt: daysAgo(30),
    podSelector: {},
    policyTypes: ['Egress'],
    egress: [{
      to: [{
        namespaceSelector: { matchLabels: { 'kubernetes.io/metadata.name': 'kube-system' } },
        podSelector: { matchLabels: { 'k8s-app': 'kube-dns' } },
      }],
      ports: [
        { protocol: 'UDP', port: 53 },
        { protocol: 'TCP', port: 53 },
      ],
    }],
  }),
  // mall-prod：支付服务出站受限（ipBlock + except）
  np({
    name: 'payment-egress-restricted', namespace: 'mall-prod', createdAt: daysAgo(14),
    podSelector: { matchLabels: { app: 'payment-service' } },
    policyTypes: ['Egress'],
    egress: [
      {
        to: [{ ipBlock: { cidr: '10.96.0.0/16', except: ['10.96.255.0/24'] } }],
        ports: [
          { protocol: 'TCP', port: 443 },
          { protocol: 'TCP', port: 8443 },
        ],
      },
      {
        to: [{ podSelector: { matchLabels: { app: 'mall-redis' } } }],
        ports: [{ protocol: 'TCP', port: 6379 }],
      },
    ],
  }),
  // mall-prod：订单服务入口（namespaceSelector + ipBlock + 端口段）
  np({
    name: 'order-ingress-policy', namespace: 'mall-prod', createdAt: daysAgo(21),
    podSelector: { matchLabels: { app: 'order-service' } },
    policyTypes: ['Ingress'],
    ingress: [
      {
        from: [{ namespaceSelector: { matchLabels: { 'kubernetes.io/metadata.name': 'mall-staging' } } }],
        ports: [{ protocol: 'TCP', port: 8081 }],
      },
      {
        from: [{ ipBlock: { cidr: '172.18.0.0/16' } }],
        ports: [{ protocol: 'TCP', port: 8081, endPort: 8090 }],
      },
    ],
  }),
  // mall-prod：指标抓取放行（matchExpressions 选择器）
  np({
    name: 'allow-metrics-scrape', namespace: 'mall-prod', createdAt: daysAgo(30),
    podSelector: {
      matchExpressions: [{ key: 'app', operator: 'In', values: ['order-service', 'payment-service'] }],
    },
    policyTypes: ['Ingress'],
    ingress: [{
      from: [{ ipBlock: { cidr: '10.0.0.0/8' } }],
      ports: [{ protocol: 'TCP', port: 9090 }],
    }],
  }),
  // klaw-test：仅入站默认拒绝
  np({
    name: 'default-deny-ingress', namespace: 'klaw-test', createdAt: daysAgo(3),
    podSelector: {},
    policyTypes: ['Ingress'],
    ingress: [],
  }),
  // klaw-test：httpbin 全命名空间入站放行（空 namespaceSelector）
  np({
    name: 'allow-httpbin-ingress', namespace: 'klaw-test', createdAt: daysAgo(3),
    podSelector: { matchLabels: { app: 'httpbin' } },
    policyTypes: ['Ingress'],
    ingress: [{
      from: [{ namespaceSelector: {} }],
      ports: [{ protocol: 'TCP', port: 80 }],
    }],
  }),
]
