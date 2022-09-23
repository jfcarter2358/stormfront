#! /usr/bin/env bash

rm /bin/stormfront
rm /bin/stormfrontd

rm -r /var/stormfront

if [[ "$(ps --no-headers -o comm 1)" == "systemd" ]]; then
    systemctl stop stormfront.service
    rm /etc/systemd/system/stormfront.service
else
    service stormfront stop
    rm /etc/init.d/stormfront
fi

