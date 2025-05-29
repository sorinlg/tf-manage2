# Homebrew Tap Setup Instructions

This document explains how to set up the Homebrew tap repository that GoReleaser will use to automatically publish brew formulas.

## Prerequisites

1. Create a new GitHub repository named `homebrew-tap` under your account
2. Initialize it with a README.md

## Repository Setup

### 1. Create the Repository

```bash
# Using GitHub CLI (recommended)
gh repo create homebrew-tap --public --description "Homebrew tap for tf-manage2"

# Or create manually via GitHub web interface
```

### 2. Initialize the Repository Structure

```bash
# Clone the new repository
git clone https://github.com/sorinlg/homebrew-tap.git
cd homebrew-tap

# Create Formula directory
mkdir -p Formula

# Create initial README
cat > README.md << 'EOF'
# homebrew-tap

Homebrew tap for tf-manage2 and related tools.

## Usage

```bash
brew tap sorinlg/tap
brew install tf-manage2
```

## Available Formulas

- `tf-manage2` - Terraform workspace manager with enhanced CI/CD detection
EOF

# Commit initial structure
git add .
git commit -m "Initial tap structure"
git push origin main
```

### 3. Test the Tap Configuration

After the first release is published, you can test the tap:

```bash
# Add the tap
brew tap sorinlg/tap

# Install tf-manage2
brew install tf-manage2

# Test the installation
tf --version
```

## How It Works

1. When a new release is created, GoReleaser automatically:
   - Builds binaries for all supported platforms
   - Creates release archives
   - Pushes a Homebrew formula to the `homebrew-tap` repository
   - Updates the formula with the correct download URLs and checksums

2. The formula will be created at `Formula/tf-manage2.rb`

3. Users can then install via `brew install sorinlg/tap/tf-manage2`

## GoReleaser Configuration

The `.goreleaser.yaml` file contains the `brews` section that handles this automatically:

```yaml
brews:
  - name: tf-manage2
    repository:
      owner: sorinlg
      name: homebrew-tap
      branch: main
    # ... other configuration
```

## Notes

- The tap repository must exist before running the first release
- GoReleaser needs write access to the tap repository (handled via GITHUB_TOKEN)
- The formula will be automatically updated on each release
- Dependencies like Terraform are marked as optional in the formula
