#!/bin/sh


set -e

apk add gcompat

rm -rf sqlc-files sqlc_1.14.0_linux_amd64.tar.gz
wget https://downloads.sqlc.dev/sqlc_1.14.0_linux_amd64.tar.gz
mkdir sqlc-files
tar -xvf sqlc_1.14.0_linux_amd64.tar.gz -C sqlc-files 

chmod +x ./sqlc-files/sqlc
./sqlc-files/sqlc generate

rm -rf sqlc-files sqlc_1.14.0_linux_amd64.tar.gz
