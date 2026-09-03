import axios from 'axios'

export const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

export const v1Api = axios.create({
  baseURL: '/api/v1',
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
      pods?: string
    }
    allocatable?: {
      cpu: string
      memory: string
      pods: string
    }
    addresses?: Array<{
      type: string
      address: string
    }>
    nodeInfo?: {
      machineID: string
      osImage: string
      containerRuntimeVersion: string
      kubeletVersion: string
      kubeProxyVersion: string
      architecture: string
      operatingSystem: string
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
  usage?: {
    cpuPercent: number
    memoryPercent: number
    pods: number
  }
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
          resources?: {
            requests?: Record<string, string>
            limits?: Record<string, string>
          }
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
  getClusters: () => v1Api.get<Cluster[]>('/clusters'),
  getCluster: (name: string) => v1Api.get<Cluster>(`/clusters/${name}`),
  getClusterStatus: (name: string) => v1Api.get<ClusterStatus>(`/clusters/${name}/status`),
  getClusterMetrics: (name: string) => v1Api.get(`/clusters/${name}/metrics`),
  getNamespaces: (name: string) => v1Api.get<Namespace[]>(`/clusters/${name}/namespaces`),
}

export const podApi = {
  listPods: (cluster: string, namespace: string) => {
    // 如果 namespace 为空，获取所有命名空间的 pods
    if (!namespace) {
      return v1Api.get<Pod[]>(`/clusters/${cluster}/pods`)
    }
    return v1Api.get<Pod[]>(`/clusters/${cluster}/namespaces/${namespace}/pods`)
  },
  getPod: (cluster: string, namespace: string, name: string) =>
    v1Api.get<Pod>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}`),
  getPodLogs: (cluster: string, namespace: string, name: string, tailLines?: number) =>
    v1Api.get<{ logs: string }>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}/logs`, {
      params: { tailLines },
    }),
  analyzePodLogs: (cluster: string, namespace: string, name: string, tailLines?: number) =>
    v1Api.get<LogAnalysis>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}/logs/analysis`, {
      params: { tailLines },
    }),
  deletePod: (cluster: string, namespace: string, name: string) =>
    v1Api.delete(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}`),
}

export const nodeApi = {
  listNodes: (cluster: string) => v1Api.get<Node[]>(`/clusters/${cluster}/nodes`),
  getNode: (cluster: string, name: string) => v1Api.get<Node>(`/clusters/${cluster}/nodes/${name}`),
  getNodeMetrics: (cluster: string) => v1Api.get<Record<string, NodeMetrics>>(`/clusters/${cluster}/nodes/metrics`),
}

export const eventApi = {
  getEvents: (cluster: string, namespace?: string) =>
    v1Api.get<Event[]>(namespace ? `/clusters/${cluster}/namespaces/${namespace}/events` : `/clusters/${cluster}/events`),
}

export const monitoringApi = {
  getStatus: (cluster: string) => v1Api.get<any>(`/clusters/${cluster}/monitor/status`),
  getAlerts: (cluster: string) => v1Api.get<any[]>(`/clusters/${cluster}/monitor/alerts`),
  getHistory: (cluster: string) => v1Api.get<any[]>(`/clusters/${cluster}/monitor/history`),
}

export const deploymentApi = {
  listDeployments: (cluster: string, namespace: string) => {
    // 如果 namespace 为空，获取所有命名空间的 deployments
    if (!namespace) {
      return v1Api.get<Deployment[]>(`/clusters/${cluster}/deployments`)
    }
    return v1Api.get<Deployment[]>(`/clusters/${cluster}/namespaces/${namespace}/deployments`)
  },
  getDeployment: (cluster: string, namespace: string, name: string) =>
    v1Api.get<Deployment>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`),
  scaleDeployment: (cluster: string, namespace: string, name: string, replicas: number) =>
    v1Api.post(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/scale`, { replicas }),
  restartDeployment: (cluster: string, namespace: string, name: string) =>
    v1Api.post(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/restart`),
  getDeploymentPods: (cluster: string, namespace: string, name: string) =>
    v1Api.get<Pod[]>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/pods`),
  getDeploymentStatus: (cluster: string, namespace: string, name: string) =>
    v1Api.get<DeploymentStatus>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/status`),
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
    sessionAffinity?: string
    externalTrafficPolicy?: string
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

export interface UnifiedResourceInfo {
  name: string
  namespace?: string
  kind: string
  creationTimestamp: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  status?: string
  raw?: unknown
}

export interface UnifiedResourceList {
  items: UnifiedResourceInfo[]
  total: number
  resourceKind: string
}

export const unifiedClusterApi = {
  getClusters: () => v1Api.get<Cluster[]>('/clusters'),
  getCluster: (name: string) => v1Api.get<Cluster>(`/clusters/${name}`),
  getClusterStatus: (name: string) => v1Api.get<ClusterStatus>(`/clusters/${name}/status`),
  getClusterMetrics: (name: string) => v1Api.get(`/clusters/${name}/metrics`),
  getNamespaces: (name: string) => v1Api.get<Namespace[]>(`/clusters/${name}/namespaces`),
}

export const unifiedResourceApi = {
  listResources: (cluster: string, kind: string, namespace?: string) => {
    if (namespace) {
      return v1Api.get<UnifiedResourceList>(`/clusters/${cluster}/namespaces/${namespace}/resources/${kind}`)
    }
    return v1Api.get<UnifiedResourceList>(`/clusters/${cluster}/resources/${kind}`)
  },
  getResource: (cluster: string, kind: string, name: string, namespace?: string) => {
    if (namespace) {
      return v1Api.get<UnifiedResourceInfo>(`/clusters/${cluster}/namespaces/${namespace}/resources/${kind}/${name}`)
    }
    return v1Api.get<UnifiedResourceInfo>(`/clusters/${cluster}/resources/${kind}/${name}`)
  },
  getMonitorStatus: (cluster: string) => v1Api.get(`/clusters/${cluster}/monitor/status`),
  getMonitorAlerts: (cluster: string) => v1Api.get(`/clusters/${cluster}/monitor/alerts`),
  getMonitorHistory: (cluster: string) => v1Api.get(`/clusters/${cluster}/monitor/history`),
}

export const serviceApi = {
  listServices: (cluster: string, namespace: string) => {
    // 如果 namespace 为空，获取所有命名空间的 services
    if (!namespace) {
      return v1Api.get<Service[]>(`/clusters/${cluster}/services`)
    }
    return v1Api.get<Service[]>(`/clusters/${cluster}/namespaces/${namespace}/services`)
  },
  getService: (cluster: string, namespace: string, name: string) =>
    v1Api.get<Service>(`/clusters/${cluster}/namespaces/${namespace}/services/${name}`),
  getServiceEndpoints: (cluster: string, namespace: string, name: string) =>
    v1Api.get<ServiceEndpoints>(`/clusters/${cluster}/namespaces/${namespace}/services/${name}/endpoints`),
  deleteService: (cluster: string, namespace: string, name: string) =>
    v1Api.delete(`/clusters/${cluster}/namespaces/${namespace}/services/${name}`),
}

export interface LogAnalysisEntry {
  line: number
  content: string
  timestamp?: string
  level: string
}

export interface SecurityEvent {
  type: string
  message: string
  severity: string
}

export interface SlowRequest {
  url: string
  responseTime: string
}

export interface LogAnalysis {
  totalLines: number
  errorCount: number
  warningCount: number
  infoCount: number
  debugCount: number
  errors?: LogAnalysisEntry[]
  warnings?: LogAnalysisEntry[]
  stackTraces?: string[]
  performanceMetrics: {
    slowRequests?: SlowRequest[]
  }
  securityEvents?: SecurityEvent[]
  logLevels: Record<string, number>
  patternStats: Record<string, number>
}

export interface RBACAnalysis {
  totalRoles: number
  totalClusterRoles: number
  totalBindings: number
  totalClusterBindings: number
  rolesByNamespace: Record<string, string[]>
  bindingsBySubject: Record<string, Array<{
    type: string
    name: string
    namespace: string
    role: string
    roleKind: string
  }>>
  bindingsByRole: Record<string, Array<{
    type: string
    name: string
    namespace: string
    subjects: string[]
  }>>
  timestamp: string
}

export const analysisApi = {
  analyzeLogs: (logs: string) => v1Api.post<LogAnalysis>('/analysis/logs', { logs }),
  analyzeRBAC: (cluster: string) => v1Api.get<RBACAnalysis>(`/clusters/${cluster}/rbac/analysis`),
}

export interface AlertRuleCondition {
  type: string
  field: string
  operator: string
  threshold: string | number | boolean
  timeWindow?: string
}

export interface AlertRule {
  id: string
  cluster?: string
  name: string
  description?: string
  enabled: boolean
  severity: 'info' | 'warning' | 'error' | 'critical'
  condition: AlertRuleCondition
  actions?: string[]
  createdAt: string
  updatedAt: string
}

export interface AlertRecord {
  id: string
  cluster: string
  ruleId: string
  ruleName: string
  ruleType: string
  resourceKind: string
  resourceName: string
  namespace?: string
  severity: 'info' | 'warning' | 'error' | 'critical'
  value: string | number | boolean
  threshold: string | number | boolean
  operator: string
  message: string
  acknowledged: boolean
  resolved: boolean
  createdAt: string
  acknowledgedAt?: string
  resolvedAt?: string
}

export interface AlertStats {
  total: number
  active: number
  bySeverity: Record<string, number>
  byStatus: Record<string, number>
  recent24h: number
}

export const alertingApi = {
  getRules: (cluster: string) => v1Api.get<AlertRule[]>(`/clusters/${cluster}/alerts/rules`),
  createRule: (cluster: string, rule: Partial<AlertRule>) => v1Api.post<AlertRule>(`/clusters/${cluster}/alerts/rules`, rule),
  updateRule: (cluster: string, id: string, rule: Partial<AlertRule>) => v1Api.put<AlertRule>(`/clusters/${cluster}/alerts/rules/${id}`, rule),
  deleteRule: (cluster: string, id: string) => v1Api.delete(`/clusters/${cluster}/alerts/rules/${id}`),
  evaluate: (cluster: string) => v1Api.post<AlertRecord[]>(`/clusters/${cluster}/alerts/evaluate`),
  getHistory: (cluster: string, limit = 50) => v1Api.get<AlertRecord[]>(`/clusters/${cluster}/alerts/history`, { params: { limit } }),
  getStats: (cluster: string) => v1Api.get<AlertStats>(`/clusters/${cluster}/alerts/stats`),
  acknowledge: (cluster: string, id: string) => v1Api.post<AlertRecord>(`/clusters/${cluster}/alerts/${id}/acknowledge`),
  resolve: (cluster: string, id: string) => v1Api.post<AlertRecord>(`/clusters/${cluster}/alerts/${id}/resolve`),
}

export type BackupMode = 'Full' | 'Incremental'
export type StorageProvider = 'S3' | 'OSS' | 'GCS' | 'Azure'

export interface BackupStorageLocation {
  provider: StorageProvider
  bucket: string
  prefix?: string
  region: string
  endpoint?: string
  credentialsSecret: string
}

export interface BackupValidationConfig {
  enabled: boolean
  consistencyCheck: boolean
}

export interface BackupItem {
  name: string
  cluster: string
  phase: string
  spec: {
    backupMode: BackupMode
    etcdEndpoints?: string[]
    storageLocation: BackupStorageLocation
    validation: BackupValidationConfig
  }
  snapshotSize: number
  snapshotLocation: string
  etcdRevision: number
  validationResult?: {
    valid: boolean
    hash?: string
    message?: string
  }
  startTime: string
  completionTime?: string
  message?: string
  createdAt: string
}

export interface BackupSummary {
  total: number
  byPhase: Record<string, number>
  byMode: Record<string, number>
  recent24h: number
}

export interface CreateBackupRequest {
  name: string
  backupMode: BackupMode
  etcdEndpoints?: string[]
  storageLocation: BackupStorageLocation
  validation: BackupValidationConfig
}

export const backupApi = {
  list: (cluster: string) => v1Api.get<BackupItem[]>(`/clusters/${cluster}/backups`),
  get: (cluster: string, name: string) => v1Api.get<BackupItem>(`/clusters/${cluster}/backups/${name}`),
  create: (cluster: string, request: CreateBackupRequest) => v1Api.post<BackupItem>(`/clusters/${cluster}/backups`, request),
  delete: (cluster: string, name: string) => v1Api.delete(`/clusters/${cluster}/backups/${name}`),
  summary: (cluster: string) => v1Api.get<BackupSummary>(`/clusters/${cluster}/backups/summary`),
}

export interface Tenant {
  id: string
  cluster?: string
  name: string
  description?: string
  namespaces: string[]
  resourceQuotas: {
    cpu: string
    memory: string
    pods: string
    services: string
    persistentVolumeClaims: string
  }
  networkPolicies: {
    enabled: boolean
    defaultDeny: boolean
  }
  rbac: {
    enabled: boolean
    defaultRole: string
  }
  createdAt: string
  updatedAt: string
}

export interface TenantUser {
  id: string
  tenantId: string
  username: string
  email?: string
  role: string
  namespaces?: string[]
  subjectKind?: 'User' | 'Group' | 'ServiceAccount'
  subjectName?: string
  subjectNamespace?: string
  createdAt: string
}

export interface TenantStatistics {
  totalTenants: number
  totalUsers: number
  totalNamespaces: number
  usersByRole: Record<string, number>
}

export interface AuditLog {
  id: string
  timestamp: string
  eventType: string
  category: string
  severity: string
  source: string
  user: string
  action: string
  resource: Record<string, string>
  result: string
  details?: Record<string, unknown>
  ipAddress?: string
  userAgent?: string
}

export interface AuditStatistics {
  totalLogs: number
  byEventType: Record<string, number>
  bySeverity: Record<string, number>
  byCategory: Record<string, number>
  byUser: Record<string, number>
  recent24h: number
}

export const tenancyApi = {
  listTenants: (params?: { cluster?: string; name?: string; namespace?: string }) => v1Api.get<Tenant[]>('/tenants', { params }),
  getTenant: (id: string) => v1Api.get<Tenant>(`/tenants/${id}`),
  createTenant: (tenant: Partial<Tenant>) => v1Api.post<Tenant>('/tenants', tenant),
  updateTenant: (id: string, tenant: Partial<Tenant>) => v1Api.put<Tenant>(`/tenants/${id}`, tenant),
  deleteTenant: (id: string) => v1Api.delete(`/tenants/${id}`),
  stats: () => v1Api.get<TenantStatistics>('/tenants/stats'),
  listUsers: (params?: { tenantId?: string; role?: string }) => v1Api.get<TenantUser[]>('/tenant-users', { params }),
  createUser: (user: Partial<TenantUser>) => v1Api.post<TenantUser>('/tenant-users', user),
  deleteUser: (id: string) => v1Api.delete(`/tenant-users/${id}`),
}

export const auditApi = {
  listLogs: (params?: { eventType?: string; category?: string; severity?: string; user?: string; limit?: number }) =>
    v1Api.get<AuditLog[]>('/audit/logs', { params }),
  stats: () => v1Api.get<AuditStatistics>('/audit/stats'),
}

export interface Ingress {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  spec: {
    ingressClassName?: string
    defaultBackend?: {
      service?: { name: string; port?: { number?: number; name?: string } }
      resource?: { apiGroup?: string; kind: string; name: string }
    }
    tls?: Array<{
      hosts?: string[]
      secretName?: string
    }>
    rules: Array<{
      host?: string
      http?: {
        paths: Array<{
          path: string
          pathType: 'Exact' | 'Prefix' | 'ImplementationSpecific'
          backend: {
            service?: { name: string; port?: { number?: number; name?: string } }
            resource?: { apiGroup?: string; kind: string; name: string }
          }
        }>
      }
    }>
  }
  status: {
    loadBalancer?: {
      ingress?: Array<{ ip?: string; hostname?: string }>
    }
  }
}

export interface NetworkPolicyPeer {
  podSelector?: {
    matchLabels?: Record<string, string>
    matchExpressions?: Array<{ key: string; operator: string; values?: string[] }>
  }
  namespaceSelector?: {
    matchLabels?: Record<string, string>
    matchExpressions?: Array<{ key: string; operator: string; values?: string[] }>
  }
  ipBlock?: {
    cidr: string
    except?: string[]
  }
}

export interface NetworkPolicyPort {
  protocol: string
  port: number | string
  endPort?: number
}

export interface NetworkPolicy {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    labels?: Record<string, string>
  }
  spec: {
    podSelector: {
      matchLabels?: Record<string, string>
      matchExpressions?: Array<{ key: string; operator: string; values?: string[] }>
    }
    policyTypes: Array<'Ingress' | 'Egress'>
    ingress?: Array<{
      from?: NetworkPolicyPeer[]
      ports?: NetworkPolicyPort[]
    }>
    egress?: Array<{
      to?: NetworkPolicyPeer[]
      ports?: NetworkPolicyPort[]
    }>
  }
}

export interface PersistentVolumeClaim {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
    finalizers?: string[]
    deletionTimestamp?: string
  }
  spec: {
    accessModes: string[]
    storageClassName?: string
    volumeName?: string
    volumeMode?: 'Filesystem' | 'Block'
    resources: {
      requests: { storage: string }
      limits?: { storage: string }
    }
    dataSource?: { apiGroup?: string; kind: string; name: string }
    selector?: { matchLabels?: Record<string, string> }
  }
  status: {
    phase: 'Pending' | 'Bound' | 'Lost' | 'Terminating'
    capacity?: { storage: string }
    accessModes?: string[]
  }
}

export interface PersistentVolume {
  metadata: {
    name: string
    creationTimestamp: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
    finalizers?: string[]
  }
  spec: {
    capacity: { storage: string }
    accessModes: string[]
    storageClassName?: string
    persistentVolumeReclaimPolicy?: 'Retain' | 'Delete' | 'Recycle'
    volumeMode?: 'Filesystem' | 'Block'
    claimRef?: { kind?: string; namespace: string; name: string; apiVersion?: string }
    hostPath?: { path: string; type?: string }
    nfs?: { server: string; path: string; readOnly?: boolean }
    csi?: {
      driver: string
      volumeHandle: string
      fsType?: string
      readOnly?: boolean
      volumeAttributes?: Record<string, string>
    }
    local?: { path: string }
  }
  status: {
    phase: 'Available' | 'Bound' | 'Released' | 'Failed'
    reason?: string
    message?: string
  }
}

export interface StorageClass {
  metadata: {
    name: string
    creationTimestamp: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  provisioner: string
  parameters?: Record<string, string>
  reclaimPolicy?: 'Delete' | 'Retain'
  volumeBindingMode?: 'Immediate' | 'WaitForFirstConsumer'
  allowVolumeExpansion?: boolean
  mountOptions?: string[]
}

export interface NetworkAnalysis {
  totalNetworkPolicies: number
  totalServices: number
  totalIngresses: number
  policiesByNamespace: Record<string, string[]>
  servicesByType: Record<string, number>
  ingressesByHost: Record<string, string[]>
  exposedServices: Array<{
    name: string
    namespace: string
    type: string
    ports: Array<{
      name?: string
      port: number
      targetPort: number
      protocol: string
      nodePort?: number
    }>
  }>
  timestamp: string
}

export interface StorageAnalysis {
  totalPVs: number
  totalPVCs: number
  totalStorageClasses: number
  pvByStatus: Record<string, number>
  pvcByStatus: Record<string, number>
  pvByStorageClass: Record<string, number>
  storageCapacity: {
    totalBytes: number
    usedBytes: number
    availableBytes: number
  }
  scByProvisioner: Record<string, number>
  timestamp: string
}

export const networkApi = {
  listIngresses: (cluster: string, namespace: string) => {
    if (!namespace) {
      return v1Api.get<Ingress[]>(`/clusters/${cluster}/ingresses`)
    }
    return v1Api.get<Ingress[]>(`/clusters/${cluster}/namespaces/${namespace}/ingresses`)
  },
  listNetworkPolicies: (cluster: string, namespace: string) => {
    if (!namespace) {
      return v1Api.get<NetworkPolicy[]>(`/clusters/${cluster}/networkpolicies`)
    }
    return v1Api.get<NetworkPolicy[]>(`/clusters/${cluster}/namespaces/${namespace}/networkpolicies`)
  },
  getNetworkAnalysis: () => v1Api.get<NetworkAnalysis>('/analysis/network'),
}

export const storageApi = {
  listPVCs: (cluster: string, namespace: string) => {
    if (!namespace) {
      return v1Api.get<PersistentVolumeClaim[]>(`/clusters/${cluster}/persistentvolumeclaims`)
    }
    return v1Api.get<PersistentVolumeClaim[]>(`/clusters/${cluster}/namespaces/${namespace}/persistentvolumeclaims`)
  },
  listPVs: (cluster: string) => v1Api.get<PersistentVolume[]>(`/clusters/${cluster}/persistentvolumes`),
  listStorageClasses: (cluster: string) => v1Api.get<StorageClass[]>(`/clusters/${cluster}/storageclasses`),
  getStorageAnalysis: () => v1Api.get<StorageAnalysis>('/analysis/storage'),
}

export default api
