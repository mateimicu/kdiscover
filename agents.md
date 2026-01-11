# Agent Guidelines for kdiscover

This document provides guidelines for AI agents and contributors working on the kdiscover codebase.

## Project Overview

**kdiscover** is a Go CLI tool that discovers and configures access to Kubernetes clusters. It currently supports AWS EKS clusters and is designed as a kubectl plugin.

### Project Structure

```
kdiscover/
├── cmd/                    # CLI commands (cobra-based)
│   ├── root.go            # Root command and global flags
│   ├── aws.go             # AWS parent command
│   ├── aws_list.go        # List EKS clusters
│   └── aws_update.go      # Update kubeconfig with clusters
├── internal/
│   ├── aws/               # AWS/EKS integration
│   │   ├── eks.go         # EKS client and cluster discovery
│   │   ├── eks_cluster.go # Cluster type for EKS
│   │   ├── eks_kubeconfig.go # Kubeconfig generation
│   │   └── regions.go     # AWS region utilities
│   ├── cluster/           # Generic cluster abstraction
│   │   ├── cluster.go     # Cluster type definition
│   │   └── testing.go     # Test helpers (mock clusters)
│   └── kubeconfig/        # Kubeconfig management
│       └── kubeconfig.go  # Load, save, merge kubeconfigs
├── main.go                # Entry point
├── go.mod                 # Go module definition
└── Makefile               # Build and test commands
```

## TDD Workflow (Required)

All changes MUST follow Test-Driven Development:

### 1. Write the Test First

Before implementing any feature or fix:

```go
func TestNewFeature(t *testing.T) {
    // Arrange: set up test data
    input := "test-input"
    expected := "expected-output"
    
    // Act: call the function under test
    result := NewFeature(input)
    
    // Assert: verify the result
    assert.Equal(t, expected, result)
}
```

### 2. Run the Test (Should Fail)

```bash
make test
# or
go test -race ./...
```

### 3. Implement the Minimum Code

Write just enough code to make the test pass.

### 4. Run the Test Again (Should Pass)

```bash
make test
```

### 5. Refactor if Needed

Clean up the code while keeping tests green.

## Testing Patterns

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestFunction(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"empty input", "", "", false},
        {"valid input", "test", "result", false},
        {"invalid input", "bad", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Mock Implementations

Use interfaces for dependencies to enable mocking. See `internal/aws/eks_test.go` for examples:

```go
// Define interface for the dependency
type ClusterGetter interface {
    GetClusters(ch chan<- *cluster.Cluster)
}

// Create mock implementation for tests
type fakeClusterGetter struct {
    Clusters []*cluster.Cluster
}

func (c *fakeClusterGetter) GetClusters(ch chan<- *cluster.Cluster) {
    for _, cls := range c.Clusters {
        ch <- cls
    }
    close(ch)
}
```

### Golden File Tests

For complex output, use golden files (see `internal/kubeconfig/kubeconfig_test.go`):

```go
func TestOutput(t *testing.T) {
    goldenPath := filepath.Join("testdata", t.Name()+".golden")
    
    if *update {
        // Update golden file
        if err := os.WriteFile(goldenPath, result, 0644); err != nil {
            t.Fatal(err)
        }
    }
    
    expected, _ := os.ReadFile(goldenPath)
    assert.Equal(t, expected, result)
}
```

Update golden files with:
```bash
make golden-update
```

### Test Helpers

Use helpers from `internal/cluster/testing.go`:

```go
// Get predictable mock clusters for testing
clusters := cluster.GetPredictableMockClusters(5)

// Get random mock clusters
clusters := cluster.GetMockClusters(10)
```

## Development Commands

```bash
# Run all tests with race detection
make test

# Run linter
make lint

# Build binary
make build

# Generate coverage report
make coverage

# Update golden test files
make golden-update

# Run all checks (lint + test)
make check
```

## Code Style

### Linting

The project uses golangci-lint with strict configuration. Run before committing:

```bash
make lint
```

### Key Rules

1. **No unused code** - Remove dead code
2. **Error handling** - Always handle errors explicitly
3. **Function length** - Keep functions focused and short
4. **Magic numbers** - Use named constants
5. **Imports** - Use goimports for formatting

### Logging

Use logrus for structured logging:

```go
import log "github.com/sirupsen/logrus"

log.WithFields(log.Fields{
    "cluster": clusterName,
    "region":  region,
}).Info("Processing cluster")
```

## Pull Request Requirements

1. **All tests pass**: `make test` must succeed
2. **Linter passes**: `make lint` must succeed
3. **New tests for new code**: Follow TDD
4. **Update golden files if needed**: `make golden-update`
5. **No decrease in coverage**: Check with `make coverage`

## Architecture Principles

### Dependency Injection

Design for testability by injecting dependencies:

```go
// Good: Accepts interface
type EKSClient struct {
    EKS    eksiface.EKSAPI  // Interface, can be mocked
    Region string
}

// Bad: Creates its own dependencies
func NewEKSClient() *EKSClient {
    sess := session.Must(session.NewSession())  // Hard to test
    return &EKSClient{EKS: eks.New(sess)}
}
```

### Domain-Driven Design

Keep domain logic separate from infrastructure:

- `internal/cluster/` - Domain model (Cluster type)
- `internal/aws/` - AWS-specific implementation
- `cmd/` - CLI/presentation layer

### Error Handling

Return errors, don't panic:

```go
// Good
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    // ...
}

// Bad
func LoadConfig(path string) *Config {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err)  // Don't do this
    }
    // ...
}
```

## Common Tasks

### Adding a New Cloud Provider

1. Create `internal/<provider>/` package
2. Implement `ClusterGetter` interface
3. Add tests with mocks
4. Add CLI command in `cmd/`
5. Update documentation

### Modifying Cluster Discovery

1. Write test in `internal/aws/eks_test.go`
2. Update mock if needed
3. Implement change in `eks.go`
4. Verify with `make test`

### Adding a New CLI Flag

1. Add test in `cmd/<command>_test.go`
2. Add flag definition
3. Wire up to business logic
4. Update help text
