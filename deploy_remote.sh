#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo "--- Building + deploying backend service"
pushd ~/form
git pull
go build -o form_exe
sudo mv form_exe /opt/form/form_exe
sudo systemctl restart form.service
popd

echo "--- Deploying FE code"
sudo rm -rf /var/www/form
sudo mv ~/form_fe /var/www/form
sudo systemctl reload nginx.service
