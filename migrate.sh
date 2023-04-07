#!/bin/bash
rm -rf migrate-files migrate.linux-amd64.tar.gz
wget https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz
mkdir migrate-files
tar -xvf migrate.linux-amd64.tar.gz -C migrate-files

chmod +x ./migrate-files/migrate
./migrate-files/migrate -database $DB_URL -path ./migrations up

rm -rf migrate.linux-amd64.tar.gz migrate-files