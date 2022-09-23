#! /usr/bin/env bash

echo "Elevating to root user for uninstall..."
sudo su
echo "Done!"

echo "Removing binary files..."
rm /bin/stormfront
rm /bin/stormfrontd
echo "Done!"

echo "Removing stormfront data..."
rm -r /var/stormfront
echo "Done!"

if [[ "$(ps --no-headers -o comm 1)" == "systemd" ]]; then
    echo "Stopping service and removing systemd files..."
    systemctl stop stormfront.service
    rm /etc/systemd/system/stormfront.service
    echo "Done!"
else
    echo "Stopping service and removing init.d files..."
    service stormfront stop
    rm /etc/init.d/stormfront
    rm /run/stormfrontd.pid
    echo "Done!"
fi

