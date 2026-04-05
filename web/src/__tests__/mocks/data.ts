// Mock 数据 - 用于测试

export const mockClusters = [
  {
    name: 'kind-test',
    kubeconfig: '/Users/test/.kube/config',
    context: 'kind-test',
  },
  {
    name: 'production',
    kubeconfig: '/Users/test/.kube/prod',
    context: 'prod-cluster',
  },
]

export const mockClusterStatus = {
  cluster: 'kind-test',
  nodes: {
    total: 3,
    ready: 3,
    notReady: 0,
  },
  pods: {
    total: 12,
    running: 10,
    pending: 2,
    failed: 0,
  },
  timestamp: '2026-04-01T12:00:00Z',
}

export const mockNamespaces = [
  { metadata: { name: 'default', creationTimestamp: '2026-01-01T00:00:00Z' } },
  { metadata: { name: 'klaw-test', creationTimestamp: '2026-03-01T00:00:00Z' } },
  { metadata: { name: 'kube-system', creationTimestamp: '2026-01-01T00:00:00Z' } },
  { metadata: { name: 'ingress-nginx', creationTimestamp: '2026-02-01T00:00:00Z' } },
]

export const mockPods = [
  {
    metadata: {
      name: 'nginx-6b66fbbd46-abc12',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T10:00:00Z',
    },
    spec: { nodeName: 'kind-test-worker' },
    status: { phase: 'Running', podIP: '10.244.1.5' },
  },
  {
    metadata: {
      name: 'nginx-6b66fbbd46-def34',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T10:00:00Z',
    },
    spec: { nodeName: 'kind-test-worker2' },
    status: { phase: 'Running', podIP: '10.244.2.3' },
  },
  {
    metadata: {
      name: 'frontend-58cb7f74c8-xyz78',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T09:30:00Z',
    },
    spec: { nodeName: 'kind-test-worker' },
    status: { phase: 'Pending', podIP: '' },
  },
  {
    metadata: {
      name: 'httpbin-7556469ddd-ghi90',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T09:00:00Z',
    },
    spec: { nodeName: 'kind-test-control-plane' },
    status: { phase: 'Running', podIP: '10.244.0.8' },
  },
]

export const mockNodes = [
  {
    metadata: {
      name: 'kind-test-control-plane',
      creationTimestamp: '2026-01-01T00:00:00Z',
    },
    status: {
      capacity: { cpu: '4', memory: '8Gi' },
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'MemoryPressure', status: 'False' },
      ],
    },
  },
  {
    metadata: {
      name: 'kind-test-worker',
      creationTimestamp: '2026-01-01T00:00:00Z',
    },
    status: {
      capacity: { cpu: '4', memory: '8Gi' },
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'MemoryPressure', status: 'False' },
      ],
    },
  },
  {
    metadata: {
      name: 'kind-test-worker2',
      creationTimestamp: '2026-01-01T00:00:00Z',
    },
    status: {
      capacity: { cpu: '4', memory: '8Gi' },
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'DiskPressure', status: 'False' },
      ],
    },
  },
]

export const mockDeployments = [
  {
    metadata: {
      name: 'nginx',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T08:00:00Z',
      labels: { app: 'nginx' },
    },
    spec: {
      replicas: 2,
      selector: { matchLabels: { app: 'nginx' } },
      template: {
        spec: {
          containers: [
            { name: 'nginx', image: 'nginx:alpine' },
          ],
        },
      },
    },
    status: {
      replicas: 2,
      availableReplicas: 2,
      readyReplicas: 2,
      updatedReplicas: 2,
      conditions: [
        { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable', message: 'Deployment has minimum availability' },
        { type: 'Progressing', status: 'True', reason: 'NewReplicaSetAvailable', message: 'ReplicaSet has successfully progressed' },
      ],
    },
  },
  {
    metadata: {
      name: 'frontend',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T07:00:00Z',
      labels: { app: 'frontend' },
    },
    spec: {
      replicas: 3,
      selector: { matchLabels: { app: 'frontend' } },
      template: {
        spec: {
          containers: [
            { name: 'httpd', image: 'httpd:alpine' },
          ],
        },
      },
    },
    status: {
      replicas: 3,
      availableReplicas: 2,
      readyReplicas: 2,
      updatedReplicas: 3,
      conditions: [
        { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable', message: 'Deployment has minimum availability' },
        { type: 'Progressing', status: 'True', reason: 'ReplicaSetUpdated', message: 'ReplicaSet is progressing' },
      ],
    },
  },
  {
    metadata: {
      name: 'httpbin',
      namespace: 'klaw-test',
      creationTimestamp: '2026-04-01T06:00:00Z',
      labels: { app: 'httpbin' },
    },
    spec: {
      replicas: 1,
      selector: { matchLabels: { app: 'httpbin' } },
      template: {
        spec: {
          containers: [
            { name: 'httpbin', image: 'kennethreitz/httpbin' },
          ],
        },
      },
    },
    status: {
      replicas: 1,
      availableReplicas: 1,
      readyReplicas: 1,
      updatedReplicas: 1,
      conditions: [
        { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable', message: 'Deployment has minimum availability' },
        { type: 'Progressing', status: 'True', reason: 'NewReplicaSetAvailable', message: 'ReplicaSet has successfully progressed' },
      ],
    },
  },
]

export const mockEvents = [
  {
    metadata: { name: 'nginx-123', namespace: 'klaw-test' },
    type: 'Normal',
    reason: 'Created',
    message: 'Created pod: nginx-6b66fbbd46-abc12',
    lastTimestamp: '2026-04-01T10:00:00Z',
  },
  {
    metadata: { name: 'nginx-124', namespace: 'klaw-test' },
    type: 'Warning',
    reason: 'FailedScheduling',
    message: '0/3 nodes are available',
    lastTimestamp: '2026-04-01T09:55:00Z',
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
}

export const mockMetricsHistory = [
  {
    clusterName: 'kind-test',
    timestamp: '2026-04-01T12:00:00Z',
    nodes: { total: 3, ready: 3, notReady: 0 },
    pods: { total: 12, running: 10, pending: 2, failed: 0 },
  },
  {
    clusterName: 'kind-test',
    timestamp: '2026-04-01T11:55:00Z',
    nodes: { total: 3, ready: 3, notReady: 0 },
    pods: { total: 11, running: 10, pending: 1, failed: 0 },
  },
]

export const mockAlerts = [
  {
    id: 'alert-1',
    cluster: 'kind-test',
    type: 'pod',
    level: 'warning',
    message: '2 pods are pending',
    createdAt: '2026-04-01T12:00:00Z',
    resolved: false,
  },
]
