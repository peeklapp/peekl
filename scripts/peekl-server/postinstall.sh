#!/bin/sh
set -e

SERVICE_NAME="peekl-server"

case "$1" in
    configure)
        # Reload systemd in case the unit file changed
        if [ -d /run/systemd/system ]; then
            systemctl daemon-reload || true

            if systemctl is-active --quiet "$SERVICE_NAME"; then
                echo "Restarting $SERVICE_NAME..."
                systemctl restart "$SERVICE_NAME"
            else
                echo "Starting $SERVICE_NAME..."
                systemctl start "$SERVICE_NAME" || true
            fi

            # Make sure it's enabled on boot, if that's desired
            systemctl enable "$SERVICE_NAME" >/dev/null 2>&1 || true
        fi
        ;;
esac

exit 0
