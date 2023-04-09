# WarmLight
WarmLight is a Telegram bot for creating and using databases of quotes.

## Installation
- [What you need](#what-you-need)
- [Creating a bot in Telegram](#creating-a-bot-in-telegram)
- [Generating SSL certificate files](#generating-ssl-certificate-files)
- [Configuration](#configuration)
- [Using with docker and docker-compose (Recommended)](#using-with-docker-and-docker-compose-recommended)
- [Using without docker](#using-without-docker)
- [Running Migrations](#running-migrations)

### What you need
To be able to use this bot, you need:
- a VPS to run your bot
- a domain name (technically it is not necessary, so in future versions it may not be needed)

### Creating a bot in Telegram
First you need to create a bot using [Bot Father](t.me/botfather). You need to get your bot token and make save it somewhere safe. Make sure also to activate inline bot functionality for bot in Bot Father. 

### Generating SSL certificate files
You need to generate SSL certificate and private key file and pass them to the bot configuration. For that purpose you can use [CertBot](https://certbot.eff.org/) or [acme.sh](acme.sh)

### Configuration
You need to copy `warmlight.sample.toml` (or rename it) to `warmlight.toml` and change it's content as it suites you. This file is written in [Toml](https://github.com/toml-lang/toml) language. 

Also make sure to set `isDev` to `false` if you are running in production.

### Using with docker and docker-compose (Recommended)
You need to [install Docker](https://docs.docker.com/engine/install/) and docker-compose. You can use this `docker-compose.yml` file and modify it to your need:
```yaml
version: "3.9"
services:
  warmlight:
    build: ./warmlight # or wherever you have cloned warmlight
    restart: always
    ports:
      - 443:443 # make sure to change the container port if you've changed it in 'warmlight.toml'
    links:
      - db:db
    volumes:
      - ./certs:/app/certs # directory of certificates
      - ./log/warmlight:/app/log # directory you want to put logs in
  db:
    image: postgres:alpine
    restart: always
    ports:
      -  "127.0.0.1:5432:5432"
    container_name: postgres-db
    environment:
      - POSTGRES_PASSWORD=1234 # database password, CHANGE IT and make sure it matches password in 'warmlight.toml'
```

Then you can run command to start postgresql.
```bash 
docker compose up -d db
``` 
Now you need to directly connect to database using 
```bash 
docker exec -it postgres-db psql -U postgres
```
and create a database. (you can use `Ctrl-D` to exit postgres prompt)
```sql
CREATE DATABASE warmlight; -- change 'warmlight' to whatever you want the database name to be
```
Afterwards, you need [run migrations](#running-migrations) and finally run the application using
```bash
docker compose up -d warmlight
```

### Using without docker
This part of documentation is still not written.

### Running migrations
Migrations are handled using [migrate](https://github.com/golang-migrate/migrate/). For simplicity, I've created a bash script. After creating a database, you can run migrations using this command:
```bash
chmod +x ./migrate.sh # make migrate.sh executable
DB_URL=postgresql://postgres:password@localhost:5432/warmlight?sslmode=disable ./migrate.sh
```




