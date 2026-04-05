// API 集成测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { server } from '../mocks/server'
import api, { 
  clusterApi, 
  podApi, 
  nodeApi, 
  deploymentApi,
  eventApi,
  monitoringApi 
} from '../../lib/api'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

describe('API Integration Tests', () => {
  describe('Cluster API', () => {
    it('should fetch clusters list', async () => {
      const response = await clusterApi.getClusters()
      expect(response.data).toHaveLength(2)
      expect(response.data[0].name).toBe('kind-test')
    })

    it('should fetch single cluster', async () => {
      const response = await clusterApi.getCluster('kind-test')
      expect(response.data.name).toBe('kind-test')
    })

    it('should fetch cluster status', async () => {
      const response = await clusterApi.getClusterStatus('kind-test')
      expect(response.data.cluster).toBe('kind-test')
      expect(response.data.nodes.total).toBe(3)
      expect(response.data.pods.running).toBe(10)
    })

    it('should fetch namespaces', async () => {
      const response = await clusterApi.getNamespaces('kind-test')
      expect(response.data.length).toBeGreaterThan(0)
      expect(response.data.some((ns: any) => ns.metadata.name === 'default')).toBe(true)
    })
  })

  describe('Pod API', () => {
    it('should fetch pods', async () => {
      const response = await podApi.listPods('kind-test', 'klaw-test')
      expect(response.data.length).toBeGreaterThan(0)
      expect(response.data[0].metadata.namespace).toBe('klaw-test')
    })

    it('should fetch pod logs', async () => {
      const response = await podApi.getPodLogs('kind-test', 'klaw-test', 'nginx-abc', 100)
      expect(response.data.logs).toContain('[mock]')
    })

    it('should delete pod', async () => {
      const response = await podApi.deletePod('kind-test', 'klaw-test', 'test-pod')
      expect(response.data.message).toContain('deleted successfully')
    })
  })

  describe('Node API', () => {
    it('should fetch nodes', async () => {
      const response = await nodeApi.listNodes('kind-test')
      expect(response.data).toHaveLength(3)
      expect(response.data[0].metadata.name).toContain('kind-test')
    })

    it('should fetch node metrics', async () => {
      const response = await nodeApi.getNodeMetrics('kind-test')
      expect(response.data).toHaveProperty('kind-test-control-plane')
      expect(response.data['kind-test-control-plane'].cpu).toBe('4')
    })
  })

  describe('Deployment API', () => {
    it('should fetch deployments', async () => {
      const response = await deploymentApi.listDeployments('kind-test', 'klaw-test')
      expect(response.data.length).toBeGreaterThan(0)
      expect(response.data.some((d: any) => d.metadata.name === 'nginx')).toBe(true)
    })

    it('should fetch deployment status', async () => {
      const response = await deploymentApi.getDeploymentStatus('kind-test', 'klaw-test', 'nginx')
      expect(response.data.name).toBe('nginx')
      expect(response.data).toHaveProperty('availableReplicas')
      expect(response.data).toHaveProperty('conditions')
    })

    it('should scale deployment', async () => {
      const response = await deploymentApi.scaleDeployment('kind-test', 'klaw-test', 'nginx', 5)
      expect(response.data.message).toContain('scaled successfully')
      expect(response.data.replicas).toBe(5)
    })

    it('should restart deployment', async () => {
      const response = await deploymentApi.restartDeployment('kind-test', 'klaw-test', 'nginx')
      expect(response.data.message).toContain('restarted successfully')
    })

    it('should fetch deployment pods', async () => {
      const response = await deploymentApi.getDeploymentPods('kind-test', 'klaw-test', 'nginx')
      expect(Array.isArray(response.data)).toBe(true)
    })
  })

  describe('Event API', () => {
    it('should fetch events', async () => {
      const response = await eventApi.getEvents('kind-test', 'klaw-test')
      expect(Array.isArray(response.data)).toBe(true)
    })
  })

  describe('Monitoring API', () => {
    it('should fetch monitoring status', async () => {
      const response = await monitoringApi.getStatus('kind-test')
      expect(response.data.active).toBe(true)
      expect(response.data.cluster).toBe('kind-test')
    })

    it('should fetch alerts', async () => {
      const response = await monitoringApi.getAlerts('kind-test')
      expect(Array.isArray(response.data)).toBe(true)
    })

    it('should fetch metrics history', async () => {
      const response = await monitoringApi.getHistory('kind-test')
      expect(response.data.length).toBeGreaterThan(0)
      expect(response.data[0]).toHaveProperty('nodes')
      expect(response.data[0]).toHaveProperty('pods')
    })
  })
})
