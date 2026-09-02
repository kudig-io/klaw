// 集群与命名空间

import { daysAgo } from '../time'

export const mockClusters = [
  {
    name: 'kind-test',
    kubeconfig: '/Users/test/.kube/config',
    context: 'kind-test',
  },
  {
    name: 'production',
    kubeconfig: '/Users/test/.kube/prod',
    context: 'prod-cluster',
  },
]

// kind-test 承载完整业务拓扑；production 仅做概览展示
export const mockNamespaces = [
  { metadata: { name: 'default', creationTimestamp: daysAgo(240) } },
  { metadata: { name: 'kube-system', creationTimestamp: daysAgo(240) } },
  { metadata: { name: 'ingress-nginx', creationTimestamp: daysAgo(180) } },
  { metadata: { name: 'klaw-test', creationTimestamp: daysAgo(60) } },
  { metadata: { name: 'mall-prod', creationTimestamp: daysAgo(45) } },
  { metadata: { name: 'mall-staging', creationTimestamp: daysAgo(45) } },
  { metadata: { name: 'data-platform', creationTimestamp: daysAgo(30) } },
]
