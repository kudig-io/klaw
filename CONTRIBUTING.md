# Contributing to Klaw

Thanks for your interest in improving Klaw! All contributions — bug reports, docs, PRs, ideas — are welcome.

## Development setup

### Prerequisites

| Component | Version | Notes |
|---|---|---|
| Go | 1.24+ | `modules/etcd-guardian` requires Go 1.26+ |
| Node.js | 18+ | for the web frontend |
| Kubernetes | 1.24+ | or a local kind cluster |
| Helm | 3.x | for in-cluster deployment |

### Get the code

```bash
git clone https://github.com/kudig-io/klaw.git
cd klaw
```

The repo is a monorepo with five independent Go modules (`./`, `operator/`, `modules/etcd-backup/`, `modules/etcd-guardian/`, `modules/etcd-guardian/backend/`). Each one can be built and tested independently.

### Backend (main app)

```bash
make build          # build frontend + backend
make dev            # run frontend and backend dev servers in parallel
make test           # go test -v ./...
make lint           # golangci-lint + eslint
make fmt            # go fmt + eslint --fix
```

Individual targets are listed by `make help`.

### Frontend

```bash
cd web
npm install
npm run dev          # Vite dev server on port 3000, /api proxied to http://localhost:8080
npm run dev:mock     # MSW mock data, no backend required
npm run test:run     # Vitest
npm run test:coverage
npm run build        # production build to web/dist
npm run lint
```

> `npm run dev:mock` is wired to start the MSW worker in dev mode. If mock data doesn't appear, double-check that `main.tsx` was not modified.

### Other modules

```bash
cd operator && make test
cd modules/etcd-guardian && make test
```

## Commit messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/) so the changelog and release notes can be generated automatically.

Format: `<type>(<scope>): <subject>`

Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`.

Examples:

```
feat(diag): add CIS benchmark analyzer
fix(web): inject Bearer token in API client
docs(readme): add English version
```

Keep the subject line under 72 characters and in the imperative mood ("add", not "added").

## Branch naming

- Feature: `feature/<short-description>`
- Bugfix: `fix/<short-description>`
- Docs: `docs/<short-description>`

## Pull request flow

1. **Fork** this repository.
2. Create a branch: `git checkout -b feature/awesome-thing`.
3. Make your changes. If the change is user-visible (features, behavior changes, breaking changes), update the relevant docs under `docs/` and the README sections that reference it.
4. Ensure local checks pass: `make lint && make test && (cd web && npm run lint && npm run test:run)`.
5. Commit using Conventional Commits.
6. Push and open a Pull Request against `main`.
7. In the PR description, link any related issues and fill in the checklist from `.github/PULL_REQUEST_TEMPLATE.md`.
8. Address review feedback by pushing additional commits (don't force-push during review).

## Reporting bugs

Open an issue using the **Bug Report** template at <https://github.com/kudig-io/klaw/issues/new/choose>. Include:

- Klaw version (`klaw version`)
- Cluster information (Kubernetes version, distribution, in-cluster vs kubeconfig)
- Relevant configuration (sanitize tokens)
- Reproduction steps and expected vs actual behavior
- Logs from `klaw diag` or the server (with sensitive data redacted)

## Suggesting features

Use the **Feature Request** template. Describe the problem you're solving, the proposed solution, and any alternatives you considered. Cross-link related issues or discussions.

## Security issues

Please **do not** open a public GitHub issue for security vulnerabilities. Follow the disclosure process in [SECURITY.md](./SECURITY.md) (GitHub private vulnerability reporting).

## Code of conduct

This project follows the [Contributor Covenant v2.1](./CODE_OF_CONDUCT.md). By participating, you agree to uphold it.

## Repository layout

```
klaw/
├── cmd/klaw/               # CLI entry: main / server / diag
├── internal/               # main app internals (api, diag, kubernetes, ...)
├── web/                    # React frontend
├── operator/               # Kudig Operator (CRDs)
├── modules/
│   ├── etcd-backup/        # etcd backup/restore client library
│   └── etcd-guardian/      # etcd backup/restore Operator + Gin backend API
├── helm/klaw/              # Helm chart
├── deployment/kind/        # kind local cluster scripts
├── configs/                # config.yaml.example
├── docs/                   # design & implementation docs
└── skills/                 # OpenClaw skill definitions
```

## Getting help

- Discussions: <https://github.com/kudig-io/klaw/discussions>
- Issues: <https://github.com/kudig-io/klaw/issues>
- Security disclosures: see [SECURITY.md](./SECURITY.md)