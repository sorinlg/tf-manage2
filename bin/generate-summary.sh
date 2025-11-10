#!/usr/bin/env bash
# Generate GitHub Actions workflow summary
# Used by release workflow to create GITHUB_STEP_SUMMARY
#
# Environment Variables:
#   SHOULD_RELEASE - true/false
#   VERSION - version tag (e.g., v1.2.3)
#   IS_PRERELEASE - true/false
#   GITHUB_REF_NAME - branch name
#   GITHUB_REPOSITORY - owner/repo

set -euo pipefail

# Validate required environment variables
: "${SHOULD_RELEASE:?SHOULD_RELEASE environment variable is required}"
: "${VERSION:?VERSION environment variable is required}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY environment variable is required}"

# Generate summary based on whether release was created
if [[ "$SHOULD_RELEASE" == "true" ]]; then
  echo "## 🚀 Release Complete"
  echo ""
  echo "**Version**: \`$VERSION\`"

  if [[ "${IS_PRERELEASE:-false}" == "true" ]]; then
    echo "**Type**: 🚧 Prerelease"
  else
    echo "**Type**: ✅ Stable"
  fi

  if [[ -n "${GITHUB_REF_NAME:-}" ]]; then
    echo "**Branch**: \`$GITHUB_REF_NAME\`"
  fi

  echo ""
  echo "### Installation"

  if [[ "${IS_PRERELEASE:-false}" == "true" ]]; then
    cat <<EOF
\`\`\`bash
# Latest prerelease from Homebrew:
brew install sorinlg/dev-tap/tf-manage2-dev

# Specific prerelease version:
brew install sorinlg/dev-tap/tf-manage2-dev@$VERSION
\`\`\`
EOF
  else
    cat <<EOF
\`\`\`bash
# Latest stable from Homebrew (recommended):
brew install sorinlg/tap/tf-manage2

# Specific version:
brew install sorinlg/tap/tf-manage2@$VERSION
\`\`\`
EOF
  fi

  echo ""
  echo "### Artifacts"
  echo "- Binaries for linux/darwin (amd64/arm64)"
  echo "- Checksums and archives"
  echo "- Shell completion scripts (bash/zsh)"
  echo ""
  echo "🔗 [View Release](https://github.com/$GITHUB_REPOSITORY/releases/tag/$VERSION)"
else
  echo "## ⏭️ Release Skipped"
  echo ""
  echo "Release \`$VERSION\` already exists"
  echo ""
  echo "🔗 [View Existing Release](https://github.com/$GITHUB_REPOSITORY/releases/tag/$VERSION)"
fi
