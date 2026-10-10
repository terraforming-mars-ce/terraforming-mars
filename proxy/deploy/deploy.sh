#!/bin/sh
# Installed as /usr/local/bin/openmars-gateway-deploy and run as root through sudo.
# It is the only thing the deploy key can run (see setup-vm.sh). Over SSH:
#   tar -c Caddyfile gateway.env openmars-gateway.service | ssh deploy@openmars.app "deploy <version>"
# where <version> is "latest" or a release tag such as v7.0.1.
set -euf

IMAGE=ghcr.io/openmars-app/openmars-proxy

# The requested command arrives as one string; split it into words
# shellcheck disable=SC2086
set -- ${1:-}
if [ "$#" -ne 2 ] || [ "$1" != "deploy" ]; then
  echo "usage: deploy <version>" >&2
  exit 2
fi
version="$2"
if ! printf '%s' "$version" | grep -Eqx 'latest|v[0-9]+\.[0-9]+\.[0-9]+'; then
  echo "invalid version '$version': use latest or a release tag such as v7.0.1" >&2
  exit 2
fi

staging=$(mktemp -d)
trap 'rm -rf "$staging"' EXIT
tar -x -f - -C "$staging" Caddyfile gateway.env openmars-gateway.service

install -m 0644 "$staging/openmars-gateway.service" /etc/systemd/system/openmars-gateway.service
install -m 0644 "$staging/Caddyfile" /etc/caddy/Caddyfile
install -d -m 0755 /etc/openmars-gateway
{
  cat "$staging/gateway.env"
  echo "OPENMARS_VERSION=$version"
} > /etc/openmars-gateway/gateway.env

docker pull "$IMAGE:$version"
systemctl daemon-reload
systemctl enable openmars-gateway.service
systemctl restart openmars-gateway.service
systemctl reload caddy.service
docker image prune -f
echo "Deployed gateway $version"
