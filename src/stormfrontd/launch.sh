#! /usr/bin/env bash

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

export STORMFRONTD_CONFIG_PATH="${HERE}/data/config.json"
export STORMFRONTD_HTTP_PORT="6674"

RUN_DIR=$(dirname $0)

pushd "${RUN_DIR}"
"${HERE}"/stormfrontd
popd
