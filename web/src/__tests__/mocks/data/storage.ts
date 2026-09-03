// PVC / PV / StorageClass mock 数据（覆盖 Bound/Pending/Released/Failed、Retain/Delete、
// Immediate/WaitForFirstConsumer、CSI/hostPath/NFS、可扩容、mountOptions、parameters 等形态）

import { daysAgo } from '../time'

// ── PersistentVolumeClaim ─────────────────────────────────

interface PVClaimSeed {
  name: string
  namespace: string
  createdAt: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  accessModes?: string[]
  storage: string
  storageClassName?: string
  volumeName?: string
  volumeMode?: 'Filesystem' | 'Block'
  dataSource?: { apiGroup?: string; kind: string; name: string }
  phase: 'Pending' | 'Bound' | 'Lost' | 'Terminating'
  capacity?: string
  deletionTimestamp?: string
  finalizers?: string[]
}

const pvc = (s: PVClaimSeed) => ({
  metadata: {
    name: s.name,
    namespace: s.namespace,
    creationTimestamp: s.createdAt,
    labels: s.labels,
    annotations: s.annotations,
    finalizers: s.finalizers,
    deletionTimestamp: s.deletionTimestamp,
  },
  spec: {
    accessModes: s.accessModes ?? ['ReadWriteOnce'],
    storageClassName: s.storageClassName,
    volumeName: s.volumeName,
    volumeMode: s.volumeMode ?? 'Filesystem',
    resources: { requests: { storage: s.storage } },
    dataSource: s.dataSource,
  },
  status: {
    phase: s.phase,
    capacity: s.capacity ? { storage: s.capacity } : undefined,
    accessModes: s.phase === 'Bound' ? (s.accessModes ?? ['ReadWriteOnce']) : undefined,
  },
})

export const mockPVCs = [
  // mall-prod：典型 Bound 卷
  pvc({
    name: 'mall-frontend-data', namespace: 'mall-prod', createdAt: daysAgo(5),
    labels: { app: 'mall-frontend' },
    annotations: {
      'pv.kubernetes.io/bind-completed': 'yes',
      'volume.beta.kubernetes.io/storage-provisioner': 'rbd.csi.ceph.com',
      'volume.kubernetes.io/storage-provisioner': 'rbd.csi.ceph.com',
    },
    storage: '10Gi', storageClassName: 'fast-ssd', volumeName: 'pvc-9f2c8a01-mall-frontend',
    phase: 'Bound', capacity: '10Gi',
    finalizers: ['kubernetes.io/pvc-protection'],
  }),
  pvc({
    name: 'mall-redis-data', namespace: 'mall-prod', createdAt: daysAgo(10),
    labels: { app: 'mall-redis' },
    storage: '5Gi', storageClassName: 'fast-ssd', volumeName: 'pvc-2b7d4e02-mall-redis',
    phase: 'Bound', capacity: '5Gi',
    finalizers: ['kubernetes.io/pvc-protection'],
  }),
  pvc({
    name: 'order-service-db', namespace: 'mall-prod', createdAt: daysAgo(4),
    labels: { app: 'order-service' },
    storage: '20Gi', storageClassName: 'fast-ssd', volumeName: 'pvc-77e1a903-order-db',
    phase: 'Bound', capacity: '20Gi',
    finalizers: ['kubernetes.io/pvc-protection'],
  }),
  // mall-prod：共享文件存储（RWX）
  pvc({
    name: 'payment-service-logs', namespace: 'mall-prod', createdAt: daysAgo(7),
    labels: { app: 'payment-service', tier: 'logging' },
    annotations: { 'volume.kubernetes.io/storage-provisioner': 'nfs.csi.k8s.io' },
    accessModes: ['ReadWriteMany'],
    storage: '30Gi', storageClassName: 'nfs-retained', volumeName: 'nfs-shared-media',
    phase: 'Bound', capacity: '30Gi',
    finalizers: ['kubernetes.io/pvc-protection'],
  }),
  // mall-staging：Pending（StorageClass 为 WaitForFirstConsumer，等待首个消费者）
  pvc({
    name: 'staging-cache', namespace: 'mall-staging', createdAt: daysAgo(1),
    labels: { app: 'mall-frontend-staging' },
    storage: '15Gi', storageClassName: 'standard',
    phase: 'Pending',
  }),
  // klaw-test：小容量本地卷
  pvc({
    name: 'nginx-html', namespace: 'klaw-test', createdAt: daysAgo(3),
    labels: { app: 'nginx' },
    storage: '1Gi', storageClassName: 'standard', volumeName: 'local-pv-frontend-node1',
    phase: 'Bound', capacity: '1Gi',
    finalizers: ['kubernetes.io/pvc-protection'],
  }),
  // default：正在删除的遗留卷
  pvc({
    name: 'legacy-archive', namespace: 'default', createdAt: daysAgo(120),
    storage: '2Gi', storageClassName: 'standard', volumeName: 'pvc-old-released',
    phase: 'Terminating', capacity: '2Gi',
    deletionTimestamp: daysAgo(1),
    finalizers: ['kubernetes.io/pvc-protection'],
  }),
]

// ── PersistentVolume ──────────────────────────────────────

interface PVSeed {
  name: string
  createdAt: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  capacity: string
  accessModes?: string[]
  storageClassName?: string
  reclaimPolicy?: 'Retain' | 'Delete'
  volumeMode?: 'Filesystem' | 'Block'
  claimRef?: { namespace: string; name: string }
  hostPath?: { path: string; type?: string }
  nfs?: { server: string; path: string }
  csi?: { driver: string; volumeHandle: string; fsType?: string; volumeAttributes?: Record<string, string> }
  local?: { path: string }
  phase: 'Available' | 'Bound' | 'Released' | 'Failed'
  reason?: string
  message?: string
}

const pv = (s: PVSeed) => ({
  metadata: {
    name: s.name,
    creationTimestamp: s.createdAt,
    labels: s.labels,
    annotations: s.annotations,
    finalizers: s.claimRef && s.phase === 'Bound' ? ['kubernetes.io/pv-protection'] : undefined,
  },
  spec: {
    capacity: { storage: s.capacity },
    accessModes: s.accessModes ?? ['ReadWriteOnce'],
    storageClassName: s.storageClassName,
    persistentVolumeReclaimPolicy: s.reclaimPolicy ?? 'Delete',
    volumeMode: s.volumeMode ?? 'Filesystem',
    claimRef: s.claimRef,
    hostPath: s.hostPath,
    nfs: s.nfs,
    csi: s.csi,
    local: s.local,
  },
  status: {
    phase: s.phase,
    reason: s.reason,
    message: s.message,
  },
})

export const mockPVs = [
  // Bound ×3（CSI）
  pv({
    name: 'pvc-9f2c8a01-mall-frontend', createdAt: daysAgo(5),
    capacity: '10Gi', storageClassName: 'fast-ssd',
    reclaimPolicy: 'Delete',
    claimRef: { namespace: 'mall-prod', name: 'mall-frontend-data' },
    csi: {
      driver: 'rbd.csi.ceph.com',
      volumeHandle: '0001-0009-rook-ceph-0000000000000001-9f2c8a01',
      fsType: 'ext4',
      volumeAttributes: { clusterID: 'rook-ceph', pool: 'replicapool', storageClass: 'fast-ssd' },
    },
    phase: 'Bound',
  }),
  pv({
    name: 'pvc-2b7d4e02-mall-redis', createdAt: daysAgo(10),
    capacity: '5Gi', storageClassName: 'fast-ssd',
    reclaimPolicy: 'Delete',
    claimRef: { namespace: 'mall-prod', name: 'mall-redis-data' },
    csi: {
      driver: 'rbd.csi.ceph.com',
      volumeHandle: '0001-0009-rook-ceph-0000000000000001-2b7d4e02',
      fsType: 'ext4',
      volumeAttributes: { clusterID: 'rook-ceph', pool: 'replicapool', storageClass: 'fast-ssd' },
    },
    phase: 'Bound',
  }),
  pv({
    name: 'pvc-77e1a903-order-db', createdAt: daysAgo(4),
    capacity: '20Gi', storageClassName: 'fast-ssd',
    reclaimPolicy: 'Delete',
    claimRef: { namespace: 'mall-prod', name: 'order-service-db' },
    csi: {
      driver: 'rbd.csi.ceph.com',
      volumeHandle: '0001-0009-rook-ceph-0000000000000001-77e1a903',
      fsType: 'xfs',
      volumeAttributes: { clusterID: 'rook-ceph', pool: 'replicapool', imageFeatures: 'layering' },
    },
    phase: 'Bound',
  }),
  // Available：NFS 共享卷（RWX）
  pv({
    name: 'nfs-shared-media', createdAt: daysAgo(90),
    capacity: '100Gi', accessModes: ['ReadWriteMany'], storageClassName: 'nfs-retained',
    reclaimPolicy: 'Retain',
    nfs: { server: '192.168.1.100', path: '/exports/media' },
    phase: 'Available',
  }),
  // Available：本地卷（hostPath，kind 集群风格）
  pv({
    name: 'local-pv-frontend-node1', createdAt: daysAgo(3),
    labels: { 'app.kubernetes.io/managed-by': 'local-path-provisioner' },
    capacity: '50Gi', storageClassName: 'standard',
    reclaimPolicy: 'Delete',
    hostPath: { path: '/var/local-path-provisioner/frontend-node1', type: 'DirectoryOrCreate' },
    phase: 'Available',
  }),
  // Released：claim 已删除，等待人工回收
  pv({
    name: 'pvc-old-released', createdAt: daysAgo(120),
    annotations: { 'pv.kubernetes.io/bound-by-controller': 'yes' },
    capacity: '8Gi', storageClassName: 'standard',
    reclaimPolicy: 'Retain',
    claimRef: { namespace: 'default', name: 'legacy-archive' },
    hostPath: { path: '/var/local-path-provisioner/legacy-archive', type: 'Directory' },
    phase: 'Released',
    reason: 'VolumeReleased',
    message: 'Claim 已删除，等待 Retain 策略下的人工回收',
  }),
  // Failed：回收失败
  pv({
    name: 'pvc-failed-detach', createdAt: daysAgo(15),
    capacity: '4Gi', storageClassName: 'fast-ssd',
    reclaimPolicy: 'Delete',
    csi: {
      driver: 'rbd.csi.ceph.com',
      volumeHandle: '0001-0009-rook-ceph-0000000000000001-aabbccdd',
      fsType: 'ext4',
    },
    phase: 'Failed',
    reason: 'VolumeFailedDelete',
    message: 'rbd image 删除失败：pool replicapool 中存在快照依赖',
  }),
]

// ── StorageClass ──────────────────────────────────────────

interface SCSeed {
  name: string
  createdAt: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  provisioner: string
  parameters?: Record<string, string>
  reclaimPolicy?: 'Delete' | 'Retain'
  volumeBindingMode?: 'Immediate' | 'WaitForFirstConsumer'
  allowVolumeExpansion?: boolean
  mountOptions?: string[]
}

const sc = (s: SCSeed) => ({
  metadata: {
    name: s.name,
    creationTimestamp: s.createdAt,
    labels: s.labels,
    annotations: s.annotations,
  },
  provisioner: s.provisioner,
  parameters: s.parameters,
  reclaimPolicy: s.reclaimPolicy ?? 'Delete',
  volumeBindingMode: s.volumeBindingMode ?? 'Immediate',
  allowVolumeExpansion: s.allowVolumeExpansion ?? false,
  mountOptions: s.mountOptions,
})

export const mockStorageClasses = [
  // kind 默认本地存储
  sc({
    name: 'standard', createdAt: daysAgo(240),
    annotations: { 'storageclass.kubernetes.io/is-default-class': 'true' },
    provisioner: 'rancher.io/local-path',
    reclaimPolicy: 'Delete', volumeBindingMode: 'WaitForFirstConsumer',
  }),
  // Ceph RBD：SSD 块存储，可扩容
  sc({
    name: 'fast-ssd', createdAt: daysAgo(200),
    labels: { tier: 'performance' },
    provisioner: 'rbd.csi.ceph.com',
    parameters: {
      clusterID: 'rook-ceph',
      pool: 'replicapool',
      imageFeatures: 'layering',
      'csi.storage.k8s.io/provisioner-secret-name': 'rook-csi-rbd-provisioner',
      'csi.storage.k8s.io/provisioner-secret-namespace': 'rook-ceph',
      'csi.storage.k8s.io/node-stage-secret-name': 'rook-csi-rbd-node',
      'csi.storage.k8s.io/node-stage-secret-namespace': 'rook-ceph',
    },
    reclaimPolicy: 'Delete', volumeBindingMode: 'Immediate', allowVolumeExpansion: true,
    mountOptions: ['discard'],
  }),
  // CephFS：共享文件存储（RWX）
  sc({
    name: 'cephfs-shared', createdAt: daysAgo(200),
    provisioner: 'cephfs.csi.ceph.com',
    parameters: {
      clusterID: 'rook-ceph',
      fsName: 'cephfs',
      'csi.storage.k8s.io/provisioner-secret-name': 'rook-csi-cephfs-provisioner',
      'csi.storage.k8s.io/provisioner-secret-namespace': 'rook-ceph',
    },
    reclaimPolicy: 'Delete', volumeBindingMode: 'Immediate', allowVolumeExpansion: true,
  }),
  // NFS：保留策略
  sc({
    name: 'nfs-retained', createdAt: daysAgo(90),
    provisioner: 'nfs.csi.k8s.io',
    parameters: { server: '192.168.1.100', share: '/exports/k8s' },
    reclaimPolicy: 'Retain', volumeBindingMode: 'Immediate',
    mountOptions: ['nfsvers=4.1', 'hard', 'noatime'],
  }),
  // 云上冷存储：WFFC + Retain
  sc({
    name: 'archive-cold', createdAt: daysAgo(60),
    labels: { tier: 'archive' },
    provisioner: 'ebs.csi.aws.com',
    parameters: { type: 'sc1', fsType: 'xfs', encrypted: 'true' },
    reclaimPolicy: 'Retain', volumeBindingMode: 'WaitForFirstConsumer', allowVolumeExpansion: true,
  }),
]
