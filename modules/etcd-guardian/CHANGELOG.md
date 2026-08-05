# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Production-ready Dockerfiles for Operator, Backend, and Web UI
- Comprehensive CI/CD pipeline with security scanning
- Kubernetes deployment manifests with RBAC
- Network policies for security isolation
- Performance tuning configuration and documentation
- Pre-commit hooks for code quality
- Docker Compose for local development
- Prometheus and Grafana monitoring setup
- Comprehensive Makefile with production targets
- GoReleaser configuration for automated releases
- Health checks and readiness probes
- Horizontal Pod Autoscaler configuration
- Pod Disruption Budget for high availability

### Changed
- Optimized Makefile with additional production targets
- Enhanced .gitignore for better artifact management
- Improved CI/CD with security scanning and dependency checks

### Fixed
- N/A

### Security
- Added network policies for pod isolation
- Implemented security contexts for all containers
- Added vulnerability scanning in CI/CD
- Configured TLS for secure communication

## [0.1.0] - 2026-01-XX

### Added
- Initial release of EtcdGuardian
- Kubernetes Operator for etcd backup and restore
- Support for full and incremental backups
- Multi-cloud storage support (S3, OSS, GCS, Azure Blob)
- Web UI for backup management
- CLI tool for command-line operations
- Backup scheduling with cron expressions
- Backup validation and consistency checks
- Prometheus metrics export
- Multi-tenancy support with RBAC
- Encryption support with KMS integration

[Unreleased]: https://github.com/etcdguardian/etcdguardian/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/etcdguardian/etcdguardian/releases/tag/v0.1.0
