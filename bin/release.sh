#!/usr/bin/env bash
# Semantic release script for actions.sync-oci-references
# Supports both interactive (local) and CI (GitHub Actions) modes
#
# Interactive Usage:
#   ./bin/release.sh           # Interactive release
#   ./bin/release.sh --dry-run # Preview next version without releasing
#
# CI Usage (GitHub Actions):
#   Sets GITHUB_ACTIONS=true environment variable
#   Writes outputs to $GITHUB_OUTPUT
#   Non-interactive, no prompts

set -euo pipefail

# Detect CI mode (GitHub Actions)
CI_MODE="${GITHUB_ACTIONS:-false}"

# Colors (disabled in CI mode)
if [[ "$CI_MODE" == "true" ]]; then
  RED=''
  GREEN=''
  YELLOW=''
  BLUE=''
  NC=''
else
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BLUE='\033[0;34m'
  NC='\033[0m'
fi

error() { echo -e "${RED}Error: $1${NC}" >&2; }
info() { echo -e "${GREEN}$1${NC}"; }
warning() { echo -e "${YELLOW}Warning: $1${NC}"; }
preview() { echo -e "${BLUE}$1${NC}"; }

# Parse arguments (interactive mode only)
dry_run=false
if [[ "$CI_MODE" == "false" ]] && [[ "${1:-}" == "--dry-run" ]]; then
  dry_run=true
  preview "========================================="
  preview "DRY RUN MODE - No changes will be made"
  preview "========================================="
  echo ""
fi

# Ensure svu is available
if ! command -v svu &> /dev/null; then
  error "svu not found. Install with: go install github.com/caarlos0/svu/v3@latest"
  exit 1
fi

# Get current branch
branch="${GITHUB_REF_NAME:-$(git rev-parse --abbrev-ref HEAD)}"

# Determine release type based on branch
case "$branch" in
  main)
    release_type="stable"
    svu_command="svu next"
    git_force_flag=""
    is_prerelease="false"
    ;;
  develop)
    release_type="prerelease"
    svu_command="svu prerelease --prerelease=rc --tag.mode=current"
    git_force_flag="-f"
    is_prerelease="true"
    ;;
  *)
    error "Unsupported branch: $branch (must be 'main' or 'develop')"
    exit 1
    ;;
esac

# Interactive mode: validate working directory
if [[ "$CI_MODE" == "false" ]]; then
  # Ensure working directory is clean
  if ! git diff-index --quiet HEAD --; then
    error "Working directory is not clean. Please commit or stash your changes."
    exit 1
  fi

  # Ensure we're up to date with remote (for stable releases only)
  if [[ "$release_type" == "stable" ]]; then
    if ! git merge-base --is-ancestor origin/"$branch" HEAD; then
      error "Branch is behind origin/$branch. Pull latest changes first."
      exit 1
    fi
  fi

  # Warn about unpushed commits
  if ! git merge-base --is-ancestor HEAD origin/"$branch"; then
    warning "Branch has unpushed commits ahead of origin/$branch"
  fi
fi

# Calculate next version
tag=$(eval $svu_command)
if [[ -z "$tag" ]]; then
  error "Failed to determine next version tag"
  exit 1
fi

# Check if release already exists (requires gh CLI)
should_release="true"
if command -v gh &> /dev/null; then
  if gh release view "$tag" &>/dev/null 2>&1; then
    if [[ "$is_prerelease" == "true" ]]; then
      # For prereleases: delete and recreate
      if [[ "$CI_MODE" == "false" ]]; then
        warning "Prerelease $tag exists, will be deleted and recreated"
      else
        echo "🔄 Prerelease $tag exists, deleting to recreate..."
      fi
      gh release delete "$tag" --yes
      should_release="true"
    else
      # For stable releases: skip if exists
      should_release="false"
      if [[ "$CI_MODE" == "false" ]]; then
        warning "Release $tag already exists"
      else
        echo "⏭️  Release $tag already exists, skipping"
      fi
    fi
  fi
else
  if [[ "$CI_MODE" == "true" ]]; then
    error "gh CLI not found (required in CI mode)"
    exit 1
  fi
fi

# Calculate previous stable tag (CI mode only)
previous_stable_tag=""
if [[ "$CI_MODE" == "true" ]] && command -v gh &> /dev/null; then
  # Get the latest stable release (excluding pre-releases)
  previous_stable_tag=$(gh release list --exclude-pre-releases --limit 1 --json tagName --jq '.[0].tagName' 2>/dev/null || echo "")
  if [[ -n "$previous_stable_tag" ]]; then
    echo "📌 Previous stable tag: $previous_stable_tag"
  else
    echo "⚠️ No previous stable release found"
  fi
fi

# Output to GITHUB_OUTPUT (CI mode only)
if [[ "$CI_MODE" == "true" ]] && [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "version=$tag"
    echo "is-prerelease=$is_prerelease"
    echo "should-release=$should_release"
    echo "previous-stable-tag=$previous_stable_tag"
  } >> "$GITHUB_OUTPUT"
fi

# Exit early if release already exists
if [[ "$should_release" == "false" ]]; then
  exit 0
fi

# Interactive mode: Display release information and confirm
if [[ "$CI_MODE" == "false" ]]; then
  info "========================================="
  info "Release Information"
  info "========================================="
  info "Branch:       $branch"
  info "Release type: $release_type"
  info "Next version: $tag"
  info "Force push:   $([[ "$git_force_flag" ]] && echo "yes (allowed for prereleases)" || echo "no")"
  info "========================================="

  # Dry run mode - preview and exit
  if [[ "$dry_run" == true ]]; then
    preview "========================================="
    preview "Preview: The following would be executed"
    preview "========================================="
    echo "  git tag $git_force_flag $tag"
    echo "  git push $git_force_flag origin $tag"
    preview "========================================="
    preview "This will trigger the GitHub Actions workflow:"
    preview "  1. Create/update tag"
    preview "  2. Run goreleaser"
    preview "     - Build binaries for linux/darwin (amd64/arm64)"
    preview "     - Create GitHub release with artifacts"
    preview "     - Generate changelog from commits"
    if [[ "$release_type" == "stable" ]]; then
      preview "     - Update Homebrew tap (sorinlg/homebrew-tap)"
    else
      preview "     - Update Homebrew dev tap (sorinlg/homebrew-dev-tap)"
    fi
    preview "========================================="
    info "Dry run complete. No changes were made."
    exit 0
  fi

  # Confirm with user
  echo ""
  read -p "Create and push tag $tag? (y/N) " -n 1 -r
  echo ""

  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    warning "Release cancelled by user"
    exit 0
  fi
fi

# Create and push tag
if [[ "$CI_MODE" == "false" ]]; then
  info "Creating tag $tag..."
fi

git tag $git_force_flag $tag

if [[ "$CI_MODE" == "false" ]]; then
  info "Pushing tag to origin..."
fi

git push $git_force_flag origin $tag

# Success message
if [[ "$CI_MODE" == "false" ]]; then
  echo ""
  info "========================================="
  info "✅ Release tag $tag created and pushed"
  info "========================================="
  info "🚀 GitHub Actions will now:"
  info "   1. Run goreleaser"
  info "      - Build binaries for linux/darwin (amd64/arm64)"
  info "      - Create GitHub release with artifacts"
  info "      - Generate changelog from commits"
  if [[ "$release_type" == "stable" ]]; then
    info "      - Update Homebrew tap (sorinlg/homebrew-tap)"
  else
    info "      - Update Homebrew dev tap (sorinlg/homebrew-dev-tap)"
  fi
  info ""
  info "📺 Watch progress:"
  info "   https://github.com/sorinlg/tf-manage2/actions"
  info ""
  info "📦 Release will be available at:"
  info "   https://github.com/sorinlg/tf-manage2/releases/tag/$tag"
  info "========================================="
else
  echo "✅ Tag $tag created and pushed"
fi
