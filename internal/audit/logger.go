package audit

import (
	"fmt"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/storage"
)

type AuditEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"eventType"`
	Category  string                 `json:"category"`
	Severity  string                 `json:"severity"`
	Source    string                 `json:"source"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Resource  map[string]string      `json:"resource"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ipAddress,omitempty"`
	UserAgent string                 `json:"userAgent,omitempty"`
}

type AuditFilter struct {
	EventType string
	Category  string
	Severity  string
	User      string
	Limit     int
}

type Statistics struct {
	TotalLogs   int            `json:"totalLogs"`
	ByEventType map[string]int `json:"byEventType"`
	BySeverity  map[string]int `json:"bySeverity"`
	ByCategory  map[string]int `json:"byCategory"`
	ByUser      map[string]int `json:"byUser"`
	Recent24h   int            `json:"recent24h"`
}

type Logger struct {
	store             *storage.Store
	logs              []AuditEvent
	securityEvents    []SecurityEvent
	complianceReports []ComplianceReport
	mu                sync.RWMutex
}

func NewLogger(store *storage.Store) *Logger {
	l := &Logger{store: store}
	l.load()
	l.loadSecurity()
	l.loadCompliance()
	return l
}

func (l *Logger) Log(event AuditEvent) AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Source == "" {
		event.Source = "klaw"
	}
	if event.User == "" {
		event.User = "system"
	}
	if event.Result == "" {
		event.Result = "success"
	}
	l.logs = append([]AuditEvent{event}, l.logs...)
	if len(l.logs) > 10000 {
		l.logs = l.logs[:10000]
	}
	_ = l.saveLocked()
	return event
}

func (l *Logger) List(filter AuditFilter) []AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []AuditEvent
	for _, event := range l.logs {
		if filter.EventType != "" && event.EventType != filter.EventType {
			continue
		}
		if filter.Category != "" && event.Category != filter.Category {
			continue
		}
		if filter.Severity != "" && event.Severity != filter.Severity {
			continue
		}
		if filter.User != "" && event.User != filter.User {
			continue
		}
		result = append(result, event)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result
}

func (l *Logger) Statistics() Statistics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := Statistics{
		TotalLogs:   len(l.logs),
		ByEventType: map[string]int{},
		BySeverity:  map[string]int{},
		ByCategory:  map[string]int{},
		ByUser:      map[string]int{},
	}
	dayAgo := time.Now().Add(-24 * time.Hour)
	for _, log := range l.logs {
		stats.ByEventType[log.EventType]++
		stats.BySeverity[log.Severity]++
		stats.ByCategory[log.Category]++
		stats.ByUser[log.User]++
		if log.Timestamp.After(dayAgo) {
			stats.Recent24h++
		}
	}
	return stats
}

func (l *Logger) load() {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.store.GetJSON("audit", "logs", &l.logs)
}

func (l *Logger) saveLocked() error {
	return l.store.PutJSON("audit", "logs", l.logs)
}

func (l *Logger) loadSecurity() {
	_, _ = l.store.GetJSON("audit", "security_events", &l.securityEvents)
}

func (l *Logger) saveSecurityLocked() error {
	return l.store.PutJSON("audit", "security_events", l.securityEvents)
}

func (l *Logger) loadCompliance() {
	_, _ = l.store.GetJSON("audit", "compliance_reports", &l.complianceReports)
}

func (l *Logger) saveComplianceLocked() error {
	return l.store.PutJSON("audit", "compliance_reports", l.complianceReports)
}
