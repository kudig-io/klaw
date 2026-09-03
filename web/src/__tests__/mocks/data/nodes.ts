// 节点（含 kind-test-worker2 内存压力故障节点）与节点指标
//
// 故事线交叉引用：worker2 usage.memoryPercent = 93 ↔ 告警记录 value '93%'（阈值 90%）。
// mockNodeMetrics.cpu 保留容量字符串（'4'，单测锁死），usage 携带实时使用率。

import { daysAgo } from '../time'

const readyConditions = [
  { type: 'Ready', status: 'True' },
  { type: 'MemoryPressure', status: 'False' },
  { type: 'DiskPressure', status: 'False' },
  { type: 'PIDPressure', status: 'False' },
]

const nodeStatus = (ip: string, extra?: { conditions?: typeof readyConditions }) => ({
  capacity: { cpu: '4', memory: '8Gi', pods: '110' },
  allocatable: { cpu: '3800m', memory: '7592Mi', pods: '110' },
  conditions: extra?.conditions ?? readyConditions,
  addresses: [
    { type: 'InternalIP', address: ip },
    { type: 'Hostname', address: 'kind' },
  ],
  nodeInfo: {
    machineID: 'kind-control-plane',
    osImage: 'Debian GNU/Linux 12 (bookworm)',
    containerRuntimeVersion: 'containerd://1.7.27',
    kubeletVersion: 'v1.32.5',
    kubeProxyVersion: 'v1.32.5',
    architecture: 'arm64',
    operatingSystem: 'linux',
  },
})

export const mockNodes = [
  {
    metadata: {
      name: 'kind-test-control-plane',
      creationTimestamp: daysAgo(240),
      labels: { 'node-role.kubernetes.io/control-plane': '' },
    },
    status: nodeStatus('172.18.0.2'),
  },
  {
    metadata: {
      name: 'kind-test-worker',
      creationTimestamp: daysAgo(240),
    },
    status: nodeStatus('172.18.0.3'),
  },
  {
    // 故障节点：内存压力，故事线源头
    metadata: {
      name: 'kind-test-worker2',
      creationTimestamp: daysAgo(240),
    },
    status: nodeStatus('172.18.0.4', {
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'MemoryPressure', status: 'True' },
        { type: 'DiskPressure', status: 'False' },
        { type: 'PIDPressure', status: 'False' },
      ],
    }),
  },
]

export const mockNodeMetrics = {
  'kind-test-control-plane': {
    name: 'kind-test-control-plane',
    cpu: '4',
    memory: '8Gi',
    usage: { cpuPercent: 38, memoryPercent: 62, pods: 4 },
  },
  'kind-test-worker': {
    name: 'kind-test-worker',
    cpu: '4',
    memory: '8Gi',
    usage: { cpuPercent: 45, memoryPercent: 71, pods: 6 },
  },
  'kind-test-worker2': {
    name: 'kind-test-worker2',
    cpu: '4',
    memory: '8Gi',
    // 93% ↔ 告警 '93%'（> 90% 阈值，内存压力源头）
    usage: { cpuPercent: 22, memoryPercent: 93, pods: 5 },
  },
}
