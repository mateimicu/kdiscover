# Maintenance Plan — kdiscover

**Date:** 2026-02-17
**Branch:** `maintenance/dependency-updates`
**Scope:** Consolidated findings from 5 independent reviews (database/data-persistence, architecture, security, maintenance, testing)

---

## Executive Summary

Five parallel reviews identified **22 unique findings** across reliability, security, code health, and documentation. After de-duplication (e.g., `io/ioutil` deprecation was flagged by all 5 reviewers), these consolidate into **16 actionable items** grouped into 4 priority tiers.

The most impactful issues are: tests that make real AWS network calls (blocking CI reliability), insecure file permissions on kubeconfig backups containing credentials, and non-atomic file writes risking data corruption.

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

### 2. Fix insecure backup file permissions

**Source:** Security Review (Critical), Database Review (Medium)
**Impact:** Kubeconfig backup files are created with 0644 (world-readable) instead of 0600. Kubeconfig files contain cluster credentials.
**Effort:** Low

**Affected code:**
- `cmd/aws_update.go:105` — `os.Create(dst)` defaults to 0666 (masked to ~0644 by umask)

**Fix:** Replace `os.Create(dst)` with `os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)`. Alternatively, read the source file's permissions via `os.Stat()` and apply them to the destination.

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

## P1 — High (Code Health)

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

### 5. Stop logging Certificate Authority data

**Source:** Security Review (Medium)
**Impact:** CA data logged in error path. Low risk (only in error cases, logs typically local), but violates defense-in-depth.
**Effort:** Low

**Affected code:**
- `internal/aws/eks.go:100` — logs `*result.Cluster.CertificateAuthority.Data`

**Fix:** Remove `certificate-authority-data` field from the log entry. Keep cluster name and ARN for debugging.

---

### 6. Improve error handling patterns

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

### 7. Add `context.Context` support to AWS calls

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

### 8. Limit concurrency for region queries

**Source:** Architecture Review (Medium)
**Impact:** One goroutine per region (currently up to ~30 AWS regions). Not dangerous at current scale but poor practice. Could cause AWS throttling.
**Effort:** Low

**Affected code:**
- `internal/aws/eks_cluster.go:64-101`

**Fix:** Add a semaphore (e.g., `golang.org/x/sync/errgroup` with `SetLimit()`) to cap concurrent region queries.

---

### 9. Replace manual file copy with `io.Copy`

**Source:** Database Review (Medium), Architecture Review (Low)
**Impact:** `copyFs` manually manages a 1MB buffer. Dead code at lines 111-113 (unreachable `if err != nil` after successful `os.Create`).
**Effort:** Low

**Affected code:**
- `cmd/aws_update.go:84-131`

**Fix:** Replace the manual read/write loop (lines 116-129) with `io.Copy(destination, source)`. Remove the dead code block at lines 111-113. Use `os.OpenFile` with correct permissions (see item 2).

---

### 10. Fix backup filename generation

**Source:** Database Review (Low)
**Impact:** Infinite `.bak.bak.bak...` chain if many backups exist. Could produce extremely long filenames.
**Effort:** Low

**Affected code:**
- `cmd/aws_update.go:133-151`

**Fix:** Use a numeric suffix pattern: `config.bak`, `config.bak.1`, `config.bak.2`, etc. Add a maximum iteration guard.

---

### 11. Add test coverage for `internal/cluster` package

**Source:** Testing Review (Medium), Maintenance Review (Medium)
**Impact:** Core domain model has 0% test coverage. `PrettyName` uses `html/template` (should be `text/template`) — untested.
**Effort:** Low

**Affected code:**
- `internal/cluster/cluster.go` — entire file untested
- `internal/cluster/testing.go` — test helper exists but no tests use it

**Fix:** Write unit tests for `NewCluster`, `GetUniqueID`, `PrettyName`, `GetConfigCluster`, `GetConfigAuthInfo`. Also replace `html/template` with `text/template` in `PrettyName` (line 7 import, line 77 usage) since context names are not HTML.

---

### 12. Update `.gitignore` with security-sensitive patterns

**Source:** Security Review (Medium)
**Impact:** Risk of accidentally committing credential files.
**Effort:** Low

**Fix:** Add patterns: `.env`, `.env.*`, `*.pem`, `*.key`, `*kubeconfig*`, `*.kubeconfig`.

---

## P3 — Low (Documentation & Cleanup)

### 13. Remove dead code and stale suppressions

**Source:** Database Review (Low), Testing Review (Low), Architecture Review (Low)
**Effort:** Low

**Items:**
- `cmd/aws_update.go:111-113` — dead `if err != nil` block (see item 9)
- `cmd/root_test.go:17` — duplicate `update` flag declaration (also in `main.go:10`, `internal/aws/eks_test.go:19`)
- `.errcheck-exclude` — duplicate file at root; update for io/ioutil removal
- Nolint directives on `main.go:9` referencing removed linters (`varcheck`, `deadcode`)

---

### 14. Resolve stale TODO comments

**Source:** Architecture Review (Medium), Maintenance Review (Medium)
**Effort:** Low

**TODOs (4):**
- `internal/aws/eks.go:27-29` — "test GetClusters function" / "use assert library" — addressed by item 1
- `internal/aws/eks.go:81` — "handle errors better here" — addressed by item 6
- `internal/aws/eks_test.go:41` — "implement DescribeCluster and ListClustersPages"
- `internal/aws/eks_test.go:97` — "populate cluster data here"

**Fix:** Resolve each TODO as part of its parent item, or remove if no longer relevant.

---

### 15. Add community and security documentation

**Source:** Security Review (Medium), Maintenance Review (Medium)
**Effort:** Low

**Missing files:**
- `SECURITY.md` — vulnerability reporting policy
- `CHANGELOG.md` — release history (last release v0.1.7, September 2023)
- `CONTRIBUTING.md` — contribution guidelines (open issue #613)

---

### 16. Repository hygiene

**Source:** Maintenance Review (Low)
**Effort:** Low

**Items:**
- Clean up 10+ stale `copilot/*` remote branches
- Fix FAQ typo: "Configur" → "Configure"
- Update outdated README project board link
- Consider adding `.editorconfig` for consistent formatting
- Consider adding a `Makefile` for common build/test/lint commands

---

## Implementation Order

The recommended implementation sequence groups related changes to minimize rework:

| Phase | Items | Rationale |
|-------|-------|-----------|
| A     | 1, 4  | Fix tests first (unblock CI), replace ioutil (mechanical) |
| B     | 2, 9, 10 | File operations cluster (permissions, io.Copy, backup naming) |
| C     | 3, 5  | Atomic writes and log sanitization (security hardening) |
| D     | 6, 7, 8 | Error handling + context + concurrency (architectural, related changes) |
| E     | 11, 12 | Test coverage and gitignore (quality gates) |
| F     | 13, 14, 15, 16 | Cleanup and documentation (low-risk, independent) |

Each phase should be a separate PR for reviewability.

---

## Findings Cross-Reference

| Finding | DB | Arch | Sec | Maint | Test |
|---------|:--:|:----:|:---:|:-----:|:----:|
| 1. Test network calls | | | | | X |
| 2. Backup permissions | X | | X | | |
| 3. Atomic writes | X | | | | |
| 4. io/ioutil deprecation | X | X | | X | X |
| 5. CA data in logs | | | X | | |
| 6. Error handling | | X | | | |
| 7. context.Context | | X | | | |
| 8. Unbounded concurrency | | X | | | |
| 9. Manual file copy | X | X | | | |
| 10. Backup naming | X | | | | |
| 11. Cluster test coverage | | | | X | X |
| 12. .gitignore patterns | | | X | | |
| 13. Dead code | X | X | | | X |
| 14. Stale TODOs | | X | | X | |
| 15. Community docs | | | X | X | |
| 16. Repo hygiene | | | | X | |
