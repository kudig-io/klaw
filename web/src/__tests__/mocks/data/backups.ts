// 备份：日常链 + 增量 + 一次最近的部分失败（故事线）

import { daysAgo, hoursAgo } from '../time'

// 输出形状对齐 BackupItem；不再使用独立 seed 接口以避免重命名混乱
interface OutBackup {
  name: string
  cluster: string
  phase: string
  spec: {
    backupMode: 'Full' | 'Incremental'
    etcdEndpoints?: string[]
    storageLocation: {
      provider: string
      bucket: string
      prefix?: string
      region: string
      endpoint?: string
      credentialsSecret: string
    }
    validation?: { enabled: boolean; consistencyCheck: boolean }
  }
  snapshotSize: number
  snapshotLocation: string
  etcdRevision: number
  message?: string
  validationResult?: { valid: boolean; hash?: string; message?: string }
  startTime: string
  completionTime?: string
  createdAt: string
}

const b = (s: {
  name: string
  cluster: string
  phase: string
  mode: 'Full' | 'Incremental'
  size: number
  revision: number
  startedAt: string
  completedAt?: string
  createdAt: string
  message?: string
  validationResult?: { valid: boolean; hash?: string; message?: string }
}): OutBackup => ({
  name: s.name,
  cluster: s.cluster,
  phase: s.phase,
  spec: {
    backupMode: s.mode,
    etcdEndpoints: ['https://etcd-0.kind-test-control-plane:2379', 'https://etcd-1.kind-test-control-plane:2379', 'https://etcd-2.kind-test-control-plane:2379'],
    storageLocation: {
      provider: 'S3' as string,
      bucket: 'klaw-backups',
      prefix: 'etcd/kind-test',
      region: 'cn-north-1',
      endpoint: 's3.cn-north-1.amazonaws.com.cn',
      credentialsSecret: 'klaw-aws-credentials',
    },
    validation: { enabled: true, consistencyCheck: true },
  },
  snapshotSize: s.size,
  snapshotLocation: `s3://klaw-backups/etcd/kind-test/${s.name}.db`,
  etcdRevision: s.revision,
  message: s.message,
  validationResult: s.validationResult,
  startTime: s.startedAt,
  completionTime: s.completedAt,
  createdAt: s.createdAt,
})

export const mockBackups: OutBackup[] = [
  // kind-test 日常链
  b({ name: 'klaw-daily-20260824-030000', cluster: 'kind-test', phase: 'Completed', mode: 'Full', size: 2_147_483_648, revision: 12845_220, validationResult: { valid: true, hash: 'a3f8e1c4b6d2', message: 'sha256 verified' }, startedAt: daysAgo(7), completedAt: daysAgo(7), createdAt: daysAgo(7) }),
  b({ name: 'klaw-daily-20260825-030000', cluster: 'kind-test', phase: 'Completed', mode: 'Full', size: 2_156_237_568, revision: 12845_220, validationResult: { valid: true, hash: 'b4e9c2d5a3f7', message: 'sha256 verified' }, startedAt: daysAgo(6), completedAt: daysAgo(6), createdAt: daysAgo(6) }),
  b({ name: 'klaw-daily-20260826-030000', cluster: 'kind-test', phase: 'Completed', mode: 'Full', size: 2_169_456_960, revision: 12849_120, validationResult: { valid: true, hash: 'c5fa1b8e6d4a', message: 'sha256 verified' }, startedAt: daysAgo(5), completedAt: daysAgo(5), createdAt: daysAgo(5) }),
  b({ name: 'klaw-daily-20260827-030000', cluster: 'kind-test', phase: 'Failed', mode: 'Full', size: 0, revision: 12849_120, validationResult: { valid: false, message: 'checksum mismatch; abort' }, startedAt: daysAgo(4), completedAt: daysAgo(4), createdAt: daysAgo(4), message: 'etcd quota exceeded during snapshot upload' }),
  b({ name: 'klaw-daily-20260828-030000', cluster: 'kind-test', phase: 'Completed', mode: 'Full', size: 2_187_354_112, revision: 12854_320, validationResult: { valid: true, hash: 'd6ab2c9f7e5b', message: 'sha256 verified' }, startedAt: daysAgo(3), completedAt: daysAgo(3), createdAt: daysAgo(3) }),
  b({ name: 'klaw-daily-20260829-030000', cluster: 'kind-test', phase: 'Completed', mode: 'Full', size: 2_198_765_440, revision: 12858_780, validationResult: { valid: true, hash: 'e7bc3d0a8f6c', message: 'sha256 verified' }, startedAt: daysAgo(2), completedAt: daysAgo(2), createdAt: daysAgo(2) }),
  b({ name: 'klaw-incremental-20260829-120000', cluster: 'kind-test', phase: 'Completed', mode: 'Incremental', size: 142_606_336, revision: 12858_780, validationResult: { valid: true, message: 'incremental verified' }, startedAt: daysAgo(2), completedAt: daysAgo(2), createdAt: daysAgo(2) }),
  // 故事线部分失败 + 最近
  b({ name: 'klaw-daily-20260830-030000', cluster: 'kind-test', phase: 'PartiallyFailed', mode: 'Full', size: 1_834_552_320, revision: 12863_510, validationResult: { valid: false, message: 'snapshot uploaded but 3 keys missing in revision range 12863_000-12863_510' }, startedAt: daysAgo(1), completedAt: daysAgo(1), createdAt: daysAgo(1), message: 'etcd snapshot write stalled after 22s; retry exhausted' }),
  b({ name: 'klaw-daily-20260831-030000', cluster: 'kind-test', phase: 'PartiallyFailed', mode: 'Full', size: 1_890_310_144, revision: 12868_990, validationResult: { valid: false, message: 'snapshot stream timeout at 30s; 2 watcher connections dropped' }, startedAt: hoursAgo(2), completedAt: hoursAgo(2), createdAt: hoursAgo(2), message: 'partial: 47.6% of in-memory revisions flushed before context deadline' }),
  // production（备用）
  b({ name: 'prod-daily-20260830-030000', cluster: 'production', phase: 'Completed', mode: 'Full', size: 5_368_709_120, revision: 9865_120, validationResult: { valid: true, hash: 'f8cd4e1b9f7d', message: 'sha256 verified' }, startedAt: daysAgo(1), completedAt: daysAgo(1), createdAt: daysAgo(1) }),
  b({ name: 'prod-daily-20260831-030000', cluster: 'production', phase: 'Completed', mode: 'Full', size: 5_391_599_104, revision: 9872_440, validationResult: { valid: true, hash: 'a9de5f2c0a8e', message: 'sha256 verified' }, startedAt: hoursAgo(2), completedAt: hoursAgo(2), createdAt: hoursAgo(2) }),
]

// 可变副本：handlers 中的 create/delete 会改这里
export const backupsStore = mockBackups.map((b_) => structuredClone(b_))