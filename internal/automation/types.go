package automation

import "time"

type ScriptType string

const (
	ScriptTypeBuiltin ScriptType = "builtin"
	ScriptTypeCustom  ScriptType = "custom"
)

type ExecutionStatus string

const (
	StatusRunning ExecutionStatus = "running"
	StatusSuccess ExecutionStatus = "success"
	StatusFailed  ExecutionStatus = "failed"
)

type Script struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Type        ScriptType             `json:"type"`
	Script      string                 `json:"script"`
	Schedule    string                 `json:"schedule,omitempty"`
	Timeout     int                    `json:"timeout"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type ScriptExecution struct {
	ID         string                 `json:"id"`
	ScriptID   string                 `json:"scriptId"`
	ScriptName string                 `json:"scriptName"`
	Trigger    string                 `json:"trigger"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Status     ExecutionStatus        `json:"status"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StartTime  time.Time              `json:"startTime"`
	EndTime    *time.Time             `json:"endTime,omitempty"`
	Duration   int                    `json:"duration"`
}

type ScriptFilter struct {
	Type    ScriptType `json:"type,omitempty"`
	Enabled *bool      `json:"enabled,omitempty"`
	Limit   int        `json:"limit,omitempty"`
}

type Statistics struct {
	Total       int            `json:"total"`
	Successful  int            `json:"successful"`
	Failed      int            `json:"failed"`
	ByScript    map[string]int `json:"byScript"`
	ByTrigger   map[string]int `json:"byTrigger"`
	AvgDuration float64        `json:"avgDuration"`
}
