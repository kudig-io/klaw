// Clusters
export interface Cluster {
  name: string
  kubeconfig: string
  context?: string
  status?: 'connected' | 'disconnected' | 'error'
  version?: string
  nodeCount?: number
}

// Pods
export interface Pod {
  name: string
  namespace: string
  status: string
  phase: string
  restarts: number
  age: string
  node?: string
  podIP?: string
  containers?: Container[]
  labels?: Record<string, string>
}

export interface Container {
  name: string
  image: string
  ready: boolean
  restartCount: number
  state: string
}

// Nodes
export interface Node {
  name: string
  status: string
  role: string
  version: string
  age: string
  internalIP: string
  osImage?: string
  kernelVersion?: string
  containerRuntime?: string
  cpu?: ResourceMetrics
  memory?: ResourceMetrics
}

export interface ResourceMetrics {
  capacity: string
  allocatable: string
  usage?: string
  usagePercent?: number
}

// Deployments
export interface Deployment {
  name: string
  namespace: string
  replicas: number
  readyReplicas: number
  availableReplicas: number
  age: string
  strategy?: string
  selector?: Record<string, string>
  labels?: Record<string, string>
  annotations?: Record<string, string>
  conditions?: DeploymentCondition[]
  containerSpec?: ContainerSpec
}

export interface DeploymentCondition {
  type: string
  status: string
  reason?: string
  message?: string
  lastUpdateTime?: string
}

export interface ContainerSpec {
  name: string
  image: string
  ports?: ContainerPort[]
  resources?: ResourceRequirements
}

export interface ContainerPort {
  name?: string
  containerPort: number
  protocol?: string
}

export interface ResourceRequirements {
  limits?: Record<string, string>
  requests?: Record<string, string>
}

export interface DeploymentStatus {
  name: string
  namespace: string
  replicas: number
  readyReplicas: number
  availableReplicas: number
  unavailableReplicas: number
  updatedReplicas: number
  conditions: DeploymentCondition[]
}

// Services
export interface Service {
  name: string
  namespace: string
  type: string
  clusterIP: string
  externalIP: string
  ports: ServicePort[]
  selector?: Record<string, string>
  age: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

export interface ServicePort {
  name?: string
  port: number
  targetPort: number | string
  protocol: string
  nodePort?: number
}

export interface ServiceEndpoints {
  serviceName: string
  namespace: string
  endpoints: EndpointSubset[]
}

export interface EndpointSubset {
  addresses?: EndpointAddress[]
  notReadyAddresses?: EndpointAddress[]
  ports?: EndpointPort[]
}

export interface EndpointAddress {
  ip: string
  hostname?: string
  nodeName?: string
  targetRef?: {
    kind: string
    name: string
    namespace: string
  }
}

export interface EndpointPort {
  name?: string
  port: number
  protocol: string
}

// Events
export interface Event {
  type: string
  reason: string
  message: string
  involvedObject: {
    kind: string
    name: string
    namespace?: string
  }
  count: number
  firstTimestamp: string
  lastTimestamp: string
}

// Monitoring
export interface ClusterStatus {
  clusterName: string
  status: 'healthy' | 'warning' | 'critical'
  nodeCount: number
  readyNodes: number
  podCount: number
  runningPods: number
  cpuUsage: number
  memoryUsage: number
  alerts: Alert[]
  lastUpdated: string
}

export interface Alert {
  id: string
  severity: 'info' | 'warning' | 'critical'
  title: string
  message: string
  resource?: string
  timestamp: string
  resolved?: boolean
}

export interface MetricsHistory {
  clusterName: string
  timeRange: string
  cpuData: DataPoint[]
  memoryData: DataPoint[]
  podData: DataPoint[]
}

export interface DataPoint {
  timestamp: string
  value: number
}

// API Response
export interface ApiResponse<T> {
  data: T
  success: boolean
  message?: string
}
