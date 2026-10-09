#!/bin/sh
# Installed as /usr/local/bin/tm-gateway-deploy and run as root through sudo.
# It is the only thing the deploy key can run (see setup-vm.sh). Over SSH:
#   tar -c Caddyfile gateway.env tm-gateway.service | ssh deploy@openmars.app "deploy <version>"
# where <version> is "latest" or a release tag such as v7.
set -euf

IMAGE=ghcr.io/terraforming-mars-ce/terraforming-mars-proxy

# The requested command arrives as one string; split it into words
# shellcheck disable=SC2086
set -- ${1:-}
if [ "$#" -ne 2 ] || [ "$1" != "deploy" ]; then
  echo "usage: deploy <version>" >&2
  exit 2
fi
version="$2"
case "$version" in
  latest | v[0-9]*) ;;
  *)
    echo "invalid version '$version': use latest or a release tag such as v7" >&2
    exit 2
    ;;
esac
case "$version" in
  *[!0-9A-Za-z.-]*)
    echo "invalid version '$version'" >&2
    exit 2
    ;;
esac

staging=$(mktemp -d)
trap 'rm -rf "$staging"' EXIT
tar -x -f - -C "$staging" Caddyfile gateway.env tm-gateway.service

install -m 0644 "$staging/tm-gateway.service" /etc/systemd/system/tm-gateway.service
install -m 0644 "$staging/Caddyfile" /etc/caddy/Caddyfile
install -d -m 0755 /etc/tm-gateway
{
  cat "$staging/gateway.env"
  echo "TM_VERSION=$version"
} > /etc/tm-gateway/gateway.env

docker pull "$IMAGE:$version"
systemctl daemon-reload
systemctl enable tm-gateway.service
systemctl restart tm-gateway.service
systemctl reload caddy.service
docker image prune -f
echo "Deployed gateway $version"
