import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

export interface Cluster {
  name: string
  kubeconfig: string
  context: string
}

export interface ClusterStatus {
  cluster: string
  nodes: {
    total: number
    ready: number
    notReady: number
  }
  pods: {
    total: number
    running: number
    pending: number
    failed: number
  }
  timestamp: string
}

export interface Namespace {
  metadata: {
    name: string
    creationTimestamp: string
  }
}

export interface Pod {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
  }
  spec: {
    nodeName: string
  }
  status: {
    phase: string
    podIP: string
  }
}

export interface Node {
  metadata: {
    name: string
    creationTimestamp: string
  }
  status: {
    capacity: {
      cpu: string
      memory: string
    }
    conditions: Array<{
      type: string
      status: string
    }>
  }
}

export interface NodeMetrics {
  name: string
  cpu: string
  memory: string
}

export interface Event {
  metadata: {
    name: string
    namespace: string
  }
  type: string
  reason: string
  message: string
  lastTimestamp: string
}

export const clusterApi = {
  getClusters: () => api.get<Cluster[]>('/clusters'),
  getCluster: (name: string) => api.get<Cluster>(`/clusters/${name}`),
  getClusterStatus: (name: string) => api.get<ClusterStatus>(`/clusters/${name}/status`),
  getClusterMetrics: (name: string) => api.get(`/clusters/${name}/metrics`),
  getNamespaces: (name: string) => api.get<Namespace[]>(`/clusters/${name}/namespaces`),
}

export const podApi = {
  listPods: (cluster: string, namespace: string) =>
    api.get<Pod[]>(`/clusters/${cluster}/namespaces/${namespace}/pods`),
  getPod: (cluster: string, namespace: string, name: string) =>
    api.get<Pod>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}`),
  getPodLogs: (cluster: string, namespace: string, name: string, tailLines?: number) =>
    api.get<{ logs: string }>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}/logs`, {
      params: { tailLines },
    }),
  deletePod: (cluster: string, namespace: string, name: string) =>
    api.delete(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}`),
}

export const nodeApi = {
  listNodes: (cluster: string) => api.get<Node[]>(`/clusters/${cluster}/nodes`),
  getNode: (cluster: string, name: string) => api.get<Node>(`/clusters/${cluster}/nodes/${name}`),
  getNodeMetrics: (cluster: string) => api.get<Record<string, NodeMetrics>>(`/clusters/${cluster}/nodes/metrics`),
}

export const eventApi = {
  getEvents: (cluster: string, namespace?: string) =>
    api.get<Event[]>(namespace ? `/clusters/${cluster}/namespaces/${namespace}/events` : `/clusters/${cluster}/events`),
}

export const monitoringApi = {
  getStatus: (cluster: string) => api.get(`/monitoring/${cluster}/status`),
  getAlerts: (cluster: string) => api.get(`/monitoring/${cluster}/alerts`),
  getHistory: (cluster: string) => api.get(`/monitoring/${cluster}/history`),
}

export default api
