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

export interface Deployment {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    labels?: Record<string, string>
  }
  spec: {
    replicas: number
    selector: {
      matchLabels: Record<string, string>
    }
    template: {
      spec: {
        containers: Array<{
          name: string
          image: string
        }>
      }
    }
  }
  status: {
    replicas: number
    availableReplicas: number
    readyReplicas: number
    updatedReplicas: number
    conditions: Array<{
      type: string
      status: string
      reason: string
      message: string
    }>
  }
}

export interface DeploymentStatus {
  name: string
  namespace: string
  replicas: number
  availableReplicas: number
  readyReplicas: number
  updatedReplicas: number
  conditions: Array<{
    type: string
    status: string
    reason: string
    message: string
  }>
}

export const clusterApi = {
  getClusters: () => api.get<Cluster[]>('/clusters'),
  getCluster: (name: string) => api.get<Cluster>(`/clusters/${name}`),
  getClusterStatus: (name: string) => api.get<ClusterStatus>(`/clusters/${name}/status`),
  getClusterMetrics: (name: string) => api.get(`/clusters/${name}/metrics`),
  getNamespaces: (name: string) => api.get<Namespace[]>(`/clusters/${name}/namespaces`),
}

export const podApi = {
  listPods: (cluster: string, namespace: string) => {
    // 如果 namespace 为空，获取所有命名空间的 pods
    if (!namespace) {
      return api.get<Pod[]>(`/clusters/${cluster}/pods`)
    }
    return api.get<Pod[]>(`/clusters/${cluster}/namespaces/${namespace}/pods`)
  },
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
  getStatus: (cluster: string) => api.get<any>(`/monitoring/${cluster}/status`),
  getAlerts: (cluster: string) => api.get<any[]>(`/monitoring/${cluster}/alerts`),
  getHistory: (cluster: string) => api.get<any[]>(`/monitoring/${cluster}/history`),
}

export const deploymentApi = {
  listDeployments: (cluster: string, namespace: string) => {
    // 如果 namespace 为空，获取所有命名空间的 deployments
    if (!namespace) {
      return api.get<Deployment[]>(`/clusters/${cluster}/deployments`)
    }
    return api.get<Deployment[]>(`/clusters/${cluster}/namespaces/${namespace}/deployments`)
  },
  getDeployment: (cluster: string, namespace: string, name: string) =>
    api.get<Deployment>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`),
  scaleDeployment: (cluster: string, namespace: string, name: string, replicas: number) =>
    api.post(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/scale`, { replicas }),
  restartDeployment: (cluster: string, namespace: string, name: string) =>
    api.post(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/restart`),
  getDeploymentPods: (cluster: string, namespace: string, name: string) =>
    api.get<Pod[]>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/pods`),
  getDeploymentStatus: (cluster: string, namespace: string, name: string) =>
    api.get<DeploymentStatus>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/status`),
}

export interface Service {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  spec: {
    type: string
    clusterIP: string
    externalIPs?: string[]
    ports: Array<{
      name?: string
      port: number
      targetPort: number
      protocol: string
      nodePort?: number
    }>
    selector?: Record<string, string>
  }
  status: {
    loadBalancer?: {
      ingress?: Array<{
        ip?: string
        hostname?: string
      }>
    }
  }
}

export interface ServiceEndpoints {
  serviceName: string
  namespace: string
  endpoints: Array<{
    addresses?: Array<{
      ip: string
      hostname?: string
      nodeName?: string
      targetRef?: {
        kind: string
        name: string
        namespace: string
      }
    }>
    notReadyAddresses?: Array<{
      ip: string
      hostname?: string
      nodeName?: string
      targetRef?: {
        kind: string
        name: string
        namespace: string
      }
    }>
    ports?: Array<{
      name?: string
      port: number
      protocol: string
    }>
  }>
}

export const serviceApi = {
  listServices: (cluster: string, namespace: string) => {
    // 如果 namespace 为空，获取所有命名空间的 services
    if (!namespace) {
      return api.get<Service[]>(`/clusters/${cluster}/services`)
    }
    return api.get<Service[]>(`/clusters/${cluster}/namespaces/${namespace}/services`)
  },
  getService: (cluster: string, namespace: string, name: string) =>
    api.get<Service>(`/clusters/${cluster}/namespaces/${namespace}/services/${name}`),
  getServiceEndpoints: (cluster: string, namespace: string, name: string) =>
    api.get<ServiceEndpoints>(`/clusters/${cluster}/namespaces/${namespace}/services/${name}/endpoints`),
  deleteService: (cluster: string, namespace: string, name: string) =>
    api.delete(`/clusters/${cluster}/namespaces/${namespace}/services/${name}`),
}

export default api
