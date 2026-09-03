// Service 与 Endpoints（endpoints 由 selector × pods 在 handler 中实时派生）

import { daysAgo } from '../time'

interface ServiceSeed {
  name: string
  namespace: string
  type: string
  clusterIP: string
  ports: Array<{ name?: string; port: number; targetPort: number; protocol: string; nodePort?: number }>
  selector?: Record<string, string>
  externalIPs?: string[]
  loadBalancerIP?: string
  sessionAffinity?: string
  externalTrafficPolicy?: string
  annotations?: Record<string, string>
  createdAt: string
}

const svc = (s: ServiceSeed) => ({
  metadata: {
    name: s.name,
    namespace: s.namespace,
    creationTimestamp: s.createdAt,
    labels: s.selector ? { ...s.selector } : {},
    annotations: s.annotations,
  },
  spec: {
    type: s.type,
    clusterIP: s.clusterIP,
    externalIPs: s.externalIPs,
    ports: s.ports,
    selector: s.selector,
    sessionAffinity: s.sessionAffinity,
    externalTrafficPolicy: s.type === 'LoadBalancer' || s.type === 'NodePort' ? (s.externalTrafficPolicy ?? 'Cluster') : undefined,
  },
  status: {
    loadBalancer: s.loadBalancerIP ? { ingress: [{ ip: s.loadBalancerIP }] } : undefined,
  },
})

export const mockServices = [
  // default
  svc({ name: 'kubernetes', namespace: 'default', type: 'ClusterIP', clusterIP: '10.96.0.1', ports: [{ name: 'https', port: 443, targetPort: 6443, protocol: 'TCP' }], createdAt: daysAgo(240) }),
  // klaw-test
  svc({ name: 'nginx', namespace: 'klaw-test', type: 'ClusterIP', clusterIP: '10.96.10.1', ports: [{ name: 'http', port: 80, targetPort: 80, protocol: 'TCP' }], selector: { app: 'nginx' }, sessionAffinity: 'None', createdAt: daysAgo(3) }),
  svc({ name: 'frontend', namespace: 'klaw-test', type: 'NodePort', clusterIP: '10.96.10.2', ports: [{ name: 'http', port: 80, targetPort: 80, protocol: 'TCP', nodePort: 30080 }], selector: { app: 'frontend' }, externalTrafficPolicy: 'Cluster', createdAt: daysAgo(3) }),
  svc({ name: 'httpbin', namespace: 'klaw-test', type: 'ClusterIP', clusterIP: '10.96.10.3', ports: [{ name: 'http', port: 8000, targetPort: 80, protocol: 'TCP' }], selector: { app: 'httpbin' }, sessionAffinity: 'None', createdAt: daysAgo(3) }),
  // mall-prod
  svc({
    name: 'mall-frontend', namespace: 'mall-prod', type: 'LoadBalancer', clusterIP: '10.96.20.1',
    ports: [{ name: 'http', port: 80, targetPort: 8080, protocol: 'TCP' }], selector: { app: 'mall-frontend' },
    loadBalancerIP: '172.18.0.100', externalTrafficPolicy: 'Local', sessionAffinity: 'ClientIP',
    annotations: { 'service.beta.kubernetes.io/kind-load-balancer': 'metallb', 'nginx.ingress.kubernetes.io/affinity': 'cookie' },
    createdAt: daysAgo(5),
  }),
  svc({ name: 'mall-gateway', namespace: 'mall-prod', type: 'ClusterIP', clusterIP: '10.96.20.2', ports: [{ name: 'http', port: 8080, targetPort: 8080, protocol: 'TCP' }], selector: { app: 'mall-gateway' }, sessionAffinity: 'None', createdAt: daysAgo(6) }),
  svc({ name: 'order-service', namespace: 'mall-prod', type: 'ClusterIP', clusterIP: '10.96.20.3', ports: [{ name: 'http', port: 8081, targetPort: 8081, protocol: 'TCP' }], selector: { app: 'order-service' }, sessionAffinity: 'None', createdAt: daysAgo(4) }),
  svc({ name: 'payment-service', namespace: 'mall-prod', type: 'ClusterIP', clusterIP: '10.96.20.5', ports: [{ name: 'http', port: 8083, targetPort: 8083, protocol: 'TCP' }], selector: { app: 'payment-service' }, sessionAffinity: 'ClientIP', createdAt: daysAgo(7) }),
  svc({ name: 'mall-redis', namespace: 'mall-prod', type: 'ClusterIP', clusterIP: '10.96.20.7', ports: [{ name: 'redis', port: 6379, targetPort: 6379, protocol: 'TCP' }], selector: { app: 'mall-redis' }, sessionAffinity: 'None', createdAt: daysAgo(10) }),
  // mall-staging
  svc({ name: 'mall-frontend-staging', namespace: 'mall-staging', type: 'ClusterIP', clusterIP: '10.96.30.1', ports: [{ name: 'http', port: 80, targetPort: 8080, protocol: 'TCP' }], selector: { app: 'mall-frontend-staging' }, sessionAffinity: 'None', createdAt: daysAgo(2) }),
  // ingress-nginx
  svc({
    name: 'ingress-nginx-controller', namespace: 'ingress-nginx', type: 'LoadBalancer', clusterIP: '10.96.50.1',
    ports: [{ name: 'http', port: 80, targetPort: 80, protocol: 'TCP' }, { name: 'https', port: 443, targetPort: 443, protocol: 'TCP' }],
    selector: { app: 'ingress-nginx' }, loadBalancerIP: '172.18.0.101', externalTrafficPolicy: 'Local',
    annotations: { 'ingress-nginx.io/publish-service': 'ingress-nginx/ingress-nginx-controller' },
    createdAt: daysAgo(180),
  }),
  // kube-system
  svc({ name: 'kube-dns', namespace: 'kube-system', type: 'ClusterIP', clusterIP: '10.96.0.10', ports: [{ name: 'dns', port: 53, targetPort: 53, protocol: 'UDP' }], selector: { 'k8s-app': 'kube-dns' }, createdAt: daysAgo(240) }),
]
