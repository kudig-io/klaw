# EtcdGuardian Performance Tuning Guide

## Overview

This guide provides recommendations for optimizing EtcdGuardian performance in production environments.

## Resource Sizing

### Small Clusters (< 50 nodes)

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Medium Clusters (50-200 nodes)

```yaml
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

### Large Clusters (> 200 nodes)

```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

## Backup Performance

### Streaming Upload

For large etcd databases (> 10GB), enable streaming upload:

```yaml
spec:
  backupMode: Full
  concurrency:
    workers: 5
    chunkSize: 10485760  # 10MB chunks
    streamingUpload: true
```

### Compression

Choose compression based on your needs:

- **gzip**: Best compression ratio, slower
- **lz4**: Fast compression, moderate ratio
- **zstd**: Balanced performance

```yaml
spec:
  compression:
    algorithm: zstd
    level: 6
```

### Parallel Processing

Adjust worker count based on available resources:

```yaml
spec:
  concurrency:
    workers: 5  # Adjust based on CPU cores
    parallelStreams: 3
```

## etcd Connection Tuning

### Timeout Settings

```yaml
etcd:
  dialTimeout: 5s
  requestTimeout: 30s
  maxRetries: 3
  retryBackoff: 1s
```

### Connection Pool

```yaml
etcd:
  keepAliveTime: 30s
  keepAliveTimeout: 10s
  maxCallSendMsgSize: 104857600  # 100MB
  maxCallRecvMsgSize: 104857600  # 100MB
```

## Storage Optimization

### S3/OSS Configuration

```yaml
storage:
  uploadTimeout: 300s
  multipartThreshold: 104857600  # 100MB
  multipartChunkSize: 52428800   # 50MB
  maxConcurrentUploads: 5
```

### Network Optimization

- Use regional endpoints to minimize latency
- Enable transfer acceleration for cross-region backups
- Configure appropriate retry policies

## Controller Tuning

### Reconcile Workers

```yaml
controller:
  maxConcurrentReconciles: 5  # Increase for high backup frequency
  reconcileTimeout: 300s
  cacheResyncInterval: 10h
```

### API Server QPS

```yaml
controller:
  apiQPS: 50
  apiBurst: 100
```

## Monitoring Performance

### Key Metrics

- `etcdguardian_backup_duration_seconds`: Backup completion time
- `etcdguardian_backup_size_bytes`: Backup size
- `etcdguardian_backup_total`: Backup count by status
- `etcdguardian_etcd_db_size_bytes`: etcd database size

### Performance Dashboard

Import the provided Grafana dashboard to monitor:
- Backup success rate
- Average backup duration
- Resource utilization
- Queue depth

## Best Practices

### 1. Schedule Optimization

```yaml
spec:
  schedule: "0 2 * * *"  # Off-peak hours
  concurrencyPolicy: Forbid
```

### 2. Resource Quotas

Set appropriate resource quotas for backup jobs:

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: etcdguardian-quota
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
```

### 3. Priority Classes

Assign priority classes to ensure backup operations:

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: etcdguardian-high-priority
value: 1000000
globalDefault: false
description: "High priority for etcd backup operations"
```

### 4. Node Affinity

Prefer nodes with better network/storage performance:

```yaml
affinity:
  nodeAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        preference:
          matchExpressions:
            - key: node-role.kubernetes.io/storage
              operator: Exists
```

## Troubleshooting

### Slow Backups

1. Check network bandwidth to storage
2. Verify etcd performance
3. Adjust chunk size and workers
4. Enable compression

### High Memory Usage

1. Reduce concurrent workers
2. Decrease buffer sizes
3. Enable streaming upload
4. Check for memory leaks

### API Server Throttling

1. Reduce reconcile frequency
2. Lower QPS/Burst settings
3. Implement leader election
4. Use caching effectively

## Performance Testing

Run performance tests before production deployment:

```bash
# Run benchmark
make benchmark

# Profile performance
make profile

# Load testing
kubectl apply -f config/test/load-test.yaml
```
