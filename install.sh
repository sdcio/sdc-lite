#!/usr/bin/env bash
set -e

${USE_SUDO:="true"}
REPO="sdcio/config-diff"
INSTALL_DIR="/usr/local/bin"
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
SHELL_NAME=$(basename "$SHELL")


# runs the given command as root (detects if we are root already)
runAsRoot() {
    local CMD="$*"

    if [ "$EUID" -ne 0 ] && [ "$USE_SUDO" = "true" ]; then
        CMD="sudo $CMD"
    fi

    $CMD
}

download() {
    if type "curl" &>/dev/null; then
        curl -L "$URL" -o "$TARBALL"
    elif type "wget" &>/dev/null; then
        wget "$URL" -O "$TARBALL"
    fi
}

# verifySupported checks that the os/arch combination is supported
verifySupported() {
    if ! type "curl" &>/dev/null && ! type "wget" &>/dev/null; then
        echo "Either curl or wget is required"
        exit 1
    fi
}



verifySupported

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
runAsRoot mv config-diff "$INSTALL_DIR/config-diff"
runAsRoot chmod +x "$INSTALL_DIR/config-diff"

# Cleanup
rm "$TARBALL"

# Install completions
echo "🔧 Setting up shell completions for $SHELL_NAME..."

case "$SHELL_NAME" in
    bash)
        COMPLETION_PATH="${HOME}/.bash_completion.d"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/config-diff" completion bash > "$COMPLETION_PATH/config-diff"
        # Only append the source line if it doesn't already exist
        if ! grep -Fxq "source $COMPLETION_PATH/config-diff" "${HOME}/.bashrc"; then
            echo "source $COMPLETION_PATH/config-diff" >> "${HOME}/.bashrc"
        fi
        ;;
    zsh)
        COMPLETION_PATH="${HOME}/.zsh/completions"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/config-diff" completion zsh > "$COMPLETION_PATH/_config-diff"
        if ! grep -Fxq "fpath=($COMPLETION_PATH \$fpath)" "${HOME}/.zshrc"; then
            echo "fpath=($COMPLETION_PATH \$fpath)" >> "${HOME}/.zshrc"
            echo "autoload -Uz compinit && compinit" >> "${HOME}/.zshrc"
        fi
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

