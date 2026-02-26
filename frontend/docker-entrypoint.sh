#!/bin/sh
set -e

# Generate runtime config from environment variables
# This allows the same Docker image to be deployed to different environments

CONFIG_FILE="/usr/share/nginx/html/runtime-config.js"

# Use API_URL env var, fallback to empty string (will use default in app)
API_URL="${API_URL:-}"

# Generate the runtime config file
cat > "$CONFIG_FILE" << EOF
window.__RUNTIME_CONFIG__ = {
  apiUrl: '${API_URL}',
};
EOF

echo "Runtime config generated at $CONFIG_FILE"
echo "  API_URL: ${API_URL:-'(not set, will use default)'}"

# Execute the CMD (nginx)
exec "$@"
