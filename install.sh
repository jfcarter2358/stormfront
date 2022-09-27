#! /usr/bin/env bash

set -eo pipefail

RELEASE_BODY=$(curl -s -H "Accept: application/vnd.github+json" https://api.github.com/repos/jfcarter2358/stormfront/releases/latest)

echo "Downloading assets..."
# Download the daemon and the CLI
for ASSET in $(echo "${RELEASE_BODY}" | jq -r '.assets[] | @base64'); do
    _jq() {
        echo ${ASSET} | base64 --decode | jq -r ${1}
    }

    NAME=$(_jq '.name')
    URL=$(_jq '.url')

    echo "Downloading ${NAME} from ${URL}"

    curl -sL -H "Accept: application/octet-stream" "${URL}" > /bin/"${NAME}"

    chmod +x /bin/"${NAME}"
done
echo "Done!"

echo "Creating data directory at /var/stormfront..."
sudo mkdir -p /var/stormfront
echo "Done!"

if [[ "$(ps --no-headers -o comm 1)" == "systemd" ]]; then
    echo "Writing systemd files..."
    # Write out our service file for stormfront
    cat << EOF > /etc/systemd/system/stormfront.service
server1:/etc/systemd/system # cat /etc/systemd/system/stormfront.service
[Unit]
Description=Stormfrontd systemd service file.

[Service]
ExecStart=/bin/stormfrontd

[Install]
WantedBy=multi-user.target
EOF

    echo "Done!"

    echo "Starting service..."
    # Reload systemd
    systemctl daemon-reload

    # Enable stormfront
    systemctl enable stormfront.service

    # Start stormfront
    systemctl start stormfront.service

    echo "Done!"
else
    echo "Writing init.d files..."
    # Write out our init script
    cat << EOF > /etc/init.d/stormfront

#!/bin/sh
. /lib/lsb/init-functions

case "\$1" in
    start)
        log_daemon_msg "Starting stormfront daemon" "stormfrontd" || true
        if start-stop-daemon --start --quiet --oknodo --chuid 0:0 --pidfile /run/stormfront.pid --exec /bin/stormfrontd; then
            log_end_msg 0 || true
        else
            log_end_msg 1 || true
        fi
        ;;
    stop)
        log_daemon_msg "Stopping stormfront daemon" "stormfrontd" || true
        if start-stop-daemon --stop --quiet --oknodo --pidfile /run/stormfront.pid --exec /bin/stormfrontd; then
            log_end_msg 0 || true
        else
            log_end_msg 1 || true
        fi
        ;;

    restart)
        log_daemon_msg "Restarting OpenBSD Secure Shell server" "stormfrontd" || true
        start-stop-daemon --stop --quiet --oknodo --retry 30 --pidfile /run/stormfront.pid --exec /bin/stormfrontd
        if start-stop-daemon --start --quiet --oknodo --chuid 0:0 --pidfile /run/stormfront.pid --exec /bin/stormfrontd; then
            log_end_msg 0 || true
        else
            log_end_msg 1 || true
        fi
        ;;

    status)
        status_of_proc -p /run/stormfrontd.pid /bin/stormfrontd stormfrontd && exit 0 || exit \$?
        ;;

    *)
        log_action_msg "Usage: /etc/init.d/stormfront {start|stop|restart|status}" || true
        exit 1
esac

exit 0
EOF

    chmod +x /etc/init.d/stormfront

    echo "Done!"

    echo "Starting service..."
    # Start stormfront
    service stormfront start &
    echo "Done!"
fi
