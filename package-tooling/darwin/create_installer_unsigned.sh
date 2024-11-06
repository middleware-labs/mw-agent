#!/bin/bash

# Set variables
REPO_ROOT=$(git rev-parse --show-toplevel)
BASE_DIR="/tmp/mw-agent-pkg"
PACKAGE_DIR="$BASE_DIR/package"
INSTALLER_DIR="$REPO_ROOT/build"
ROOT_DIR="$BASE_DIR/root"
SCRIPTS_DIR="$BASE_DIR/scripts"
RESOURCE_DIR="$BASE_DIR/resources"
IDENTIFIER="io.middleware.mw-agent"
INSTALLER_NAME="mw-macos-agent-setup-${ARCH}.pkg"
RELEASE_VERSION=$1

# Prepare directories for the installer
mkdir -p $BASE_DIR
mkdir -p $PACKAGE_DIR
mkdir -p $RESOURCE_DIR
mkdir -p $ROOT_DIR/Library/LaunchDaemons
mkdir -p $ROOT_DIR/opt/mw-agent/
mkdir -p $ROOT_DIR/etc/mw-agent/
mkdir -p $SCRIPTS_DIR

# Copy required files
sudo cp $REPO_ROOT/build/mw-host-agent $ROOT_DIR/opt/mw-agent/mw-agent
sudo cp $REPO_ROOT/package-tooling/agent-config.yaml.sample $ROOT_DIR/etc/mw-agent/agent-config.yaml
sudo cp -r $REPO_ROOT/package-tooling/darwin/resources/* $RESOURCE_DIR
sudo cp -r $REPO_ROOT/package-tooling/darwin/preinstall $SCRIPTS_DIR
sudo chmod +x $SCRIPTS_DIR/preinstall

sudo cp -r $REPO_ROOT/package-tooling/darwin/prompt_user.applescript $SCRIPTS_DIR

sudo cp -r $REPO_ROOT/package-tooling/darwin/postinstall $SCRIPTS_DIR
sudo chmod +x $SCRIPTS_DIR/postinstall

sudo cp -r $REPO_ROOT/package-tooling/darwin/uninstall.sh $ROOT_DIR/opt/mw-agent/uninstall.sh
sudo chmod +x $ROOT_DIR/opt/mw-agent/uninstall.sh

sudo cp -r $REPO_ROOT/package-tooling/darwin/io.middleware.mw-agent.plist $ROOT_DIR/Library/LaunchDaemons/

sudo cp -r $REPO_ROOT/package-tooling/darwin/distribution.xml $BASE_DIR/distribution.xml

# Create component package for the installer
sudo pkgbuild --root $ROOT_DIR \
         --identifier $IDENTIFIER \
         --version $RELEASE_VERSION \
         --install-location / \
         --scripts $SCRIPTS_DIR \
         $PACKAGE_DIR/middleware_agent.pkg

# Check if pkgbuild command failed
if [ $? -ne 0 ]; then
  echo "Error: pkgbuild command failed"
  exit 1
fi

# Create the installer package
sudo productbuild --distribution $BASE_DIR/distribution.xml \
             --package-path $PACKAGE_DIR \
             --resources $RESOURCE_DIR \
             $INSTALLER_DIR/$INSTALLER_NAME

# Check if productbuild command failed
if [ $? -ne 0 ]; then
  echo "Error: productbuild command failed"
  exit 1
fi

echo "Installer package created at $INSTALLER_DIR/$INSTALLER_NAME"
