#!/bin/bash

# Check if required environment variables are set
if [ -z "$APPLE_DEVELOPER_ID_APPLICATION" ] || [ -z "$APPLE_DEVELOPER_ID_INSTALLER" ] || [ -z "$APPLE_ID" ] || [ -z "$APPLE_ID_PASSWORD" ] || [ -z "$KEYCHAIN_PROFILE" ] || [ -z "$KEYCHAIN_NAME" ]; then
  echo "Error: One or more required environment variables are not set."
  echo "Required variables: APPLE_DEVELOPER_ID_APPLICATION, APPLE_DEVELOPER_ID_INSTALLER, APPLE_ID, APPLE_ID_PASSWORD, KEYCHAIN_PROFILE, KEYCHAIN_NAME"
  exit 1
fi

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

# Unlock the keychain
security unlock-keychain -p "$APPLE_KEYCHAIN_PASSWORD" $KEYCHAIN_NAME
# Prepare directories for the installer
mkdir -p $BASE_DIR
mkdir -p $PACKAGE_DIR
mkdir -p $RESOURCE_DIR
mkdir -p $ROOT_DIR/Library/LaunchDaemons
mkdir -p $ROOT_DIR/opt/mw-agent/
mkdir -p $ROOT_DIR/etc/mw-agent/
mkdir -p $SCRIPTS_DIR
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

echo "Signing the mw-agent binary with hardened runtime"
sudo codesign --sign "$APPLE_DEVELOPER_ID_APPLICATION" --options runtime --timestamp "$ROOT_DIR/opt/mw-agent/mw-agent"
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

# Sign the installer package
echo "Signing the inner installer package"
sudo codesign --sign "$APPLE_DEVELOPER_ID_APPLICATION" --options runtime --timestamp "$PACKAGE_DIR/middleware_agent.pkg"

# Check if productsign command failed
if [ $? -ne 0 ]; then
  echo "Error: productsign command failed"
  exit 1
fi

# Create and sign the installer package
echo "Signing the product installer package"
sudo productbuild --distribution $BASE_DIR/distribution.xml \
             --package-path $PACKAGE_DIR \
             --resources $RESOURCE_DIR \
             --sign "$APPLE_DEVELOPER_ID_INSTALLER" \
             --keychain $KEYCHAIN_NAME \
             $INSTALLER_DIR/$INSTALLER_NAME

# Check if productbuild command failed
if [ $? -ne 0 ]; then
  echo "Error: productbuild command failed"
  exit 1
fi

echo "Code Signing the final installer package"
sudo codesign --sign "$APPLE_DEVELOPER_ID_APPLICATION" --options runtime --timestamp $INSTALLER_DIR/$INSTALLER_NAME

# Check if productbuild command failed
if [ $? -ne 0 ]; then
  echo "Error: codesign command failed for the final installer package"
  exit 1
fi

# Verify the package signature
pkgutil --check-signature $INSTALLER_DIR/$INSTALLER_NAME
codesign -dv --verbose=4 $INSTALLER_DIR/$INSTALLER_NAME

echo "Notarizing the installer package, team id: $APPLE_DEVELOPER_TEAM_ID"
xcrun notarytool store-credentials "$KEYCHAIN_PROFILE" --apple-id $APPLE_ID --password $APPLE_ID_PASSWORD --team-id "$APPLE_DEVELOPER_TEAM_ID"

echo "Submitting the installer package for notarization"
sudo xcrun notarytool submit $INSTALLER_DIR/$INSTALLER_NAME --keychain-profile "$KEYCHAIN_PROFILE" --wait

# Check if notarytool command failed
if [ $? -ne 0 ]; then
  echo "Error: notarytool command failed"
  exit 1
fi

echo "Stapling the notarization ticket to the installer package"
sudo xcrun stapler staple $INSTALLER_DIR/$INSTALLER_NAME

# Check if stapler command failed
if [ $? -ne 0 ]; then
  echo "Error: stapler command failed"
  exit 1
fi
echo "Installer package created at $INSTALLER_DIR/$INSTALLER_NAME"
