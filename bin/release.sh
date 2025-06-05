#!/usr/bin/env bash

# This script is used to publish (pre)releases to GitHub

set -euo pipefail

info() {
  echo -e "\033[1;34m[INFO]\033[0m $@"
}
error() {
  echo -e "\033[1;31m[ERROR]\033[0m $@" >&2
}
warning() {
  echo -e "\033[1;33m[WARNING]\033[0m $@"
}

# Parse command line arguments
dry_run=false
while [[ $# -gt 0 ]]; do
  case $1 in
    --dry-run)
      dry_run=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [--dry-run] [--help]"
      echo ""
      echo "Options:"
      echo "  --dry-run    Show what would be done without actually doing it"
      echo "  --help       Show this help message"
      exit 0
      ;;
    *)
      error "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Get release type based on branch name
branch=$(git rev-parse --abbrev-ref HEAD)
if [[ "$branch" == "main" ]]; then
  release_type="stable"
elif [[ "$branch" == "develop" ]]; then
  release_type="prerelease"
else
  error "Unsupported branch: $branch"
  exit 1
fi

# Ensure svu is available
if ! command -v svu &> /dev/null; then
  error "svu command not found. Please install svu to use this script."
  exit 1
fi

# Ensure git is available
if ! command -v git &> /dev/null; then
  error "git command not found. Please install git to use this script."
  exit 1
fi

# Ensure goreleaser is available
if ! command -v goreleaser &> /dev/null; then
  error "goreleaser command not found. Please install goreleaser to use this script."
  exit 1
fi

# Ensure the working directory is clean
if ! git diff-index --quiet HEAD --; then
  error "Working directory is not clean. Please commit or stash your changes before releasing."
  exit 1
fi

# Ensure the current branch is up to date, but only if it's not a prerelease
if [[ "$release_type" != "prerelease" ]]; then
  # Check if we're missing remote commits
  if ! git merge-base --is-ancestor origin/"$branch" HEAD; then
    error "Current branch is missing commits from origin/$branch. Please pull the latest changes."
    exit 1
  fi
fi

# Check if we have unpushed commits (warn but don't fail)
if ! git merge-base --is-ancestor HEAD origin/"$branch"; then
  warning "Current branch has unpushed commits ahead of origin/$branch."
fi

# Validate GoReleaser configuration
info "Validating GoReleaser configuration..."
if ! goreleaser check; then
  error "GoReleaser configuration validation failed. Please fix the configuration before releasing."
  exit 1
fi
info "GoReleaser configuration validation passed."

####################################################
# Determine the command to use based on release type
####################################################
# Set the svu command based on the release type
svu_command="svu next"
if [[ "$release_type" == "prerelease" ]]; then
  svu_command="svu prerelease --pre-release=rc"
fi

# Only force push if it's a prerelease
git_force_flag=""
if [[ "$release_type" == "prerelease" ]]; then
  git_force_flag="-f"
fi

# Create a new tag and push it
tag=$(eval $svu_command)
if [[ -z "$tag" ]]; then
  error "Failed to determine the next version tag."
  exit 1
fi

info "Source branch: $branch"
info "Release type: $release_type"
info "Force push allowed: $([[ "$git_force_flag" ]] && echo "yes" || echo "no")"
info "Creating tag: $tag"

if [[ "$dry_run" == true ]]; then
  warning "DRY RUN MODE: The following commands would be executed:"
  echo "  git push $git_force_flag origin $branch"
  echo "  git tag $git_force_flag $tag"
  echo "  git push $git_force_flag origin $tag"
  info "Dry run complete. No changes were made."
else
  # Push the branch
  git push $git_force_flag origin "$branch"
  # Create the tag
  git tag $git_force_flag $tag
  git push $git_force_flag origin $tag
  info "Release $tag created and pushed successfully!"
fi
