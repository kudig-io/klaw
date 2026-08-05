# Contributing to EtcdGuardian

Thank you for your interest in contributing to EtcdGuardian! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/etcdguardian/etcdguardian/issues)
2. If not, create a new issue with:
   - Clear description of the bug
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details (OS, Kubernetes version, etc.)

### Suggesting Enhancements

1. Check existing [Issues](https://github.com/etcdguardian/etcdguardian/issues) for similar suggestions
2. Create a new issue with:
   - Clear description of the enhancement
   - Use case and benefits
   - Possible implementation approach

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linters
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.22+
- Node.js 20+ (for Web UI)
- Docker
- Kubernetes cluster (for testing)
- Make

### Local Development

```bash
# Clone the repository
git clone https://github.com/etcdguardian/etcdguardian.git
cd etcdguardian

# Install dependencies
make deps

# Run tests
make test

# Run linter
make lint

# Build binaries
make build-all

# Start development environment
make compose-dev
```

### Pre-commit Hooks

Install pre-commit hooks:

```bash
pip install pre-commit
pre-commit install
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Run `gofmt` and `goimports` before committing
- Ensure all tests pass
- Maintain test coverage above 70%
- Document exported functions and types

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Test changes
- `chore:` Build process or auxiliary tool changes

Example:
```
feat(backup): add support for incremental backups

This commit adds support for incremental etcd backups using the watch API.
The implementation includes:
- Watch-based change tracking
- Delta snapshot creation
- Efficient storage utilization

Closes #123
```

### Pull Request Guidelines

1. Keep PRs focused and small
2. Update documentation if needed
3. Add tests for new functionality
4. Ensure CI passes
5. Request review from maintainers

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./pkg/snapshot -v
```

### Integration Tests

```bash
# Run integration tests
make test-integration
```

### E2E Tests

```bash
# Run e2e tests
make test-e2e
```

## Documentation

- Update README.md for user-facing changes
- Update API documentation for CRD changes
- Add inline comments for complex logic
- Update CHANGELOG.md for notable changes

## Release Process

1. Update VERSION file
2. Update CHANGELOG.md
3. Create release PR
4. After merge, tag the release
5. CI/CD will handle the rest

## Getting Help

- Join our [Discussions](https://github.com/etcdguardian/etcdguardian/discussions)
- Ask questions in issues
- Reach out to maintainers

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
