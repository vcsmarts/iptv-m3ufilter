#!/usr/bin/env bash
set -e

GO_VERSION="${GO_VERSION:-$(go list -m -f '{{.Version}}' go@latest)}"

GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TAR}"

# Get the currently installed version of Go
INSTALLED_VERSION=$(go version 2>/dev/null | awk '{print $3}' | sed 's/^go//' || echo "")
version_greater_equal() {
    # Compare two version numbers
    # Usage: version_greater_equal version1 version2
    # Returns 0 if version1 >= version2, 1 otherwise
    dpkg --compare-versions "$1" ge "$2"
}

confirm_upgrade() {
    echo -n "Do you want to proceed with the upgrade? (y/N): "
    read -r response
    case "$response" in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            echo "Upgrade cancelled."
            exit 0
            ;;
    esac
}

if [ -z "$INSTALLED_VERSION" ]; then
    echo "Go is not installed. Installing version ${GO_VERSION}..."
    confirm_upgrade
elif version_greater_equal "$INSTALLED_VERSION" "$GO_VERSION"; then
    echo "You already have Go version ${INSTALLED_VERSION} installed, which is >= ${GO_VERSION}. No upgrade needed."
    exit 0
else
    echo "Installed Go version: ${INSTALLED_VERSION}"
    echo "Target Go version: ${GO_VERSION}"
    echo "This will upgrade Go from version ${INSTALLED_VERSION} to ${GO_VERSION}."
    confirm_upgrade
fi

# Create temporary directory for download
TMP_DIR=$(mktemp -d)
# clean up when the bashscript exits.
trap "rm -rf ${TMP_DIR}" EXIT

echo "Downloading Go version ${GO_VERSION}..."
wget -q -O "${TMP_DIR}/${GO_TAR}" -- ${GO_URL}

echo "Removing old version of Go..."
sudo rm -rf /usr/local/go

echo "Installing Go version ${GO_VERSION}..."
sudo tar -C /usr/local -xzf "${TMP_DIR}/${GO_TAR}"

# Update PATH if needed
if ! grep -q "/usr/local/go/bin" ~/.bash_profile; then
    echo "Updating PATH in ~/.bash_profile..."
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile
    source ~/.bash_profile
fi

echo "Go has been successfully upgraded to version:"
go version