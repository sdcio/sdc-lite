#!/usr/bin/env bash
set -e

REPO="sdcio/config-diff"
INSTALL_DIR="/usr/local/bin"
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names to Go release format
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

echo "📦 Installing config-diff for $OS/$ARCH..."

# Get latest release tag from GitHub API
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -oP '"tag_name":\s*"\K(.*)(?=")')

if [ -z "$LATEST_TAG" ]; then
    echo "❌ Failed to fetch latest release tag"
    exit 1
fi

echo "➡️ Latest version: $LATEST_TAG"

# Download tarball
TARBALL="config-diff_${LATEST_TAG#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$TARBALL"

echo "⬇️ Downloading $URL..."
curl -L "$URL" -o "$TARBALL"

# Extract and install
echo "📂 Extracting..."
tar -xzf "$TARBALL"

echo "🚀 Installing to $INSTALL_DIR..."
sudo mv config-diff "$INSTALL_DIR/config-diff"
sudo chmod +x "$INSTALL_DIR/config-diff"

# Cleanup
rm "$TARBALL"

# Install completions
echo "🔧 Setting up shell completions for $SHELL_NAME..."

case "$SHELL_NAME" in
    bash)
        COMPLETION_PATH="${HOME}/.bash_completion.d"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/config-diff" completion bash > "$COMPLETION_PATH/config-diff"
        echo "source $COMPLETION_PATH/config-diff" >> "${HOME}/.bashrc"
        ;;
    zsh)
        COMPLETION_PATH="${HOME}/.zsh/completions"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/config-diff" completion zsh > "$COMPLETION_PATH/_config-diff"
        echo "fpath=($COMPLETION_PATH \$fpath)" >> "${HOME}/.zshrc"
        echo "autoload -Uz compinit && compinit" >> "${HOME}/.zshrc"
        ;;
    fish)
        COMPLETION_PATH="${HOME}/.config/fish/completions"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/config-diff" completion fish > "$COMPLETION_PATH/config-diff.fish"
        ;;
    *)
        echo "⚠️  Shell completions not set up: unsupported shell ($SHELL_NAME)"
        ;;
esac

echo "✅ config-diff installed successfully!"
echo "ℹ️  Restart your shell or run 'source ~/.bashrc' / 'source ~/.zshrc' to enable completions."
config-diff --version
