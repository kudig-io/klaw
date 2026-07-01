package etcdbackup

import "time"

// BackupPhase represents the lifecycle phase of an EtcdBackup.
type BackupPhase string

const (
	BackupPhasePending            BackupPhase = "Pending"
	BackupPhaseValidating         BackupPhase = "Validating"
	BackupPhasePreparing          BackupPhase = "Preparing"
	BackupPhaseSnapshotting       BackupPhase = "Snapshotting"
	BackupPhaseUploading          BackupPhase = "Uploading"
	BackupPhaseValidatingSnapshot BackupPhase = "ValidatingSnapshot"
	BackupPhaseTriggeringVelero   BackupPhase = "TriggeringVelero"
	BackupPhaseCompleted          BackupPhase = "Completed"
	BackupPhaseFailed             BackupPhase = "Failed"
)

// RestorePhase represents the lifecycle phase of an EtcdRestore.
type RestorePhase string

const (
	RestorePhasePending     RestorePhase = "Pending"
	RestorePhasePreparing   RestorePhase = "Preparing"
	RestorePhaseDownloading RestorePhase = "Downloading"
	RestorePhaseRestoring   RestorePhase = "Restoring"
	RestorePhaseCompleted   RestorePhase = "Completed"
	RestorePhaseFailed      RestorePhase = "Failed"
)

// StorageProvider defines supported object-storage backends.
type StorageProvider string

const (
	StorageProviderS3    StorageProvider = "S3"
	StorageProviderOSS   StorageProvider = "OSS"
	StorageProviderGCS   StorageProvider = "GCS"
	StorageProviderAzure StorageProvider = "Azure"
)

// StorageLocation holds bucket-level configuration.
type StorageLocation struct {
	Provider          StorageProvider `json:"provider"`
	Bucket            string          `json:"bucket"`
	Prefix            string          `json:"prefix,omitempty"`
	Region            string          `json:"region"`
	Endpoint          string          `json:"endpoint,omitempty"`
	CredentialsSecret string          `json:"credentialsSecret"`
}

// ValidationConfig controls snapshot integrity checks.
type ValidationConfig struct {
	Enabled          bool `json:"enabled"`
	ConsistencyCheck bool `json:"consistencyCheck"`
}

// RetentionPolicy defines cleanup rules.
type RetentionPolicy struct {
	MaxBackups int    `json:"maxBackups,omitempty"`
	MaxAge     string `json:"maxAge,omitempty"`
}

// BackupSpec is the user-facing spec for creating a backup.
type BackupSpec struct {
	BackupMode        BackupMode       `json:"backupMode"`
	EtcdEndpoints     []string         `json:"etcdEndpoints,omitempty"`
	StorageLocation   StorageLocation  `json:"storageLocation"`
	Validation        ValidationConfig `json:"validation"`
	RetentionPolicy   RetentionPolicy  `json:"retentionPolicy,omitempty"`
	VeleroIntegration bool             `json:"veleroIntegration,omitempty"`
}

// BackupMode indicates full or incremental snapshot.
type BackupMode string

const (
	BackupModeFull        BackupMode = "Full"
	BackupModeIncremental BackupMode = "Incremental"
)

// Backup represents a single etcd backup record.
type Backup struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Cluster          string            `json:"cluster"`
	Phase            BackupPhase       `json:"phase"`
	Spec             BackupSpec        `json:"spec"`
	SnapshotSize     int64             `json:"snapshotSize"`
	SnapshotLocation string            `json:"snapshotLocation"`
	EtcdRevision     int64             `json:"etcdRevision"`
	ValidationResult *ValidationResult `json:"validationResult,omitempty"`
	StartTime        time.Time         `json:"startTime"`
	CompletionTime   *time.Time        `json:"completionTime,omitempty"`
	Message          string            `json:"message,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
}

// ValidationResult captures hash/checksum validation output.
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Hash    string `json:"hash,omitempty"`
	Message string `json:"message,omitempty"`
}

// Restore represents an etcd restore operation.
type Restore struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Cluster        string        `json:"cluster"`
	BackupID       string        `json:"backupId"`
	Phase          RestorePhase  `json:"phase"`
	StorageLocation StorageLocation `json:"storageLocation"`
	StartTime      time.Time     `json:"startTime"`
	CompletionTime *time.Time    `json:"completionTime,omitempty"`
	Message        string        `json:"message,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
}

// Schedule defines a recurring backup schedule.
type Schedule struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Cluster         string          `json:"cluster"`
	CronExpression  string          `json:"cronExpression"`
	BackupTemplate  BackupSpec      `json:"backupTemplate"`
	Enabled         bool            `json:"enabled"`
	NextRun         *time.Time      `json:"nextRun,omitempty"`
	LastRun         *time.Time      `json:"lastRun,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// Summary provides aggregated backup statistics.
type Summary struct {
	Total     int            `json:"total"`
	ByPhase   map[string]int `json:"byPhase"`
	ByMode    map[string]int `json:"byMode"`
	Recent24h int            `json:"recent24h"`
}
