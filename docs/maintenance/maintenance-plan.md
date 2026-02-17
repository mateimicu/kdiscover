# Maintenance Plan — kdiscover

**Date:** 2026-02-17
**Branch:** `maintenance/dependency-updates`
**Scope:** Consolidated findings from 5 independent reviews (database/data-persistence, architecture, security, maintenance, testing)

---

## Executive Summary

Five parallel reviews identified **28 unique findings** across reliability, security, code health, and documentation. After de-duplication (e.g., `io/ioutil` deprecation was flagged by all 5 reviewers, backup file security flagged by 3), these consolidate into **19 actionable items** grouped into 4 priority tiers.

The most impactful issues are: tests that make real AWS network calls (blocking CI reliability), a TOCTOU race condition with insecure file permissions on kubeconfig backups, CI/CD workflow security gaps (custom PAT, `pull_request_target`, missing permissions), and non-atomic file writes risking data corruption.

---

## P0 — Critical (Reliability & Security)

### 1. Fix tests making real AWS network calls

**Source:** Testing Review (Critical), Architecture Review (related)
**Impact:** Test suite cannot complete reliably; `cmd/` and `internal/aws/` tests timeout in CI.
**Effort:** Medium

**Affected tests:**
- `TestQueryAllRegions` — `cmd/aws_test.go:37`
- `Test_CascadingPersistPreRunEHackWithLoggingLevels` — `cmd/root_test.go:37`
- `TestGetEKSClusters` — `internal/aws/eks_cluster_test.go:39`

**Fix:** These tests execute real Cobra commands that call `aws.GetEKSClusters()` with live AWS SDK sessions. The `internal/aws/` package already has mock patterns (mock EKS client with `eksiface.EKSAPI`). Refactor `cmd/` tests to inject mock `ClusterGetter` implementations rather than hitting real AWS endpoints. Add `t.Short()` guards or build tags for any remaining integration tests.

---

### 2. Fix TOCTOU race and insecure backup file permissions

**Source:** Security Review (Critical), Database Review (Medium)
**Impact:** Two compounding issues in `copyFs`:
1. **TOCTOU race condition:** `os.Stat(dst)` at line 100 checks if the destination exists, then `os.Create(dst)` at line 105 creates it. Between the check and create, another process could create the file — the backup silently overwrites it.
2. **Insecure permissions:** `os.Create(dst)` defaults to 0666 (masked to ~0644 by umask). Kubeconfig backup files containing cluster credentials become world-readable.
**Effort:** Low

**Affected code:**
- `cmd/aws_update.go:100-105` — stat-then-create race + insecure default permissions

**Fix:** Replace both the `os.Stat` check (line 100-103) and `os.Create` (line 105) with a single atomic call:
```go
os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
```
`O_EXCL` makes creation fail atomically if the file already exists, eliminating both the race and the permissions issue in one change.

---

### 3. Implement atomic kubeconfig writes

**Source:** Database Review (High)
**Impact:** If `kubeconfig.Persist()` is interrupted mid-write (crash, signal, disk full), the kubeconfig file is left truncated/corrupted. Users lose access to all configured clusters.
**Effort:** Medium

**Affected code:**
- `internal/kubeconfig/kubeconfig.go:46-48` — `clientcmd.WriteToFile()` writes directly
- `cmd/aws_update.go:71` — caller site

**Fix:** Write to a temporary file in the same directory, then `os.Rename()` (atomic on POSIX). Pattern: write to `<path>.tmp.<random>`, `fsync`, then rename to `<path>`.

---

## P1 — High (Code Health & CI/CD Security)

### 4. Replace deprecated `io/ioutil` usage

**Source:** All 5 reviews flagged this
**Impact:** `io/ioutil` is deprecated since Go 1.16 (5+ years ago). Current `go 1.24.0` in go.mod. Linter suppressions mask the issue.
**Effort:** Low

**Affected files (7):**
- `cmd/root.go:6,44` — `ioutil.Discard` → `io.Discard`
- `cmd/root_test.go:7,42,50,51,63` — `ioutil.TempDir` → `os.MkdirTemp`, `ioutil.Discard` → `io.Discard`
- `cmd/aws_test.go:6,41,52` — same replacements
- `cmd/aws_update_test.go:5,12,21,40,51,63,71` — `ioutil.TempDir` → `os.MkdirTemp`, `ioutil.WriteFile` → `os.WriteFile`
- `cmd/version_test.go:6,26` — `ioutil.TempDir` → `os.MkdirTemp`
- `internal/aws/eks_test.go:8,241,301` — `ioutil.Discard` → `io.Discard`
- `.errcheck-exclude:1` and `.golangci.yml:41` — update references

**Fix:** Mechanical replacements: `ioutil.ReadFile` → `os.ReadFile`, `ioutil.WriteFile` → `os.WriteFile`, `ioutil.TempDir` → `os.MkdirTemp`, `ioutil.Discard` → `io.Discard`.

---

### 5. Harden CI/CD workflow security

**Source:** Security Review (High — expanded)
**Impact:** Three workflow configuration issues that weaken the CI/CD security posture.
**Effort:** Low-Medium

**Findings:**

**5a. Custom `GH_PAT` in release workflow (HIGH)**
- `release.yaml:34` — `GITHUB_TOKEN: ${{ secrets.GH_PAT }}` uses a custom Personal Access Token instead of the default `GITHUB_TOKEN`.
- PATs are long-lived, have broader scope than needed, and are harder to audit/rotate. If compromised, the attacker gets all permissions the PAT grants across all repos it can access.
- **Fix:** Evaluate whether the default `GITHUB_TOKEN` (scoped to the repository, short-lived, auto-rotated) is sufficient for goreleaser and krew-release-bot. If the PAT is needed for cross-repo krew-index updates, document why and ensure it has minimal scopes.

**5b. `pull_request_target` trigger in auto-merge workflow (HIGH)**
- `auto-merge.yaml:2` — `on: pull_request_target` runs the workflow in the context of the base branch with write permissions.
- Mitigated by `if: ${{ github.actor == 'dependabot[bot]' }}` — only runs for Dependabot PRs. However, `pull_request_target` is a known attack vector: if the workflow ever checks out PR code or adds steps that process PR content, it could be exploited.
- **Fix:** Add an explicit comment documenting why `pull_request_target` is used and the security invariant (no checkout of PR code). Consider adding `environment:` protection or using `pull_request` trigger with a separate approval workflow.

**5c. Missing explicit permissions in CI workflow (HIGH)**
- `golang-ci.yaml` has no top-level `permissions:` block. It inherits the repository default, which may grant overly broad write access.
- Contrast with `release.yaml` (has `permissions: contents: write`) and `auto-merge.yaml` (has explicit permissions).
- **Fix:** Add `permissions: read-all` (or minimal `contents: read`) at the top level of `golang-ci.yaml`. This follows the principle of least privilege and matches GitHub's security hardening recommendations.

---

### 6. Stop logging Certificate Authority data

**Source:** Security Review (Medium)
**Impact:** CA data logged in error path. Low risk (only in error cases, logs typically local), but violates defense-in-depth.
**Effort:** Low

**Affected code:**
- `internal/aws/eks.go:100` — logs `*result.Cluster.CertificateAuthority.Data`

**Fix:** Remove `certificate-authority-data` field from the log entry. Keep cluster name and ARN for debugging.

---

### 7. Improve error handling patterns

**Source:** Architecture Review (High)
**Impact:** Errors silently swallowed in cluster listing. Users get incomplete results with no indication of failures.
**Effort:** Medium

**Affected code:**
- `internal/aws/eks.go:44-51` — `detailCluster` errors logged but not surfaced
- `internal/aws/eks.go:64-69` — `ListClustersPages` error only logged
- `internal/aws/eks.go:81` — TODO: "handle errors better here"
- `internal/aws/eks_cluster.go:64-101` — no error propagation from goroutines

**Fix:** Return errors alongside clusters from `GetClusters()` (e.g., `[]error` return or error channel). Let callers decide whether partial results are acceptable. Address the TODO at line 81 by wrapping AWS errors with context.

---

### 8. Add `context.Context` support to AWS calls

**Source:** Architecture Review (High)
**Impact:** No way to cancel long-running region scans. No timeout for AWS API calls. Users must kill the process.
**Effort:** Medium-High

**Affected code:**
- `internal/aws/eks.go:30` — `GetClusters` takes no context
- `internal/aws/eks.go:74` — `detailCluster` takes no context
- `internal/aws/eks_cluster.go:40` — `GetEKSClusters` takes no context

**Fix:** Add `context.Context` as first parameter to `GetClusters`, `detailCluster`, `GetEKSClusters`, and `getEKSClusters`. Pass context to AWS SDK calls (`ListClustersPagesWithContext`, `DescribeClusterWithContext`). Wire context from Cobra command through the call chain.

---

## P2 — Medium (Quality)

### 9. Limit concurrency for region queries

**Source:** Architecture Review (Medium)
**Impact:** One goroutine per region (currently up to ~30 AWS regions). Not dangerous at current scale but poor practice. Could cause AWS throttling.
**Effort:** Low

**Affected code:**
- `internal/aws/eks_cluster.go:64-101`

**Fix:** Add a semaphore (e.g., `golang.org/x/sync/errgroup` with `SetLimit()`) to cap concurrent region queries.

---

### 10. Replace manual file copy with `io.Copy`

**Source:** Database Review (Medium), Architecture Review (Low)
**Impact:** `copyFs` manually manages a 1MB buffer. Dead code at lines 111-113 (unreachable `if err != nil` after successful `os.Create`).
**Effort:** Low

**Affected code:**
- `cmd/aws_update.go:84-131`

**Fix:** Replace the manual read/write loop (lines 116-129) with `io.Copy(destination, source)`. Remove the dead code block at lines 111-113. Use `os.OpenFile` with correct permissions (see item 2).

---

### 11. Fix backup filename generation

**Source:** Database Review (Low)
**Impact:** Infinite `.bak.bak.bak...` chain if many backups exist. Could produce extremely long filenames.
**Effort:** Low

**Affected code:**
- `cmd/aws_update.go:133-151`

**Fix:** Use a numeric suffix pattern: `config.bak`, `config.bak.1`, `config.bak.2`, etc. Add a maximum iteration guard.

---

### 12. Add test coverage for `internal/cluster` package

**Source:** Testing Review (Medium), Maintenance Review (Medium)
**Impact:** Core domain model has 0% test coverage. `PrettyName` uses `html/template` (should be `text/template`) — untested.
**Effort:** Low

**Affected code:**
- `internal/cluster/cluster.go` — entire file untested
- `internal/cluster/testing.go` — test helper exists but no tests use it

**Fix:** Write unit tests for `NewCluster`, `GetUniqueID`, `PrettyName`, `GetConfigCluster`, `GetConfigAuthInfo`. Also replace `html/template` with `text/template` in `PrettyName` (line 7 import, line 77 usage) since context names are not HTML.

---

### 13. Update `.gitignore` with security-sensitive patterns

**Source:** Security Review (Medium)
**Impact:** Risk of accidentally committing credential files.
**Effort:** Low

**Fix:** Add patterns: `.env`, `.env.*`, `*.pem`, `*.key`, `*kubeconfig*`, `*.kubeconfig`.

---

### 14. Triage Dependabot security alerts

**Source:** Security Review (expanded)
**Impact:** 5 open Dependabot alerts (1 high, 4 moderate) on the default branch. After Phase 2 dependency updates on `maintenance/dependency-updates`, these may already be resolved but the alerts remain open on `master`.
**Effort:** Low

**Fix:** After merging the dependency update PR to `master`, verify alerts auto-close. For any remaining alerts, check if the vulnerable dependency is actually reachable in the code path (transitive dependency from `go-pretty`, `go-version`, etc.). Dismiss confirmed false positives with documented justification.

---

## P3 — Low (Documentation & Cleanup)

### 15. Remove dead code and stale suppressions

**Source:** Database Review (Low), Testing Review (Low), Architecture Review (Low)
**Effort:** Low

**Items:**
- `cmd/aws_update.go:111-113` — dead `if err != nil` block (see item 10)
- `cmd/root_test.go:17` — duplicate `update` flag declaration (also in `main.go:10`, `internal/aws/eks_test.go:19`)
- `.errcheck-exclude` — duplicate file at root; update for io/ioutil removal
- Nolint directives on `main.go:9` referencing removed linters (`varcheck`, `deadcode`)

---

### 16. Resolve stale TODO comments

**Source:** Architecture Review (Medium), Maintenance Review (Medium)
**Effort:** Low

**TODOs (4):**
- `internal/aws/eks.go:27-29` — "test GetClusters function" / "use assert library" — addressed by item 1
- `internal/aws/eks.go:81` — "handle errors better here" — addressed by item 7
- `internal/aws/eks_test.go:41` — "implement DescribeCluster and ListClustersPages"
- `internal/aws/eks_test.go:97` — "populate cluster data here"

**Fix:** Resolve each TODO as part of its parent item, or remove if no longer relevant.

---

### 17. Add community and security documentation

**Source:** Security Review (Medium), Maintenance Review (Medium)
**Effort:** Low

**Missing files:**
- `SECURITY.md` — vulnerability reporting policy
- `CHANGELOG.md` — release history (last release v0.1.7, September 2023)
- `CONTRIBUTING.md` — contribution guidelines (open issue #613)

---

### 18. Repository hygiene

**Source:** Maintenance Review (Low)
**Effort:** Low

**Items:**
- Clean up 10+ stale `copilot/*` remote branches
- Fix FAQ typo: "Configur" → "Configure"
- Update outdated README project board link
- Consider adding `.editorconfig` for consistent formatting
- Consider adding a `Makefile` for common build/test/lint commands

---

### 19. Add input validation for AWS region parameters

**Source:** Security Review (Low — expanded)
**Impact:** The `--aws-partitions` flag values are passed directly to `aws.GetRegions()` and then to `session.NewSession()` without validation. Invalid regions produce unhelpful errors from the AWS SDK. Low security risk (invalid regions just fail API calls), but degrades UX.
**Effort:** Low

**Affected code:**
- `cmd/aws.go` — partition/region flag handling
- `internal/aws/regions.go` — region resolution

**Fix:** Validate that supplied partition names are in the known set before creating sessions. Return a clear error for unknown partitions.

---

## Implementation Order

The recommended implementation sequence groups related changes to minimize rework:

| Phase | Items | Rationale |
|-------|-------|-----------|
| A     | 1, 4  | Fix tests first (unblock CI), replace ioutil (mechanical) |
| B     | 2, 10, 11 | File operations cluster (TOCTOU + permissions, io.Copy, backup naming) |
| C     | 3, 5, 6, 13 | Security hardening (atomic writes, CI/CD workflows, log sanitization, gitignore) |
| D     | 7, 8, 9 | Architectural (error handling, context, concurrency) |
| E     | 12, 14 | Quality gates (test coverage, Dependabot triage) |
| F     | 15-19 | Cleanup, documentation, and low-priority items |

Each phase should be a separate PR for reviewability.

---

## Findings Cross-Reference

| Finding | DB | Arch | Sec | Maint | Test |
|---------|:--:|:----:|:---:|:-----:|:----:|
| 1. Test network calls | | | | | X |
| 2. TOCTOU + backup permissions | X | | X | | |
| 3. Atomic writes | X | | | | |
| 4. io/ioutil deprecation | X | X | | X | X |
| 5. CI/CD workflow security | | | X | | |
| 6. CA data in logs | | | X | | |
| 7. Error handling | | X | | | |
| 8. context.Context | | X | | | |
| 9. Unbounded concurrency | | X | | | |
| 10. Manual file copy | X | X | | | |
| 11. Backup naming | X | | | | |
| 12. Cluster test coverage | | | | X | X |
| 13. .gitignore patterns | | | X | | |
| 14. Dependabot alert triage | | | X | | |
| 15. Dead code | X | X | | | X |
| 16. Stale TODOs | | X | | X | |
| 17. Community docs | | | X | X | |
| 18. Repo hygiene | | | | X | |
| 19. Region input validation | | | X | | |
