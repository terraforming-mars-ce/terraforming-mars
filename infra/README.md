# infra

Docker Compose files for running the game locally or on a server. For full deployment docs (environment variables, reverse proxy setup, Kubernetes), see [docs/install](../docs/install/).

## Files

- `docker-compose.yml` -- builds images from source
- `docker-compose.prod.yml` -- pulls pre-built images from ghcr.io
- `.env.example` -- environment variable template
- `auto-deploy.sh` -- cron script that checks for new images and redeploys

## Quick start

```bash
cp .env.example .env
# Edit .env with your values (all optional -- game works without them)

# Build from source
docker compose up -d

# Or use pre-built images
docker compose -f docker-compose.prod.yml up -d
```

Game is at `http://localhost:3000`.

## Auto-deploy

The `auto-deploy.sh` script compares local and remote image digests and redeploys if there's a newer version. Set it up as a cron job:

```bash
crontab -e
# Add: */5 * * * * /path/to/infra/auto-deploy.sh
```

See [CRON_SETUP.md](CRON_SETUP.md) for details.

## Bug reporting

Bug reporting is optional. To enable it, fill in the GitHub App and Claude env vars in `.env`. The `GITHUB_PRIVATE_KEY_FILE` variable should point to your GitHub App private key on the host -- it gets bind-mounted into the container as a read-only file.

If you skip these, the game runs fine but the in-game bug report button shows "not available".
