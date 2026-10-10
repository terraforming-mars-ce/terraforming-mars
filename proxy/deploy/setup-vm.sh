#!/bin/sh
# One-time setup of the gateway VM (Debian or Ubuntu), run as root:
#   scp -r proxy/deploy root@<vm>:/tmp/openmars-gateway
#   ssh root@<vm> sh /tmp/openmars-gateway/setup-vm.sh "ssh-ed25519 AAAA... github-deploy"
# Installs Docker and Caddy, and creates a `deploy` user whose SSH key can only
# run the root-owned deploy script. Then run the Deploy workflow for the gateway.
set -eu

public_key="${1:?usage: setup-vm.sh \"<deploy public key>\"}"
here=$(cd "$(dirname "$0")" && pwd)

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
if ! command -v caddy >/dev/null 2>&1; then
  apt-get update
  apt-get install -y caddy
fi

if ! id deploy >/dev/null 2>&1; then
  useradd --create-home --shell /bin/sh deploy
fi

install -o root -g root -m 0755 "$here/deploy.sh" /usr/local/bin/openmars-gateway-deploy

sudoers=$(mktemp)
echo "deploy ALL=(root) NOPASSWD: /usr/local/bin/openmars-gateway-deploy" > "$sudoers"
visudo -cf "$sudoers"
install -o root -g root -m 0440 "$sudoers" /etc/sudoers.d/openmars-gateway-deploy
rm -f "$sudoers"

install -d -o deploy -g deploy -m 0700 /home/deploy/.ssh
printf 'restrict,command="sudo /usr/local/bin/openmars-gateway-deploy \\"$SSH_ORIGINAL_COMMAND\\"" %s\n' \
  "$public_key" > /home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh/authorized_keys
chmod 0600 /home/deploy/.ssh/authorized_keys

echo "Gateway VM ready. Store this line as the GATEWAY_KNOWN_HOSTS secret:"
printf 'openmars.app %s\n' "$(cut -d' ' -f1,2 /etc/ssh/ssh_host_ed25519_key.pub)"
