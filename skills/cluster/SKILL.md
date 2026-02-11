---
title: Cluster Management
description: Manage Kubernetes cluster lifecycle and configuration
author: kudig-io
version: 1.0.0
category: DevOps
requires:
  - go
  - kubectl
  - helm
  - kubernetes/client-go
provides:
  - cluster_management
  - cluster_configuration
  - cluster_monitoring
  - cluster_security
---

# Cluster Management Skill

This skill allows you to manage Kubernetes cluster lifecycle and configuration through OpenClaw. It provides commands to handle cluster operations.

## Available Commands

### Cluster Lifecycle
- `klaw cluster create <cluster-name> <config-file>` - Create a new Kubernetes cluster
- `klaw cluster delete <cluster-name>` - Delete a Kubernetes cluster
- `klaw cluster upgrade <cluster-name> <version>` - Upgrade a Kubernetes cluster

### Cluster Configuration
- `klaw cluster config get <cluster-name>` - Get cluster configuration
- `klaw cluster config set <cluster-name> <key> <value>` - Set cluster configuration
- `klaw cluster config reload <cluster-name>` - Reload cluster configuration

### Cluster Monitoring
- `klaw cluster monitor status <cluster-name>` - Get cluster monitoring status
- `klaw cluster monitor metrics <cluster-name> <metric>` - Get cluster metrics
- `klaw cluster monitor events <cluster-name>` - Get cluster events
- `klaw cluster monitor chart <cluster-name>` - Generate and send monitoring chart
- `klaw cluster monitor alerts <cluster-name>` - Get cluster alerts
- `klaw cluster monitor start <cluster-name>` - Start cluster monitoring
- `klaw cluster monitor stop <cluster-name>` - Stop cluster monitoring

### Cluster Security
- `klaw cluster security audit <cluster-name>` - Run security audit on cluster
- `klaw cluster security policy <cluster-name> <policy>` - Apply security policy
- `klaw cluster security secret <cluster-name> <operation> <secret-name>` - Manage cluster secrets

### Cluster Resources
- `klaw cluster resources list <cluster-name>` - List cluster resources
- `klaw cluster resources usage <cluster-name>` - Get resource usage
- `klaw cluster resources quota <cluster-name> <namespace>` - Get resource quota for namespace

## Example Usage

```bash
# Create a new cluster
klaw cluster create dev-cluster configs/cluster-dev.yaml

# Get cluster status
klaw cluster monitor status dev-cluster

# Run security audit
klaw cluster security audit dev-cluster
```
