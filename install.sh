#!/usr/bin/env bash
set -e

${USE_SUDO:="true"}
REPO="sdcio/sdc-lite"
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

echo "📦 Installing sdc-lite for $OS/$ARCH..."

# Get latest release tag from GitHub API
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -oE '"tag_name":\s*"[^"]+"' | cut -d '"' -f 4)

if [ -z "$LATEST_TAG" ]; then
    echo "❌ Failed to fetch latest release tag"
    exit 1
fi

echo "➡️ Latest version: $LATEST_TAG"

# Download tarball
TARBALL="sdc-lite_${LATEST_TAG#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$TARBALL"

echo "⬇️ Downloading $URL..."
curl -L "$URL" -o "$TARBALL"

# Extract and install
echo "📂 Extracting..."
tar -xzf "$TARBALL"

echo "🚀 Installing to $INSTALL_DIR..."
runAsRoot mv sdc-lite "$INSTALL_DIR/sdc-lite"
runAsRoot chmod +x "$INSTALL_DIR/sdc-lite"

# Cleanup
rm "$TARBALL"

# Install completions
echo "🔧 Setting up shell completions for $SHELL_NAME..."

case "$SHELL_NAME" in
    bash)
        COMPLETION_PATH="${HOME}/.bash_completion.d"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/sdc-lite" completion bash > "$COMPLETION_PATH/sdc-lite"
        # Only append the source line if it doesn't already exist
        if ! grep -Fxq "source $COMPLETION_PATH/sdc-lite" "${HOME}/.bashrc"; then
            echo "source $COMPLETION_PATH/sdc-lite" >> "${HOME}/.bashrc"
        fi
        ;;
    zsh)
        COMPLETION_PATH="${HOME}/.zsh/completions"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/sdc-lite" completion zsh > "$COMPLETION_PATH/_sdc-lite"
        if ! grep -Fxq "fpath=($COMPLETION_PATH \$fpath)" "${HOME}/.zshrc"; then
            echo "fpath=($COMPLETION_PATH \$fpath)" >> "${HOME}/.zshrc"
            echo "autoload -Uz compinit && compinit" >> "${HOME}/.zshrc"
        fi
        ;;
    fish)
        COMPLETION_PATH="${HOME}/.config/fish/completions"
        mkdir -p "$COMPLETION_PATH"
        "$INSTALL_DIR/sdc-lite" completion fish > "$COMPLETION_PATH/sdc-lite.fish"
        ;;
    *)
        echo "⚠️  Shell completions not set up: unsupported shell ($SHELL_NAME)"
        ;;
esac

echo "✅ sdc-lite installed successfully!"
echo "ℹ️  Restart your shell or run 'source ~/.bashrc' / 'source ~/.zshrc' to enable completions."
sdc-lite --version

