#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

bun run build
scp -r dist crockeo@homeserver.crockeo.net:~/form
ssh crockeo@homeserver.crockeo.net "bash -s" < ./deploy_remote.sh
