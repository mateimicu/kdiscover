# Contributing to kdiscover

First off, thank you for considering contributing to kdiscover! It's people like you that make kdiscover such a great tool for discovering and managing Kubernetes clusters.

## Code of Conduct

This project and everyone participating in it is governed by our commitment to creating a welcoming and inclusive environment. By participating, you are expected to uphold this standard and treat all community members with respect.

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for kdiscover. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

**Before Submitting A Bug Report:**
- Check the [FAQ](docs/FAQ.md) for common issues and solutions
- Ensure the bug was not already reported by searching on GitHub under [Issues](https://github.com/mateimicu/kdiscover/issues)
- Check if the issue exists in the latest version

**How Do I Submit A (Good) Bug Report?**

Bugs are tracked as [GitHub issues](https://github.com/mateimicu/kdiscover/issues). Create an issue using the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md) and provide the following information:

- Use a clear and descriptive title
- Describe the exact steps to reproduce the problem
- Provide specific examples and the exact commands you used
- Include the output of `kdiscover version`
- Describe the behavior you observed and what behavior you expected
- Include your environment details (OS, shell, etc.)

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for kdiscover, including completely new features and minor improvements to existing functionality.

**Before Submitting An Enhancement Suggestion:**
- Check the [project board](https://github.com/mateimicu/kdiscover/projects/1) and [milestones](https://github.com/mateimicu/kdiscover/milestones) to see if the enhancement is already planned
- Check if there's already an issue for your enhancement

**How Do I Submit A (Good) Enhancement Suggestion?**

Enhancement suggestions are tracked as [GitHub issues](https://github.com/mateimicu/kdiscover/issues). Create an issue using the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) and provide the following information:

- Use a clear and descriptive title
- Provide a step-by-step description of the suggested enhancement
- Provide specific examples to demonstrate the steps
- Describe the current behavior and explain which behavior you expected instead
- Explain why this enhancement would be useful to most kdiscover users

### Pull Requests

The process described here has several goals:
- Maintain kdiscover's quality
- Fix problems that are important to users
- Engage the community in working toward the best possible kdiscover
- Enable a sustainable system for kdiscover's maintainers to review contributions

Please follow these steps to have your contribution considered by the maintainers:

1. **Fork** the repository and create your branch from `master`
2. **Make your changes** following our coding standards
3. **Add tests** for your changes if applicable
4. **Ensure tests pass** by running `go test ./...`
5. **Ensure code follows standards** by running the linter
6. **Update documentation** if you changed APIs or functionality
7. **Commit your changes** using clear and descriptive commit messages
8. **Push to your fork** and submit a pull request

## Development Environment Setup

### Prerequisites

- Go 1.17 or later
- Git
- Access to AWS CLI configured (for testing AWS functionality)

### Getting Started

1. **Clone the repository**
   ```bash
   git clone https://github.com/mateimicu/kdiscover.git
   cd kdiscover
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the project**
   ```bash
   go build main.go
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## Development Workflow

### Coding Standards

- Follow Go best practices and idioms
- Use `gofmt` to format your code
- Run `golangci-lint` to check for linting issues
- Write clear, readable code with appropriate comments
- Follow the existing code style in the repository

### Testing

- Write tests for new functionality
- Ensure all tests pass before submitting a PR
- Include both unit tests and integration tests where appropriate
- Test coverage should not decrease with new changes

### Commit Messages

- Use clear and meaningful commit messages
- Start with a brief summary (50 characters or less)
- Include more detailed explanation if necessary
- Reference issues and pull requests where applicable

Example:
```
Add support for GKE cluster discovery

Implements basic GKE cluster listing functionality similar to EKS.
Includes authentication via service account and region filtering.

Fixes #123
```

### Branch Naming

Use descriptive branch names that indicate the type of work:
- `feature/add-gke-support`
- `bugfix/fix-context-naming`
- `docs/update-readme`

## Project Structure

```
├── cmd/                 # CLI commands and command-line interface
├── internal/           # Internal packages
│   ├── aws/           # AWS-specific functionality
│   ├── cluster/       # Cluster abstraction and utilities
│   └── kubeconfig/    # Kubeconfig manipulation
├── docs/              # Documentation
├── .github/           # GitHub templates and workflows
└── main.go           # Entry point
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests for a specific package
go test ./internal/aws/
```

## Running the Linter

The project uses `golangci-lint`. You can run it locally:

```bash
golangci-lint run
```

## Documentation

- Update the README.md if your changes affect usage
- Update the [FAQ](docs/FAQ.md) if your changes address common questions
- Add comments to complex code sections
- Update godoc comments for public APIs

## Release Process

Releases are handled by maintainers. The project uses:
- Semantic versioning
- Automated releases via GitHub Actions
- Release notes generated from commit messages and pull requests

## Getting Help

- Check the [FAQ](docs/FAQ.md) for common questions
- Look at existing [issues](https://github.com/mateimicu/kdiscover/issues) and [pull requests](https://github.com/mateimicu/kdiscover/pulls)
- Create a new issue if you need help

## Recognition

Contributors are recognized in release notes and we appreciate all forms of contribution, including:
- Code contributions
- Documentation improvements
- Bug reports
- Feature suggestions
- Testing and feedback

Thank you for contributing to kdiscover! 🚀