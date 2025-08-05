#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo "--- Building + Copying frontend"
bun run build
scp -r dist crockeo@homeserver.crockeo.net:~/form_fe

echo "--- Executing remote deploy"
ssh crockeo@homeserver.crockeo.net "bash -s" < ./deploy_remote.sh
