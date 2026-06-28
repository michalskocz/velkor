#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_succ()  { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_err()   { echo -e "${RED}[ERROR]${NC} $1"; >&2; }
fatal()     { log_err "$1"; exit 1; }

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
    else
        fatal "Script requires root privileges (or installed 'sudo') to install packages."
    fi
fi

detect_pm() {
    if command -v apt-get >/dev/null 2>&1; then
        PM="apt-get"
        PM_UPDATE="$SUDO apt-get update -y"
        PM_INSTALL="$SUDO apt-get install -y"
        PKG_GO="golang"
        PKG_CONTAINER="docker.io"
    elif command -v dnf >/dev/null 2>&1; then
        PM="dnf"
        PM_UPDATE="$SUDO dnf check-update || true"
        PM_INSTALL="$SUDO dnf install -y"
        PKG_GO="golang"
        PKG_CONTAINER="podman"
    elif command -v pacman >/dev/null 2>&1; then
        PM="pacman"
        PM_UPDATE="$SUDO pacman -Sy"
        PM_INSTALL="$SUDO pacman -S --noconfirm"
        PKG_GO="go"
        PKG_CONTAINER="docker"
    else
        fatal "No supported package manager found (apt, dnf, pacman). Install dependencies manually."
    fi
    log_info "Detected package manager: $PM"
}

install_pkg() {
    local pkg_name=$1
    log_info "Installing $pkg_name..."
    $PM_UPDATE >/dev/null 2>&1 || log_warn "Repository update might have failed, attempting to install anyway..."
    $PM_INSTALL "$pkg_name"
    log_succ "Installed $pkg_name."
}

check_dependencies() {
    local pm_detected=false

    if command -v go >/dev/null 2>&1; then
        log_succ "Go is already installed: $(go version)"
    else
        log_warn "'go' is not installed."
        if ! $pm_detected; then detect_pm; pm_detected=true; fi
        install_pkg "$PKG_GO"
    fi

    if command -v docker >/dev/null 2>&1; then
        log_succ "Detected Docker: $(docker --version)"
    elif command -v podman >/dev/null 2>&1; then
        log_succ "Detected Podman: $(podman --version)"
    else
        log_warn "Neither Docker nor Podman is installed."
        if ! $pm_detected; then detect_pm; pm_detected=true; fi
        log_info "Installing default container engine for your distribution ($PKG_CONTAINER)..."
        install_pkg "$PKG_CONTAINER"

        if [ "$PKG_CONTAINER" = "docker" ] || [ "$PKG_CONTAINER" = "docker.io" ]; then
            if command -v systemctl >/dev/null 2>&1; then
                log_info "Attempting to start Docker service..."
                $SUDO systemctl start docker || log_warn "Failed to start docker daemon, please do it manually."
                $SUDO systemctl enable docker || true
            fi
        fi
    fi
}

log_info "Starting environment check..."
check_dependencies

log_info "Building the project..."

if go build -o velkor ./exe; then
    log_succ "Successfully built the 'velkor' executable."
else
    fatal "Error building the 'velkor' project."
fi

log_info "Running application in DEBUG mode..."
./velkor -d DEBUG -c 1
