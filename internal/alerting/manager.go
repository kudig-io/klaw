package alerting

import (
	"fmt"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/storage"
)

type Condition struct {
	Type       string      `json:"type"`
	Field      string      `json:"field"`
	Operator   string      `json:"operator"`
	Threshold  interface{} `json:"threshold"`
	TimeWindow string      `json:"timeWindow,omitempty"`
}

type Rule struct {
	ID          string    `json:"id"`
	Cluster     string    `json:"cluster,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	Severity    string    `json:"severity"`
	Condition   Condition `json:"condition"`
	Actions     []string  `json:"actions,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Record struct {
	ID            string      `json:"id"`
	Cluster       string      `json:"cluster"`
	RuleID        string      `json:"ruleId"`
	RuleName      string      `json:"ruleName"`
	RuleType      string      `json:"ruleType"`
	ResourceKind  string      `json:"resourceKind"`
	ResourceName  string      `json:"resourceName"`
	Namespace     string      `json:"namespace,omitempty"`
	Severity      string      `json:"severity"`
	Value         interface{} `json:"value"`
	Threshold     interface{} `json:"threshold"`
	Operator      string      `json:"operator"`
	Message       string      `json:"message"`
	Acknowledged  bool        `json:"acknowledged"`
	Resolved      bool        `json:"resolved"`
	CreatedAt     time.Time   `json:"createdAt"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	ResolvedAt    *time.Time  `json:"resolvedAt,omitempty"`
}

type Stats struct {
	Total      int            `json:"total"`
	Active     int            `json:"active"`
	BySeverity map[string]int `json:"bySeverity"`
	ByStatus   map[string]int `json:"byStatus"`
	Recent24h  int            `json:"recent24h"`
}

type Manager struct {
	resources    *kubernetes.Resources
	store        *storage.Store
	rules        []Rule
	history      []Record
	activeAlerts map[string]Record
	mu           sync.RWMutex
}

func NewManager(resources *kubernetes.Resources, store *storage.Store) *Manager {
	m := &Manager{
		resources:    resources,
		store:        store,
		activeAlerts: make(map[string]Record),
	}
	m.load()
	return m
}

func (m *Manager) GetRules(cluster string) []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Rule
	for _, rule := range m.rules {
		if rule.Cluster == "" || rule.Cluster == cluster {
			result = append(result, rule)
		}
	}
	return result
}

func (m *Manager) AddRule(rule Rule) (Rule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("alert-%d", now.UnixNano())
	}
	rule.Enabled = rule.Enabled || !rule.Enabled && rule.CreatedAt.IsZero()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	m.rules = append(m.rules, rule)
	if err := m.saveRulesLocked(); err != nil {
		return Rule{}, err
	}
	return rule, nil
}

func (m *Manager) UpdateRule(ruleID string, updates Rule) (Rule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, rule := range m.rules {
		if rule.ID != ruleID {
			continue
		}
		updates.ID = ruleID
		updates.CreatedAt = rule.CreatedAt
		updates.UpdatedAt = time.Now()
		m.rules[i] = updates
		if err := m.saveRulesLocked(); err != nil {
			return Rule{}, err
		}
		return updates, nil
	}
	return Rule{}, fmt.Errorf("alert rule not found: %s", ruleID)
}

func (m *Manager) DeleteRule(ruleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, rule := range m.rules {
		if rule.ID != ruleID {
			continue
		}
		m.rules = append(m.rules[:i], m.rules[i+1:]...)
		return m.saveRulesLocked()
	}
	return fmt.Errorf("alert rule not found: %s", ruleID)
}

func (m *Manager) GetHistory(cluster string, limit int) []Record {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Record
	for _, record := range m.history {
		if cluster == "" || record.Cluster == cluster {
			result = append(result, record)
		}
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func (m *Manager) GetStats(cluster string) Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		BySeverity: map[string]int{"critical": 0, "error": 0, "warning": 0, "info": 0},
		ByStatus:   map[string]int{"acknowledged": 0, "resolved": 0, "pending": 0},
	}
	dayAgo := time.Now().Add(-24 * time.Hour)

	for _, record := range m.history {
		if cluster != "" && record.Cluster != cluster {
			continue
		}

		stats.Total++
		stats.BySeverity[record.Severity]++
		if record.Resolved {
			stats.ByStatus["resolved"]++
		} else if record.Acknowledged {
			stats.ByStatus["acknowledged"]++
		} else {
			stats.ByStatus["pending"]++
		}
		if record.CreatedAt.After(dayAgo) {
			stats.Recent24h++
		}
	}

	for _, record := range m.activeAlerts {
		if cluster == "" || record.Cluster == cluster {
			stats.Active++
		}
	}

	return stats
}

func (m *Manager) Acknowledge(recordID string) (*Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, record := range m.history {
		if record.ID != recordID {
			continue
		}
		now := time.Now()
		record.Acknowledged = true
		record.AcknowledgedAt = &now
		m.history[i] = record
		if active, ok := m.activeAlerts[m.activeKey(record)]; ok {
			active.Acknowledged = true
			active.AcknowledgedAt = &now
			m.activeAlerts[m.activeKey(record)] = active
		}
		if err := m.saveHistoryLocked(); err != nil {
			return nil, err
		}
		return &record, nil
	}
	return nil, fmt.Errorf("alert not found: %s", recordID)
}

func (m *Manager) Resolve(recordID string) (*Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, record := range m.history {
		if record.ID != recordID {
			continue
		}
		now := time.Now()
		record.Resolved = true
		record.ResolvedAt = &now
		m.history[i] = record
		delete(m.activeAlerts, m.activeKey(record))
		if err := m.saveHistoryLocked(); err != nil {
			return nil, err
		}
		return &record, nil
	}
	return nil, fmt.Errorf("alert not found: %s", recordID)
}

func (m *Manager) EvaluateCluster(cluster string) ([]Record, error) {
	rules := m.GetRules(cluster)
	if len(rules) == 0 {
		return nil, nil
	}

	pods, err := m.resources.ListPods(cluster, "")
	if err != nil {
		return nil, err
	}
	nodes, err := m.resources.ListNodes(cluster)
	if err != nil {
		return nil, err
	}
	events, err := m.resources.ListEvents(cluster, "")
	if err != nil {
		return nil, err
	}

	var triggered []Record
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		switch strings.ToLower(rule.Condition.Type) {
		case "pod":
			for _, pod := range pods {
				if record, ok := m.evaluatePodRule(cluster, rule, pod); ok {
					if saved, created, err := m.storeRecord(record); err == nil && created {
						triggered = append(triggered, saved)
					}
				}
			}
		case "node":
			for _, node := range nodes {
				if record, ok := m.evaluateNodeRule(cluster, rule, node); ok {
					if saved, created, err := m.storeRecord(record); err == nil && created {
						triggered = append(triggered, saved)
					}
				}
			}
		case "event":
			for _, event := range events {
				if record, ok := m.evaluateEventRule(cluster, rule, event); ok {
					if saved, created, err := m.storeRecord(record); err == nil && created {
						triggered = append(triggered, saved)
					}
				}
			}
		}
	}

	return triggered, nil
}

func (m *Manager) evaluatePodRule(cluster string, rule Rule, pod corev1.Pod) (Record, bool) {
	var value interface{}
	switch rule.Condition.Field {
	case "restartCount":
		total := int32(0)
		for _, status := range pod.Status.ContainerStatuses {
			total += status.RestartCount
		}
		value = int(total)
	case "phase":
		value = string(pod.Status.Phase)
	case "age":
		value = int(time.Since(pod.CreationTimestamp.Time).Seconds())
	default:
		return Record{}, false
	}

	if !compareValues(value, rule.Condition.Operator, rule.Condition.Threshold) {
		return Record{}, false
	}

	return Record{
		ID:           fmt.Sprintf("alert-record-%d", time.Now().UnixNano()),
		Cluster:      cluster,
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		RuleType:     rule.Condition.Type,
		ResourceKind: "Pod",
		ResourceName: pod.Name,
		Namespace:    pod.Namespace,
		Severity:     rule.Severity,
		Value:        value,
		Threshold:    rule.Condition.Threshold,
		Operator:     rule.Condition.Operator,
		Message:      fmt.Sprintf("%s triggered for pod %s", rule.Name, pod.Name),
		CreatedAt:    time.Now(),
	}, true
}

func (m *Manager) evaluateNodeRule(cluster string, rule Rule, node corev1.Node) (Record, bool) {
	var value interface{}
	switch rule.Condition.Field {
	case "ready":
		value = isNodeConditionTrue(node.Status.Conditions, corev1.NodeReady)
	case "diskPressure":
		value = isNodeConditionTrue(node.Status.Conditions, corev1.NodeDiskPressure)
	case "memoryPressure":
		value = isNodeConditionTrue(node.Status.Conditions, corev1.NodeMemoryPressure)
	default:
		return Record{}, false
	}

	if !compareValues(value, rule.Condition.Operator, rule.Condition.Threshold) {
		return Record{}, false
	}

	return Record{
		ID:           fmt.Sprintf("alert-record-%d", time.Now().UnixNano()),
		Cluster:      cluster,
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		RuleType:     rule.Condition.Type,
		ResourceKind: "Node",
		ResourceName: node.Name,
		Severity:     rule.Severity,
		Value:        value,
		Threshold:    rule.Condition.Threshold,
		Operator:     rule.Condition.Operator,
		Message:      fmt.Sprintf("%s triggered for node %s", rule.Name, node.Name),
		CreatedAt:    time.Now(),
	}, true
}

func (m *Manager) evaluateEventRule(cluster string, rule Rule, event corev1.Event) (Record, bool) {
	var value interface{}
	switch rule.Condition.Field {
	case "type":
		value = event.Type
	case "reason":
		value = event.Reason
	case "message":
		value = event.Message
	default:
		return Record{}, false
	}

	if !compareValues(value, rule.Condition.Operator, rule.Condition.Threshold) {
		return Record{}, false
	}

	return Record{
		ID:           fmt.Sprintf("alert-record-%d", time.Now().UnixNano()),
		Cluster:      cluster,
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		RuleType:     rule.Condition.Type,
		ResourceKind: "Event",
		ResourceName: event.Name,
		Namespace:    event.Namespace,
		Severity:     rule.Severity,
		Value:        value,
		Threshold:    rule.Condition.Threshold,
		Operator:     rule.Condition.Operator,
		Message:      fmt.Sprintf("%s triggered for event %s", rule.Name, event.Name),
		CreatedAt:    time.Now(),
	}, true
}

func (m *Manager) storeRecord(record Record) (Record, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.activeKey(record)
	if _, exists := m.activeAlerts[key]; exists {
		return Record{}, false, nil
	}

	m.activeAlerts[key] = record
	m.history = append([]Record{record}, m.history...)
	if len(m.history) > 1000 {
		m.history = m.history[:1000]
	}

	if err := m.saveHistoryLocked(); err != nil {
		return Record{}, false, err
	}
	return record, true, nil
}

func (m *Manager) activeKey(record Record) string {
	return fmt.Sprintf("%s:%s:%s:%s:%v", record.Cluster, record.RuleID, record.ResourceKind, record.ResourceName, record.Value)
}

func (m *Manager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, _ = m.store.GetJSON("alerting", "rules", &m.rules)
	if len(m.rules) == 0 {
		m.rules = defaultRules()
		_ = m.saveRulesLocked()
	}

	_, _ = m.store.GetJSON("alerting", "history", &m.history)
	for _, record := range m.history {
		if !record.Resolved {
			m.activeAlerts[m.activeKey(record)] = record
		}
	}
}

func (m *Manager) saveRulesLocked() error {
	return m.store.PutJSON("alerting", "rules", m.rules)
}

func (m *Manager) saveHistoryLocked() error {
	return m.store.PutJSON("alerting", "history", m.history)
}

func defaultRules() []Rule {
	now := time.Now()
	return []Rule{
		{
			ID:          "pod-crash-looping",
			Name:        "Pod 崩溃循环",
			Description: "检测 Pod 重启次数是否过高",
			Enabled:     true,
			Severity:    "critical",
			Condition:   Condition{Type: "pod", Field: "restartCount", Operator: ">", Threshold: 5, TimeWindow: "5m"},
			Actions:     []string{"log", "notify"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "node-not-ready",
			Name:        "节点不可用",
			Description: "检测节点是否不可用",
			Enabled:     true,
			Severity:    "critical",
			Condition:   Condition{Type: "node", Field: "ready", Operator: "==", Threshold: false, TimeWindow: "1m"},
			Actions:     []string{"log", "notify"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "failed-events",
			Name:        "失败事件",
			Description: "检测 Warning 类型事件",
			Enabled:     true,
			Severity:    "warning",
			Condition:   Condition{Type: "event", Field: "type", Operator: "==", Threshold: "Warning", TimeWindow: "5m"},
			Actions:     []string{"log"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

func isNodeConditionTrue(conditions []corev1.NodeCondition, target corev1.NodeConditionType) bool {
	for _, condition := range conditions {
		if condition.Type == target {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func compareValues(value interface{}, operator string, threshold interface{}) bool {
	switch v := value.(type) {
	case int:
		t, ok := toFloat64(threshold)
		if !ok {
			return false
		}
		return compareFloat(float64(v), operator, t)
	case bool:
		tb, ok := threshold.(bool)
		if !ok {
			return false
		}
		switch operator {
		case "==":
			return v == tb
		case "!=":
			return v != tb
		default:
			return false
		}
	case string:
		ts, ok := threshold.(string)
		if !ok {
			return false
		}
		switch operator {
		case "==":
			return v == ts
		case "!=":
			return v != ts
		case "contains":
			return strings.Contains(v, ts)
		case "not-contains":
			return !strings.Contains(v, ts)
		default:
			return false
		}
	default:
		return false
	}
}

func compareFloat(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}
