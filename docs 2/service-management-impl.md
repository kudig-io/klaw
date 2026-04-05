# Service Management Implementation

## Overview

This document describes the implementation of Service management functionality in Klaw, completed on 2026-04-01.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Frontend      │     │   Backend API    │     │   Kubernetes    │
│   (React/TS)    │────▶│   (Go/Gorilla)   │────▶│   API Server    │
└─────────────────┘     └──────────────────┘     └─────────────────┘
       │                         │                         │
       ▼                         ▼                         ▼
ServicesPage.tsx          server.go               Resources
ServiceDetailDrawer       handlers                (client-go)
API Client (api.ts)       resources.go
```

## API Endpoints

### List Services
```
GET /api/clusters/{cluster}/services
GET /api/clusters/{cluster}/namespaces/{namespace}/services
```

**Query Parameters:**
- `cluster` (path): Cluster name
- `namespace` (path, optional): Namespace name (empty for all namespaces)

**Response:** Array of Service objects

### Get Service
```
GET /api/clusters/{cluster}/namespaces/{namespace}/services/{name}
```

**Response:** Service object

### Get Service Endpoints
```
GET /api/clusters/{cluster}/namespaces/{namespace}/services/{name}/endpoints
```

**Response:** Endpoints object with subsets containing addresses and ports

## Data Models

### Service
```typescript
interface Service {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  spec: {
    type: string                    // ClusterIP, NodePort, LoadBalancer, ExternalName
    clusterIP: string
    externalIPs?: string[]
    ports: Array<{
      name?: string
      port: number
      targetPort: number
      protocol: string              // TCP, UDP, SCTP
      nodePort?: number
    }>
    selector?: Record<string, string>
  }
  status: {
    loadBalancer?: {
      ingress?: Array<{
        ip?: string
        hostname?: string
      }>
    }
  }
}
```

### Endpoints
```typescript
interface ServiceEndpoints {
  serviceName: string
  namespace: string
  endpoints: Array<{
    addresses?: Array<{
      ip: string
      hostname?: string
      nodeName?: string
      targetRef?: {
        kind: string
        name: string
        namespace: string
      }
    }>
    notReadyAddresses?: Array<...>
    ports?: Array<{
      name?: string
      port: number
      protocol: string
    }>
  }>
}
```

## Frontend Components

### ServicesPage

**Location:** `web/src/pages/ServicesPage.tsx`

**Features:**
- Service list with cluster/namespace filtering
- "All Namespaces" support using special value `_all`
- Service type badges with color coding
- Port summary display
- Selector labels display
- Delete with confirmation
- Detail drawer trigger

**State Management:**
```typescript
const [services, setServices] = useState<Service[]>([])
const [selectedCluster, setSelectedCluster] = useState('')
const [selectedNamespace, setSelectedNamespace] = useState('')
const [selectedService, setSelectedService] = useState<Service | null>(null)
```

**URL Parameters:**
- `cluster`: Selected cluster name
- `namespace`: Selected namespace (optional)

### ServiceDetailDrawer

**Location:** `web/src/components/ServiceDetailDrawer.tsx`

**Features:**
- Slide-in drawer from right
- Three-tab navigation
- Copy-to-clipboard functionality
- Endpoint health indicators

**Tabs:**

1. **Overview**
   - Basic information grid
   - External IPs list
   - Load Balancer ingress
   - Selector labels
   - Labels
   - Annotations

2. **Ports**
   - Port cards with:
     - Port name
     - Port number
     - Target Port
     - Node Port (if applicable)
     - Protocol badge

3. **Endpoints**
   - Ready addresses with pod references
   - Not ready addresses (highlighted)
   - Endpoint ports
   - Copy IP buttons
   - Refresh button

## Shared Components

### ClusterSelector
**Props:**
```typescript
interface ClusterSelectorProps {
  clusters: string[]
  selected: string
  onSelect: (cluster: string) => void
}
```

### NamespaceSelector
**Props:**
```typescript
interface NamespaceSelectorProps {
  cluster: string
  selected: string
  onSelect: (namespace: string) => void
  showAllNamespaces?: boolean
}
```

**Behavior:**
- Fetches namespaces from API when cluster changes
- Shows "All Namespaces" option when enabled
- Converts `_all` to empty string for API calls

### RefreshButton
**Props:**
```typescript
interface RefreshButtonProps {
  onClick: () => void
  isLoading?: boolean
}
```

## Toast Notification System

**Location:** `web/src/contexts/ToastContext.tsx`

**Usage:**
```typescript
const { showToast } = useToast()
showToast('Service deleted successfully', 'success')
showToast('Failed to load services', 'error')
```

**Features:**
- Auto-dismiss after 3 seconds
- Four types: success, error, warning, info
- Stacked notifications
- Smooth animations

## Backend Implementation

### Resources (internal/kubernetes/resources.go)

```go
func (r *Resources) ListServices(clusterName, namespace string) ([]corev1.Service, error)
func (r *Resources) GetService(clusterName, namespace, serviceName string) (*corev1.Service, error)
func (r *Resources) DeleteService(clusterName, namespace, serviceName string) error
func (r *Resources) GetServiceEndpoints(clusterName, namespace, serviceName string) (*corev1.Endpoints, error)
```

### Handlers (internal/api/server.go)

```go
func (s *Server) handleListAllServices(w http.ResponseWriter, r *http.Request)
func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetService(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetServiceEndpoints(w http.ResponseWriter, r *http.Request)
```

## Service Type Colors

| Type | Background | Text |
|------|-----------|------|
| LoadBalancer | bg-blue-100 | text-blue-800 |
| NodePort | bg-purple-100 | text-purple-800 |
| ClusterIP | bg-green-100 | text-green-800 |
| ExternalName | bg-orange-100 | text-orange-800 |

## Testing

### Manual Testing Commands

```bash
# List all services
curl http://localhost:8080/api/clusters/kind-my-k8s/services

# List services in namespace
curl http://localhost:8080/api/clusters/kind-my-k8s/namespaces/kube-system/services

# Get service details
curl http://localhost:8080/api/clusters/kind-my-k8s/namespaces/kube-system/services/kube-dns

# Get endpoints
curl http://localhost:8080/api/clusters/kind-my-k8s/namespaces/kube-system/services/kube-dns/endpoints
```

### Test Cases

1. **Service List**
   - [ ] Load services for cluster
   - [ ] Filter by namespace
   - [ ] Show "All Namespaces"
   - [ ] Empty state display

2. **Service Detail**
   - [ ] Open drawer
   - [ ] Switch tabs
   - [ ] Copy IP addresses
   - [ ] View endpoints

3. **Delete Service**
   - [ ] Confirmation dialog
   - [ ] Success toast
   - [ ] List refresh
   - [ ] Error handling

## Future Enhancements

1. **Create/Update Service**
   - YAML editor integration
   - Form-based creation
   - Validation

2. **Service Graph**
   - Visual service dependency map
   - Traffic flow visualization

3. **Ingress Integration**
   - Show associated Ingress resources
   - TLS certificate status

4. **Service Mesh**
   - Istio/Linkerd integration
   - Traffic splitting visualization

## Files Changed

### Backend
- `internal/kubernetes/resources.go` - Added Service methods
- `internal/api/server.go` - Added Service handlers and routes

### Frontend
- `web/src/pages/ServicesPage.tsx` - New page
- `web/src/components/ServiceDetailDrawer.tsx` - New component
- `web/src/components/ClusterSelector.tsx` - New component
- `web/src/components/NamespaceSelector.tsx` - New component
- `web/src/components/RefreshButton.tsx` - New component
- `web/src/contexts/ToastContext.tsx` - New context
- `web/src/types/api.ts` - New type definitions
- `web/src/lib/api.ts` - Added serviceApi
- `web/src/App.tsx` - Added route and navigation
- `web/src/main.tsx` - Added ToastProvider

### Documentation
- `CHANGELOG.md` - Created with version history
- `DEVELOPMENT_PLAN.md` - Updated progress
- `docs/service-management-impl.md` - This document
