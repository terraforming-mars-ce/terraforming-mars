#!/bin/sh
set -e

# Write the server list the loader reads from OPENMARS_SERVERS, e.g.
#   OPENMARS_SERVERS="rackaracka=https://terraforming-mars.rackaracka.net,saffronbun=https://tm.saffronbun.com"
# The order is the order players see in the picker.

SERVERS_JS="${SERVERS_JS:-/usr/share/nginx/html/servers.js}"

if [ -z "${OPENMARS_SERVERS:-}" ]; then
  echo "OPENMARS_SERVERS is required (alias=https://server,...)" >&2
  exit 1
fi

entries=""
old_ifs="$IFS"
IFS=','
for pair in $OPENMARS_SERVERS; do
  alias="${pair%%=*}"
  url="${pair#*=}"
  case "$alias" in
    '' | *[!a-z0-9-]*)
      echo "Invalid server alias '$alias': use lowercase a-z, 0-9 and -" >&2
      exit 1
      ;;
  esac
  case "$url" in
    http://* | https://*) ;;
    *)
      echo "Invalid URL for '$alias': $url" >&2
      exit 1
      ;;
  esac
  case "$url" in
    *[\"\\\ \<\>]*)
      echo "Invalid characters in URL for '$alias': $url" >&2
      exit 1
      ;;
  esac
  entries="$entries{\"alias\":\"$alias\",\"url\":\"$url\"},"
done
IFS="$old_ifs"

echo "window.__OPENMARS_SERVERS__ = [${entries%,}];" > "$SERVERS_JS"
echo "Servers written to $SERVERS_JS: $OPENMARS_SERVERS"

exec "$@"
