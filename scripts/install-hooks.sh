#!/bin/bash
# Install git hooks for Nivo project

set -e

echo "Installing git hooks..."

# Get the root directory of the git repository
GIT_ROOT=$(git rev-parse --show-toplevel)
HOOKS_DIR="$GIT_ROOT/.git/hooks"
SCRIPTS_HOOKS_DIR="$GIT_ROOT/scripts/hooks"

# Check if hooks directory exists
if [ ! -d "$HOOKS_DIR" ]; then
    echo "Error: .git/hooks directory not found"
    echo "Are you in a git repository?"
    exit 1
fi

# Install pre-commit hook
if [ -f "$SCRIPTS_HOOKS_DIR/pre-commit" ]; then
    echo "Installing pre-commit hook..."
    cp "$SCRIPTS_HOOKS_DIR/pre-commit" "$HOOKS_DIR/pre-commit"
    chmod +x "$HOOKS_DIR/pre-commit"
    echo "✅ pre-commit hook installed"
else
    echo "⚠️  pre-commit hook script not found at $SCRIPTS_HOOKS_DIR/pre-commit"
fi

# Future hooks can be added here
# Example:
# if [ -f "$SCRIPTS_HOOKS_DIR/pre-push" ]; then
#     echo "Installing pre-push hook..."
#     cp "$SCRIPTS_HOOKS_DIR/pre-push" "$HOOKS_DIR/pre-push"
#     chmod +x "$HOOKS_DIR/pre-push"
#     echo "✅ pre-push hook installed"
# fi

echo ""
echo "✅ Git hooks installation complete!"
echo ""
echo "Installed hooks:"
ls -la "$HOOKS_DIR" | grep -v sample | grep -E '^-rwx' || echo "  (none found)"
echo ""
echo "To bypass hooks temporarily, use: git commit --no-verify"
