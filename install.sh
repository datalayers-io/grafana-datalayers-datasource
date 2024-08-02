#!/usr/bin/env bash
set -e

# Make sure grafana cli is installed
which grafana > /dev/null || (echo "Grafana has not been installed in the server, install it first."; exit 1)

# Parse shell args
while getopts h:u:p: opt
do
  case "${opt}" in
    h) HOST=${OPTARG};;
    u) USERNAME=${OPTARG};;
    p) PASSWORD=${OPTARG};;
  esac
done

# Make grafana custom.ini to current directory
tee custom.ini > /dev/null << EOF
[paths]
plugins = $PWD/plugins
provisioning = $PWD/provisioning

[plugins]
allow_loading_unsigned_plugins = datalayersio-datasource
EOF
echo "Generated custom.ini"

# Make grafana provisioning file
mkdir -p provisioning/datasources
mkdir -p provisioning/dashboards
mkdir -p provisioning/plugins
mkdir -p provisioning/notifiers
mkdir -p provisioning/alerting
tee provisioning/datasources/datalayersio-datasource.yaml > /dev/null << EOF
apiVersion: 1
datasources:
  - name: Datalayers
    type: datalayersio-datasource
    orgId: 1
    url: http://$HOST
    jsonData:
      host: "$HOST"
      selectedAuthType: "username/password"
      username: "$USERNAME"
    secureJsonData:
      password: "$PASSWORD"
    version: 1
    editable: true
EOF
echo "Generated provisioning/datasources/datalayersio-datasource.yaml"


# Set repository information
REPO_OWNER="datalayers-io"
REPO_NAME="grafana-datalayers-datasource"
API_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases"

# Fetch all release information
ALL_RELEASES=$(curl -s $API_URL)

# Check if a valid release list is found
if [[ $ALL_RELEASES == *"Not Found"* ]]; then
    echo "Releases not found for the $REPO_NAME repository."
    exit 1
fi

# Extract the latest pre-release tag from the release information
LATEST_RELEASE_TAG=$(echo "$ALL_RELEASES" | grep -Eo '"tag_name": "[^"]*' | sed -E 's/"tag_name": "//' | head -n 1)
VERSION=$(echo "$LATEST_RELEASE_TAG" | sed 's/^v//')
if [[ -z "$LATEST_RELEASE_TAG" ]]; then
    echo "No pre-release found for the $REPO_NAME repository."
    exit 1
fi

echo "The latest release tag of $REPO_NAME is: $LATEST_RELEASE_TAG"


# Install plugin by grafana cli
echo "Installing grafana plugin: datalayersio-datasource"
grafana cli --pluginsDir "$PWD/plugins" --pluginUrl https://github.com/datalayers-io/grafana-datalayers-datasource/releases/download/$LATEST_RELEASE_TAG/datalayersio-datasource-$VERSION.zip plugins install datalayersio-datasource

echo -e "
Please run the following command at your grafana homepath.\n
grafana server --config $PWD/custom.ini"