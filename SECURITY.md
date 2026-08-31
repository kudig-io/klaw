# Security Policy

## Supported Versions

Klaw is under active early-stage development. Only the latest commit on `main` receives security fixes.

| Version / Branch | Supported |
|---|---|
| `main` (latest) | ✅ |
| Older commits, tags, releases | ❌ |

We strongly recommend running Klaw pinned to a specific commit hash rather than a moving tag, so you can review and update deliberately.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Use GitHub's private vulnerability reporting:

1. Go to <https://github.com/kudig-io/klaw/security/advisories/new>
2. Provide as much of the following as you can:
   - A clear description of the issue and its impact
   - Reproduction steps, ideally with a minimal config (redact tokens / kubeconfigs)
   - The affected commit hash or release
   - Your assessment of severity (e.g. cluster-admin escalation, information disclosure, denial of service)
3. We will acknowledge within **5 business days** and aim to ship a fix or mitigation within **30 days** for high-severity issues.

You can also reach the maintainers by opening a GitHub Discussion tagged `security` for non-sensitive questions about hardening (auth setup, network exposure, etc.) — never share exploit details there.

## Scope

In-scope vulnerabilities include, but are not limited to:

- **Authentication / authorization bypass** in the HTTP API (`/api/v1/*`) or ChatOps webhook (`:8081/webhook/*`)
- **Token handling issues** (e.g. non-constant-time comparison, token leakage in logs)
- **Web UI security** (XSS, open redirects, CSRF where relevant)
- **Cluster privilege escalation** when Klaw is configured to act on a target cluster
- **Configuration injection** (e.g. YAML deserialization, path traversal in config load)
- **Supply-chain risks** (typosquatting in `go.mod` / `package.json`, CI workflow tampering)

Out of scope:

- Issues that require physical access to the host running Klaw
- Social engineering or phishing against maintainers
- Vulnerabilities in upstream dependencies that have not been demonstrated against Klaw (please report upstream first)
- Best-practice recommendations without a concrete exploit (use Discussions instead)

## Hardening Recommendations

The defaults are safe for local development, but production deployments should:

- **Reverse proxy / Ingress in front of Klaw** — the Web UI does not auto-inject the Bearer token (see [Known Limitations in README](./README.md#已知限制)); put Klaw behind an auth layer.
- **Inject secrets via environment variables or Kubernetes Secrets** — never commit `KLAW_API_TOKEN`, DingTalk app secret, kubeconfigs, or AI provider keys.
- **`chmod 600` any kubeconfig** stored under `configs/` and add it to `.gitignore` (already covered by the `*.kubeconfig` rule).
- **Use a non-`cluster-admin` ServiceAccount** in production Helm deployments where possible. The chart's default ClusterRole is intentionally broad for the diagnostics engine to read cluster-wide state — scope it down for your use case.
- **Restrict the ChatOps webhook port (`:8081`)** to your messaging platform's IP allowlist if your provider publishes one.
- **Run behind TLS** — Klaw itself doesn't terminate TLS; use a reverse proxy.

## Coordinated Disclosure

We follow a 90-day coordinated disclosure timeline. If you need more time before public disclosure, let us know and we'll work with you.