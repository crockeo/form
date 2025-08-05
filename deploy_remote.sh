#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

sudo rm -rf /var/www/form
sudo mv ~/form /var/www/form
sudo systemctl reload nginx.service
