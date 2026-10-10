# Kubernetes

Example manifests are in this directory. Adjust namespaces, hostnames, and TLS config to match your cluster.

## Secrets

Two secrets to create. Neither should be checked into git.

**Claude OAuth token:**

```bash
kubectl create secret generic openmars-claude \
  --namespace <namespace> \
  --from-literal=oauth-token=<your-token>
```

Get the token by running `claude setup-token` on any machine with Claude Code installed.

**GitHub App private key:**

```bash
kubectl create secret generic openmars-github-app \
  --namespace <namespace> \
  --from-file=private-key=/path/to/your/private-key.pem
```

## Deploying

Apply the manifests:

```bash
kubectl apply -f app-deployment.yaml
kubectl apply -f app-service.yaml
kubectl apply -f ingress.yaml
```

Or all at once:

```bash
kubectl apply -f .
```

## How secrets are handled

The private key gets mounted as a read-only file inside the container, not passed as an env var. Kubernetes mounts secrets as tmpfs, so the key only exists in memory on the node.

The Claude OAuth token is passed as a regular env var sourced from a secret. This is fine -- it's a bearer token, same as any API key in a pod.

## Ingress

The example ingress uses nginx-ingress and sends every path to the one service. The 3600s timeouts matter -- without them, the default 60s timeout will kill WebSocket connections during idle moments in a game.

If you're using a different ingress controller, make sure it supports WebSocket upgrades and has configurable read/send timeouts.
