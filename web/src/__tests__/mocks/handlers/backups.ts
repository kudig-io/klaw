// 备份 handlers

import { http, HttpResponse } from 'msw'
import { derive } from '../data/index'
import type { BackupItem } from '../../../lib/api'
import { store, appendAudit, nextBackupName, now } from '../store'

export const backupHandlers = [
  http.get('/api/v1/clusters/:cluster/backups', ({ params }) =>
    HttpResponse.json(store.backups.filter((b) => b.cluster === params.cluster))),
  http.get('/api/v1/clusters/:cluster/backups/summary', ({ params }) =>
    HttpResponse.json(derive.backupSummary(params.cluster as string))),
  http.get('/api/v1/clusters/:cluster/backups/:name', ({ params }) => {
    const b = store.backups.find((b) => b.cluster === params.cluster && b.name === params.name)
    return b ? HttpResponse.json(b) : new HttpResponse(null, { status: 404 })
  }),
  http.post('/api/v1/clusters/:cluster/backups', async ({ params, request }) => {
    const body = await request.json() as any
    const revision = 12868_990 + store.backups.length * 47
    const backup: BackupItem = {
      name: body.name || nextBackupName(),
      cluster: params.cluster as string,
      phase: 'Completed',
      spec: {
        backupMode: body.backupMode || 'Full',
        etcdEndpoints: body.etcdEndpoints || [],
        storageLocation: body.storageLocation || {
          provider: 'S3', bucket: 'klaw-backups', prefix: 'etcd/' + params.cluster,
          region: 'cn-north-1', credentialsSecret: 'klaw-aws-credentials',
        },
        validation: body.validation || { enabled: true, consistencyCheck: true },
      },
      snapshotSize: 2_200_000_000 + Math.floor(Math.random() * 100_000_000),
      snapshotLocation: `s3://klaw-backups/etcd/${params.cluster}/${body.name || nextBackupName()}.db`,
      etcdRevision: revision,
      validationResult: { valid: true, hash: Math.random().toString(16).slice(2, 14), message: 'sha256 verified' },
      startTime: now(),
      completionTime: now(),
      createdAt: now(),
    }
    store.backups.unshift(backup)
    appendAudit({
      eventType: 'backup.create', category: 'backup', severity: 'info', user: 'oncall',
      action: `create backup ${backup.name}`, resource: { kind: 'Backup', name: backup.name },
      cluster: params.cluster as string, result: 'success',
    })
    return HttpResponse.json(backup, { status: 201 })
  }),
  http.delete('/api/v1/clusters/:cluster/backups/:name', ({ params }) => {
    const idx = store.backups.findIndex((b) => b.cluster === params.cluster && b.name === params.name)
    if (idx < 0) return new HttpResponse(null, { status: 404 })
    const [removed] = store.backups.splice(idx, 1)
    appendAudit({
      eventType: 'backup.delete', category: 'backup', severity: 'info', user: 'oncall',
      action: `delete backup ${removed.name}`, resource: { kind: 'Backup', name: removed.name },
      cluster: params.cluster as string, result: 'success',
    })
    return HttpResponse.json({ message: `Backup ${removed.name} deleted successfully` })
  }),
]