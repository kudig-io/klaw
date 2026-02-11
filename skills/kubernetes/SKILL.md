---
title: Kubernetes Management
description: Manage Kubernetes clusters through OpenClaw
author: kudig-io
version: 1.0.0
category: DevOps
requires:
  - go
  - kubectl
  - kubernetes/client-go
provides:
  - cluster_status
  - pod_management
  - deployment_management
  - service_management
---

# Kubernetes Management Skill

This skill allows you to manage Kubernetes clusters through OpenClaw. It provides various commands to interact with your clusters.

## Available Commands

### Cluster Status
- `klaw kubernetes cluster status <cluster-name>` - Get the status of a Kubernetes cluster
- `klaw kubernetes cluster metrics <cluster-name>` - Get detailed metrics for a cluster
- `klaw kubernetes cluster chart <cluster-name>` - Generate and send monitoring chart for a cluster

### Pod Management
- `klaw kubernetes pod list <cluster-name> <namespace>` - List all pods in a namespace
- `klaw kubernetes pod describe <cluster-name> <namespace> <pod-name>` - Describe a specific pod
- `klaw kubernetes pod delete <cluster-name> <namespace> <pod-name>` - Delete a pod
- `klaw kubernetes pod logs <cluster-name> <namespace> <pod-name>` - Get logs from a pod
- `klaw kubernetes pod chart <cluster-name> <namespace> <pod-name>` - Generate monitoring chart for a pod

### Node Management
- `klaw kubernetes node list <cluster-name>` - List all nodes in the cluster
- `klaw kubernetes node describe <cluster-name> <node-name>` - Describe a specific node
- `klaw kubernetes node metrics <cluster-name>` - Get metrics for all nodes
- `klaw kubernetes node chart <cluster-name> <node-name>` - Generate monitoring chart for a node

### Deployment Management
- `klaw kubernetes deployment list <cluster-name> <namespace>` - List all deployments in a namespace
- `klaw kubernetes deployment scale <cluster-name> <namespace> <deployment-name> <replicas>` - Scale a deployment
- `klaw kubernetes deployment rollout <cluster-name> <namespace> <deployment-name>` - Rollout a deployment
- `klaw kubernetes deployment status <cluster-name> <namespace> <deployment-name>` - Get deployment status

### Service Management
- `klaw kubernetes service list <cluster-name> <namespace>` - List all services in a namespace
- `klaw kubernetes service describe <cluster-name> <namespace> <service-name>` - Describe a specific service

### Monitoring
- `klaw kubernetes monitor start <cluster-name>` - Start monitoring for a cluster
- `klaw kubernetes monitor stop <cluster-name>` - Stop monitoring for a cluster
- `klaw kubernetes monitor status <cluster-name>` - Get monitoring status for a cluster
- `klaw kubernetes monitor chart <cluster-name>` - Send current monitoring chart
- `klaw kubernetes monitor alerts <cluster-name>` - Get all alerts for a cluster

### Resource Usage
- `klaw kubernetes resource usage <cluster-name>` - Get resource usage summary
- `klaw kubernetes resource chart <cluster-name>` - Generate resource usage chart

## Example Usage

```bash
# Get cluster status
klaw kubernetes cluster status default

# Get cluster metrics
klaw kubernetes cluster metrics default

# Send monitoring chart
klaw kubernetes cluster chart default

# List pods in default namespace
klaw kubernetes pod list default default

# Get pod logs
klaw kubernetes pod logs default default nginx-pod

# Scale a deployment
klaw kubernetes deployment scale default default nginx-deployment 3

# List nodes
klaw kubernetes node list default

# Get node metrics
klaw kubernetes node metrics default

# Start monitoring
klaw kubernetes monitor start default

# Get monitoring alerts
klaw kubernetes monitor alerts default

# Get resource usage
klaw kubernetes resource usage default
```
