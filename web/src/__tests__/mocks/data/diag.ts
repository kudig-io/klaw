// 集群诊断 issue 库（用于 /diag/run）
// severity 序列化为小写（前端页面 normalizeSeverity 兼容），cn_name 给出中文友好名。

interface IssueSeed {
  severity: 'critical' | 'error' | 'warning' | 'info'
  cn: string
  en: string
  analyzer: string
  location?: string
  details?: string
  remediation?: { suggestion: string; command?: string }
}

const ISSUES: IssueSeed[] = [
  // 故事线核心 issues
  {
    severity: 'critical', cn: '节点内存压力', en: 'Node MemoryPressure', analyzer: 'node-resource',
    location: 'kind-test-worker2',
    details: '内存使用率 93%（7.4Gi/8Gi），kubelet 已进入 MemoryPressure 状态并触发了 Pod 驱逐；该节点上的 Pod 正被重新调度到剩余节点。',
    remediation: { suggestion: '检查节点上内存占用最高的 Pod 并考虑迁移；定位内存泄漏或扩容节点。', command: 'kubectl describe node kind-test-worker2 | grep -A 10 Conditions' },
  },
  {
    severity: 'critical', cn: '容器重启风暴', en: 'Pod Restart Storm', analyzer: 'pod-restart',
    location: 'mall-prod/mall-frontend-7d9c5f8b4-z8x3c',
    details: '容器 mall-frontend 在 30 分钟内重启 8 次；最近一次崩溃日志显示 JavaScript heap out of memory，OOMKilled → CrashLoopBackOff。',
    remediation: { suggestion: '检查 v8 堆大小限制（--max-old-space-size）或应用级内存泄漏；建议先扩容至 2Gi 并加内存监控。', command: 'kubectl logs -p -n mall-prod mall-frontend-7d9c5f8b4-z8x3c --tail 200' },
  },
  {
    severity: 'error', cn: 'Pod 调度失败', en: 'Pod Unschedulable', analyzer: 'scheduler',
    location: 'mall-prod/mall-frontend-7d9c5f8b4-q7w2e',
    details: '0/3 节点可用：2 节点 Insufficient memory（worker2 内存压力），1 节点存在 untolerated taint。',
    remediation: { suggestion: '解除 kind-test-worker2 的内存压力后 Pod 将被自动调度；或添加内存更大的节点。', command: 'kubectl describe pod -n mall-prod mall-frontend-7d9c5f8b4-q7w2e | grep -A 5 Events' },
  },
  {
    severity: 'warning', cn: '副本不可用', en: 'Deployment Replicas Unavailable', analyzer: 'deployment',
    location: 'mall-prod/mall-frontend',
    details: 'Deployment mall-frontend 仅 1/3 副本可用，ProgressDeadlineExceeded；缩容到 3 后无可用副本。',
    remediation: { suggestion: '先修复 Pod 级故障（重启风暴/调度失败），Deployment 进度会自动恢复。', command: 'kubectl rollout status deployment/mall-frontend -n mall-prod' },
  },
  {
    severity: 'warning', cn: '副本不可用', en: 'Deployment Replicas Unavailable', analyzer: 'deployment',
    location: 'data-platform/spark-driver',
    details: 'Deployment spark-driver 0/2 副本可用（节点调度资源不足）；oncall 已 acknowledge。',
    remediation: { suggestion: '临时缩容或扩容节点；考虑使用 priorityClass 调度。', command: 'kubectl scale deployment spark-driver -n data-platform --replicas=1' },
  },
  {
    severity: 'warning', cn: '副本拉取失败', en: 'ImagePullBackOff', analyzer: 'image',
    location: 'mall-staging/cart-service-stg-29',
    details: 'Liveness 探针失败，节点压力导致镜像拉取超时。',
    remediation: { suggestion: '在节点缓解后重启 Pod 即可恢复；或提前在 staging 节点上预拉镜像。', command: 'kubectl rollout restart deployment/cart-service-stg -n mall-staging' },
  },
  // 健康/信息类 issues
  {
    severity: 'info', cn: 'API Server 证书即将过期', en: 'APIServer Certificate Expiring', analyzer: 'cluster',
    location: 'kind-test-control-plane',
    details: 'API Server 客户端证书剩余有效期 45 天，建议在 30 天内轮换。',
    remediation: { suggestion: '通过 cluster-bootstrap 触发证书轮换。', command: 'kubeadm certs renew apiserver' },
  },
  {
    severity: 'info', cn: '节点内核版本一致', en: 'Node Kernel Consistent', analyzer: 'cluster',
    details: '所有节点运行相同内核版本（5.15.0-89-generic），kubelet 版本 v1.29.2，无需升级。',
  },
  {
    severity: 'info', cn: 'etcd 快照备份完成', en: 'etcd Snapshot Backup Completed', analyzer: 'backup',
    location: 'klaw-daily-20260831-030000',
    details: '最近一次全量备份已写入对象存储；当前部分失败次数：1（cluster=kind-test），运维团队已通知。',
    remediation: { suggestion: '检查 etcd snapshot 流超时配置（--snapshot-timeout），或将备份任务拆分到非高峰期。' },
  },
]

// 模拟后端 Severity.MarshalJSON 的小写序列化 + cn_name
export const buildDiagIssues = (nodeFilter?: string) => {
  const issues = ISSUES.filter((i) => !nodeFilter || (i.location && i.location.includes(nodeFilter)))
    .map((i) => ({
      severity: i.severity,
      cn_name: i.cn,
      en_name: i.en,
      analyzer_name: i.analyzer,
      location: i.location,
      details: i.details,
      remediation: i.remediation,
    }))
  return issues
}

export const buildDiagResponse = (nodeFilter?: string) => ({
  data: { nodeName: nodeFilter || '' },
  results: [],
  issues: buildDiagIssues(nodeFilter),
})