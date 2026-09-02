// 工作负载：Deployment + Pod（含故障故事线）
//
// 故事线：kind-test-worker2 内存压力 → mall-prod/frontend 1 个 Pod Pending（FailedScheduling）、
// 1 个 CrashLoopBackOff → 触发 critical 告警 → 诊断页发现问题。
// 其余工作负载健康。klaw-test 下的 nginx/frontend/httpbin 为测试锚点，勿改名。
//
// 约束（被单测锁死）：Running Pod 数必须恰好 10；全集群 Deployment 名全局唯一；
// 仅 klaw-test/nginx 显示 2/2。

import { daysAgo, hoursAgo } from '../time'

interface PodSeed {
  name: string
  namespace: string
  node: string
  phase: string
  podIP: string
  labels: Record<string, string>
  createdAt: string
}

const pod = (s: PodSeed) => ({
  metadata: {
    name: s.name,
    namespace: s.namespace,
    creationTimestamp: s.createdAt,
    labels: s.labels,
  },
  spec: { nodeName: s.node },
  status: { phase: s.phase, podIP: s.podIP },
})

// ── klaw-test（测试锚点）─────────────────────────────────
// 4 Running + 1 Pending
const klawTestPods = [
  pod({ name: 'nginx-6b66fbbd46-abc12', namespace: 'klaw-test', node: 'kind-test-worker', phase: 'Running', podIP: '10.244.1.5', labels: { app: 'nginx', 'pod-template-hash': '6b66fbbd46' }, createdAt: daysAgo(2) }),
  pod({ name: 'nginx-6b66fbbd46-def34', namespace: 'klaw-test', node: 'kind-test-worker2', phase: 'Running', podIP: '10.244.2.3', labels: { app: 'nginx', 'pod-template-hash': '6b66fbbd46' }, createdAt: daysAgo(2) }),
  pod({ name: 'frontend-58cb7f74c8-xyz78', namespace: 'klaw-test', node: 'kind-test-worker', phase: 'Running', podIP: '10.244.1.21', labels: { app: 'frontend', 'pod-template-hash': '58cb7f74c8' }, createdAt: daysAgo(1) }),
  pod({ name: 'frontend-58cb7f74c8-pqr90', namespace: 'klaw-test', node: '', phase: 'Pending', podIP: '', labels: { app: 'frontend', 'pod-template-hash': '58cb7f74c8' }, createdAt: hoursAgo(3) }),
  pod({ name: 'httpbin-7556469ddd-ghi90', namespace: 'klaw-test', node: 'kind-test-control-plane', phase: 'Running', podIP: '10.244.0.8', labels: { app: 'httpbin', 'pod-template-hash': '7556469ddd' }, createdAt: daysAgo(3) }),
]

// ── mall-prod（故障故事线主战场）─────────────────────────────
// 4 Running + 1 Pending + 1 CrashLoopBackOff
const mallProdPods = [
  pod({ name: 'mall-frontend-7d9c5f8b4-z8x3c', namespace: 'mall-prod', node: 'kind-test-worker2', phase: 'CrashLoopBackOff', podIP: '10.244.2.41', labels: { app: 'mall-frontend', tier: 'web', 'pod-template-hash': '7d9c5f8b4' }, createdAt: hoursAgo(2) }),
  pod({ name: 'mall-frontend-7d9c5f8b4-q7w2e', namespace: 'mall-prod', node: '', phase: 'Pending', podIP: '', labels: { app: 'mall-frontend', tier: 'web', 'pod-template-hash': '7d9c5f8b4' }, createdAt: hoursAgo(2) }),
  pod({ name: 'mall-gateway-5b8f7d9c6-h4j6k', namespace: 'mall-prod', node: 'kind-test-worker', phase: 'Running', podIP: '10.244.1.32', labels: { app: 'mall-gateway', tier: 'gateway', 'pod-template-hash': '5b8f7d9c6' }, createdAt: daysAgo(6) }),
  pod({ name: 'order-service-6c9d8e7f5-a1b2c', namespace: 'mall-prod', node: 'kind-test-worker', phase: 'Running', podIP: '10.244.1.33', labels: { app: 'order-service', tier: 'backend', 'pod-template-hash': '6c9d8e7f5' }, createdAt: daysAgo(4) }),
  pod({ name: 'payment-service-4d5c6b7a8-q9z1x', namespace: 'mall-prod', node: 'kind-test-worker2', phase: 'Running', podIP: '10.244.2.42', labels: { app: 'payment-service', tier: 'backend', 'pod-template-hash': '4d5c6b7a8' }, createdAt: daysAgo(7) }),
  pod({ name: 'mall-redis-6b5a4c3d2-p1q2r', namespace: 'mall-prod', node: 'kind-test-worker2', phase: 'Running', podIP: '10.244.2.45', labels: { app: 'mall-redis', tier: 'cache', 'pod-template-hash': '6b5a4c3d2' }, createdAt: daysAgo(10) }),
]

// ── mall-staging（健康）────────────────────────────────────
// 2 Running
const mallStagingPods = [
  pod({ name: 'mall-frontend-staging-3e2f1a4b-v5w6x', namespace: 'mall-staging', node: 'kind-test-worker', phase: 'Running', podIP: '10.244.1.51', labels: { app: 'mall-frontend-staging', 'pod-template-hash': '3e2f1a4b' }, createdAt: daysAgo(2) }),
  pod({ name: 'order-service-staging-7a6b5c4d-y7z8a', namespace: 'mall-staging', node: 'kind-test-worker', phase: 'Running', podIP: '10.244.1.52', labels: { app: 'order-service-staging', 'pod-template-hash': '7a6b5c4d' }, createdAt: daysAgo(2) }),
]

// Running 总计：4 + 4 + 2 = 10 ✓（CrashLoopBackOff 不计入 Running）
// Pending 总计：2
export const mockPods = [
  ...klawTestPods,
  ...mallProdPods,
  ...mallStagingPods,
]

// ── Deployments ─────────────────────────────────────────────
interface DeploymentSeed {
  name: string
  namespace: string
  replicas: number
  available: number
  image: string
  app: string
  createdAt: string
  progressing?: { reason: string; message: string }
}

const deploy = (d: DeploymentSeed) => ({
  metadata: {
    name: d.name,
    namespace: d.namespace,
    creationTimestamp: d.createdAt,
    labels: { app: d.app },
  },
  spec: {
    replicas: d.replicas,
    selector: { matchLabels: { app: d.app } },
    template: {
      spec: {
        containers: [{ name: d.app, image: d.image }],
      },
    },
  },
  status: {
    replicas: d.replicas,
    availableReplicas: d.available,
    readyReplicas: d.available,
    updatedReplicas: d.replicas,
    conditions: [
      d.available >= d.replicas
        ? { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable', message: 'Deployment has minimum availability' }
        : { type: 'Available', status: 'False', reason: 'MinimumReplicasUnavailable', message: `Only ${d.available}/${d.replicas} replicas available` },
      d.progressing
        ? { type: 'Progressing', status: 'True', ...d.progressing }
        : { type: 'Progressing', status: 'True', reason: 'NewReplicaSetAvailable', message: 'ReplicaSet has successfully progressed' },
    ],
  },
})

export const mockDeployments = [
  // klaw-test（测试锚点）
  deploy({ name: 'nginx', namespace: 'klaw-test', replicas: 2, available: 2, image: 'nginx:alpine', app: 'nginx', createdAt: daysAgo(3) }),
  deploy({ name: 'frontend', namespace: 'klaw-test', replicas: 3, available: 2, image: 'httpd:alpine', app: 'frontend', createdAt: daysAgo(3), progressing: { reason: 'ReplicaSetUpdated', message: 'ReplicaSet is progressing' } }),
  deploy({ name: 'httpbin', namespace: 'klaw-test', replicas: 1, available: 1, image: 'kennethreitz/httpbin', app: 'httpbin', createdAt: daysAgo(3) }),
  // mall-prod（mall-frontend 为故障 Deployment：3 副本仅 1 可用）
  deploy({ name: 'mall-frontend', namespace: 'mall-prod', replicas: 3, available: 1, image: 'registry.local/mall/frontend:v2.4.1', app: 'mall-frontend', createdAt: daysAgo(5), progressing: { reason: 'ProgressDeadlineExceeded', message: 'ReplicaSet mall-frontend-7d9c5f8b4 has timed out progressing' } }),
  deploy({ name: 'mall-gateway', namespace: 'mall-prod', replicas: 3, available: 2, image: 'registry.local/mall/gateway:v1.9.0', app: 'mall-gateway', createdAt: daysAgo(6), progressing: { reason: 'ReplicaSetUpdated', message: 'ReplicaSet mall-gateway-5b8f7d9c6 is progressing' } }),
  deploy({ name: 'order-service', namespace: 'mall-prod', replicas: 1, available: 1, image: 'registry.local/mall/order:v3.1.2', app: 'order-service', createdAt: daysAgo(4) }),
  deploy({ name: 'payment-service', namespace: 'mall-prod', replicas: 1, available: 1, image: 'registry.local/mall/payment:v2.0.3', app: 'payment-service', createdAt: daysAgo(7) }),
  deploy({ name: 'mall-redis', namespace: 'mall-prod', replicas: 1, available: 1, image: 'redis:7-alpine', app: 'mall-redis', createdAt: daysAgo(10) }),
  // mall-staging
  deploy({ name: 'mall-frontend-staging', namespace: 'mall-staging', replicas: 1, available: 1, image: 'registry.local/mall/frontend:v2.5.0-rc.1', app: 'mall-frontend-staging', createdAt: daysAgo(2) }),
  deploy({ name: 'order-service-staging', namespace: 'mall-staging', replicas: 1, available: 1, image: 'registry.local/mall/order:v3.2.0-rc.2', app: 'order-service-staging', createdAt: daysAgo(2) }),
]

// 供 handlers 使用的可变副本（scale/restart 会改内存态）
export const deploymentsStore = mockDeployments.map((d) => structuredClone(d))