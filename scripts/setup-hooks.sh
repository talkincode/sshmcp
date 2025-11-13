#!/bin/bash
# Setup Git hooks for this repository

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
HOOKS_DIR="$PROJECT_ROOT/.githooks"
GIT_HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

echo "🔧 Setting up Git hooks..."
echo ""

# Check if .git directory exists
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo "❌ Error: Not a git repository"
    exit 1
fi

# Configure git to use .githooks directory
echo "📝 Configuring git hooks path..."
git config core.hooksPath "$HOOKS_DIR"

echo "✅ Git hooks configured successfully!"
echo ""
echo "Installed hooks:"
for hook in "$HOOKS_DIR"/*; do
    if [ -f "$hook" ]; then
        hook_name=$(basename "$hook")
        echo "  • $hook_name"
    fi
done

echo ""
echo "📚 Hook descriptions:"
echo "  • pre-commit:  Runs tests, formatting, and linting before commit"
echo "  • commit-msg:  Enforces conventional commit message format"
echo "  • pre-push:    Runs comprehensive checks before pushing"
echo ""
echo "💡 Tips:"
echo "  • Skip hooks temporarily: git commit --no-verify"
echo "  • Test a hook manually: .githooks/pre-commit"
echo "  • Install golangci-lint for better linting: brew install golangci-lint"
echo ""
echo "🎉 Setup complete!"
