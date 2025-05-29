# Development Guide

## 🛠️ Local Development Setup

### Prerequisites
- Go 1.23 or later (as specified in go.mod)
- Terraform installed and available in PATH
- Git for version control

### Quick Start
1. **Clone the repository**:
   ```bash
   git clone https://github.com/sorinlg/tf-manage2.git
   cd tf-manage2
   ```

2. **Build the project**:
   ```bash
   go build -o tf .
   ```

3. **Test with the example project**:
   ```bash
   cd examples/tfm-project
   go run ../../ project1 sample_module dev instance_x init
   go run ../../ project1 sample_module dev instance_x plan
   go run ../../ project1 sample_module dev instance_x apply
   ```

### Release management
- Use [svu](https://github.com/caarlos0/svu) for semantic versioning based on conventional commits
- Use [GoReleaser](https://goreleaser.com/) for building and releasing binaries
- Follow the [official guide](https://goreleaser.com/install/) for installation instructions

#### Initial setup
```bash
# Install svu (Go-based semantic versioning)
go install github.com/caarlos0/svu@latest

# Install GoReleaser
brew install goreleaser/tap/goreleaser
brew install goreleaser

# Initialize GoReleaser in the project
goreleaser init

# Check current version and what would be next
svu current
svu next

# Local testing with GoReleaser
goreleaser release --snapshot --clean

# Release (done automatically by GitHub Actions)
export GITHUB_TOKEN="$(gh auth token)"
git tag $(svu next)
git push origin $(svu next)
goreleaser release --clean
```

#### svu Usage Examples
```bash
# Check current version
svu current

# Get next version based on conventional commits
svu next

# Get next pre-release version (for develop branch)
svu prerelease --pre-release=rc

# Force specific version types
svu major   # Force major version bump
svu minor   # Force minor version bump
svu patch   # Force patch version bump

# Check what commits would trigger version bump
git log $(svu current)..HEAD --oneline
```

#### Commit Message Strategy

**For develop branch** (increments only RC number):
- Use non-bumping commit types to avoid changing base version
- `chore:` - maintenance tasks, dependency updates
- `docs:` - documentation changes
- `style:` - code formatting, whitespace
- `refactor:` - code refactoring without functionality changes
- `test:` - adding or updating tests
- `ci:` - CI/CD configuration changes

**For main branch** (proper semantic versioning):
- `feat:` - new features (minor version bump)
- `fix:` - bug fixes (patch version bump)
- `BREAKING CHANGE:` - breaking changes (major version bump)

**Example develop workflow**:
```bash
git commit -m "chore: update dependencies"
git commit -m "refactor: improve error handling"
git commit -m "test: add integration tests"
# Result: v0.1.0 → v0.1.1-rc.0 → v0.1.1-rc.1 → v0.1.1-rc.2
```

## 🏗️ Project Structure

```
tf-manage2/
├── main.go                           # CLI entrypoint
├── internal/
│   ├── cli/cli.go                    # Command line interface & argument parsing
│   ├── config/config.go              # Configuration file parsing (.tfm.conf)
│   ├── framework/
│   │   ├── runner.go                 # Command execution framework
│   │   └── printer.go                # ANSI colored output & formatting
│   └── terraform/manager.go          # Terraform operations & business logic
└── examples/tfm-project/             # Test project for development
```

## 🔄 Development Workflow

### Iterating with go run
The most efficient way to test changes during development:

```bash
# Navigate to test project
cd examples/tfm-project

# Test different commands using go run
go run ../../ project1 sample_module dev instance_x init
go run ../../ project1 sample_module dev instance_x plan
go run ../../ project1 sample_module dev instance_x apply
go run ../../ project1 sample_module dev instance_x destroy

# Test validation
go run ../../ invalid_product sample_module dev instance_x plan

# Test workspace management
go run ../../ project1 sample_module dev instance_x workspace list

# Test with custom flags
go run ../../ project1 sample_module dev instance_x 'init -reconfigure'
```

### Building for Testing
When you need a stable binary for extended testing:

```bash
# Build optimized binary
go build -ldflags="-s -w" -o tf .

# Test the binary
./tf project1 sample_module dev instance_x plan
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config
go test ./internal/terraform
```

## 🧪 Testing Strategy

### Manual Testing Checklist
Test these scenarios during development:

#### ✅ Basic Operations
- [ ] `init` - Initialize new terraform workspace
- [ ] `plan` - Generate execution plan
- [ ] `apply` - Apply changes (test interactive prompt)
- [ ] `destroy` - Destroy resources (test interactive prompt)
- [ ] `output` - Display outputs

#### ✅ Validation Tests
- [ ] Invalid product name
- [ ] Invalid module name
- [ ] Invalid environment
- [ ] Missing config file
- [ ] Invalid repository configuration

#### ✅ Workspace Management
- [ ] Create new workspace automatically
- [ ] Select existing workspace
- [ ] List workspaces
- [ ] Workspace naming convention

#### ✅ Variable Integration
- [ ] Check terraform variables are passed correctly
- [ ] Verify `tfm_*` variables in terraform output
- [ ] Test variable integration in triggers

Example terraform output showing variable integration:
```hcl
# Terraform output showing tf-manage variables
Changes to Outputs:
  ~ example = {
      ~ null_resource = {
          ~ triggers = {
              ~ always_run      = "2025-05-24T20:56:59Z" -> (known after apply)
                # tf-manage variables automatically injected:
                "env" = "dev"
                "module" = "sample_module"
                "module_instance" = "instance_x"
                "product" = "project1"
                "repo" = "tfm-project"
            }
        }
    }
```

#### ✅ Interactive vs Unattended Mode
```bash
# Test operator mode (interactive)
go run ../../ project1 sample_module dev instance_x apply

# Test manual override
TF_EXEC_MODE_OVERRIDE=1 go run ../../ project1 sample_module dev instance_x apply

# Test CI/CD detection - GitHub Actions
GITHUB_ACTIONS=true go run ../../ project1 sample_module dev instance_x apply

# Test CI/CD detection - GitLab CI
GITLAB_CI=true go run ../../ project1 sample_module dev instance_x apply

# Test CI/CD detection - CircleCI
CIRCLECI=true go run ../../ project1 sample_module dev instance_x apply

# Test CI/CD detection - Jenkins (environment variable)
JENKINS_URL=http://jenkins.example.com go run ../../ project1 sample_module dev instance_x apply

# Test CI/CD detection - Jenkins (legacy)
USER=jenkins go run ../../ project1 sample_module dev instance_x apply

# Test CI/CD detection - Generic CI
CI=true go run ../../ project1 sample_module dev instance_x apply

# Verify execution mode output:
# - "operator" = Interactive mode (prompts for confirmation)
# - "unattended" = Auto-approval mode for CI/CD
USER=jenkins go run ../../ project1 sample_module dev instance_x apply
```

### Example Project Structure
The `examples/tfm-project` provides a complete test environment:

```
tfm-project/
├── .tfm.conf                        # Repository configuration
└── terraform/
    ├── environments/                 # Environment-specific configs
    │   └── project1/
    │       ├── dev/
    │       │   └── sample_module/
    │       │       └── instance_x.tfvars
    │       ├── staging/
    │       └── prod/
    └── modules/                      # Reusable terraform modules
        └── sample_module/
            ├── main.tf              # Main terraform configuration
            ├── variables.tf         # Input variables
            ├── outputs.tf           # Output values
            └── tfm.tf              # tf-manage integration variables
```

## 🐛 Debugging

### Enable Debug Output
```bash
# Add debug output to any function
framework.Debug("Debug message here")

# Check configuration parsing
framework.Debug(fmt.Sprintf("Config: %+v", config))
```

### Common Issues

#### Issue: Command not found
```bash
# Make sure terraform is in PATH
which terraform

# Or specify full path in development
export PATH="/usr/local/bin:$PATH"
```

#### Issue: Workspace errors
```bash
# Reset terraform state for clean testing
rm -rf .terraform/
go run ../../ project1 sample_module dev instance_x init
```

#### Issue: Permission errors
```bash
# Ensure terraform directories are writable
chmod -R 755 terraform/
```

## 📝 Code Style Guidelines

### Go Best Practices
- Use `gofmt` for consistent formatting
- Follow Go naming conventions (PascalCase for exports, camelCase for internal)
- Keep functions focused and single-purpose
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### Error Handling
```go
// Preferred error handling pattern
if err := someOperation(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use framework.Error for user-facing errors
if !isValid {
    framework.Error("Configuration is invalid")
    return fmt.Errorf("validation failed")
}
```

### Framework Usage
```go
// Use framework functions for consistent output
framework.Info("Starting operation")
framework.Debug("Debug information")
framework.Error("Error occurred")

// Use framework command execution
result := framework.RunCmd(command, message, flags, failMessage)
```

## 🚀 Performance Testing

### Benchmarking
```bash
# Time command execution
time go run ../../ project1 sample_module dev instance_x plan

# Compare with bash version
time /path/to/bash/tf project1 sample_module dev instance_x plan
```

### Memory Profiling
```bash
# Build with profiling
go build -o tf .

# Run with memory profiling
GODEBUG=gctrace=1 ./tf project1 sample_module dev instance_x plan
```

## 🔗 Integration with Original tf-manage

### Side-by-side Testing
You can run both versions to compare behavior:

```bash
# Test tf-manage2 (Go version)
cd examples/tfm-project
go run ../../ project1 sample_module dev instance_x plan

# Test original tf-manage (bash version, if available)
/path/to/original/tf project1 sample_module dev instance_x plan
```

### Configuration Compatibility
The Go version reads the same `.tfm.conf` files:

```bash
#!/bin/bash
export __tfm_repo_name='tfm-project'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
```

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Terraform CLI Documentation](https://www.terraform.io/docs/cli/index.html)
- [Original tf-manage Repository](https://github.com/adobe/tf-manage) - Original bash implementation (v6.4.0)
- [What's New in tf-manage2](WHATS_NEW.md)
