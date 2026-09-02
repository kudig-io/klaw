// Aggregated MSW handlers (split by domain under ./handlers/)

import { clusterHandlers, nodeHandlers, podHandlers, deploymentHandlers, eventHandlers } from './cluster'
import { monitorHandlers, alertHandlers } from './monitoring'
import { serviceHandlers } from './services'
import { backupHandlers } from './backups'
import { tenancyHandlers, auditHandlers } from './governance'
import { diagHandlers, sosHandlers } from './diag'
import { analysisHandlers } from './analysis'

export const handlers = [
  ...clusterHandlers,
  ...nodeHandlers,
  ...podHandlers,
  ...deploymentHandlers,
  ...eventHandlers,
  ...monitorHandlers,
  ...alertHandlers,
  ...serviceHandlers,
  ...backupHandlers,
  ...tenancyHandlers,
  ...auditHandlers,
  ...diagHandlers,
  ...sosHandlers,
  ...analysisHandlers,
]

export { analysisHandlers, clusterHandlers, nodeHandlers, podHandlers, deploymentHandlers, eventHandlers, monitorHandlers, alertHandlers, serviceHandlers, backupHandlers, tenancyHandlers, auditHandlers, diagHandlers, sosHandlers }