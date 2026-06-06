#!/bin/bash
# Run once on the VPS as root to prepare the deployment environment.
set -e

# Directories
mkdir -p /opt/izoblichi/{backend,frontend/build,data}

# Install Bun if not present
if ! command -v bun &>/dev/null; then
    curl -fsSL https://bun.sh/install | bash
    # Make bun available system-wide
    ln -sf /root/.bun/bin/bun /usr/local/bin/bun
fi

# Install systemd services
cp "$(dirname "$0")/izoblichi-backend.service" /etc/systemd/system/
cp "$(dirname "$0")/izoblichi-frontend.service" /etc/systemd/system/

systemctl daemon-reload
systemctl enable izoblichi-backend izoblichi-frontend

echo ""
echo "Setup complete."
echo "Next steps:"
echo "  1. Copy procurement.duckdb to /opt/izoblichi/data/procurement.duckdb"
echo "  2. Add the GitHub Actions secrets (see README or deploy.yml comments)"
echo "  3. Push to main — the pipeline will build and start both services"
