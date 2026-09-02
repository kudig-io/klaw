// 节点（含 kind-test-worker2 内存压力故障节点）与节点指标

import { daysAgo } from '../time'

const readyConditions = [
  { type: 'Ready', status: 'True' },
  { type: 'MemoryPressure', status: 'False' },
  { type: 'DiskPressure', status: 'False' },
  { type: 'PIDPressure', status: 'False' },
]

export const mockNodes = [
  {
    metadata: {
      name: 'kind-test-control-plane',
      creationTimestamp: daysAgo(240),
    },
    status: {
      capacity: { cpu: '4', memory: '8Gi' },
      conditions: readyConditions,
    },
  },
  {
    metadata: {
      name: 'kind-test-worker',
      creationTimestamp: daysAgo(240),
    },
    status: {
      capacity: { cpu: '4', memory: '8Gi' },
      conditions: readyConditions,
    },
  },
  {
    // 故障节点：内存压力，故事线源头
    metadata: {
      name: 'kind-test-worker2',
      creationTimestamp: daysAgo(240),
    },
    status: {
      capacity: { cpu: '4', memory: '8Gi' },
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'MemoryPressure', status: 'True' },
        { type: 'DiskPressure', status: 'False' },
        { type: 'PIDPressure', status: 'False' },
      ],
    },
  },
]

export const mockNodeMetrics = {
  'kind-test-control-plane': {
    name: 'kind-test-control-plane',
    cpu: '4',
    memory: '8Gi',
  },
  'kind-test-worker': {
    name: 'kind-test-worker',
    cpu: '4',
    memory: '8Gi',
  },
  'kind-test-worker2': {
    name: 'kind-test-worker2',
    cpu: '4',
    memory: '8Gi',
  },
}
