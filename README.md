# WarmLight
WarmLight is a Telegram bot for creating and using databases of quotes.

## Usage

```text
WarmLight Bot
This bot is created to help you saving your quotes. You can add a quote simply by sending a message in this format:
We should forget about small efficiencies, say about 97% of the time: premature optimization is the root of all evil.
sources: Donald Ervin Knuth
#programming #optimization 
There are several important commands in this bot:
/getsources you can search your sources, it will return results and you can view their info and also edit them. For example:
/getsources Animal Farm
will search for a source with name of "Animal Farm". Also, you can use source type specifier to search more specifically for source. For example:
/getsources Animal Farm @book
Will only search for books with name "Animal Farm".
/setactivesource will activate a source for certain amount of time. During that time period, every quote you send will automatically be added to that source. For example:
/setactivesource Animal Farm, 20
Will set source "Animal Farm" as active source for "20 minutes". The time period is optional, for example:
/setactivesource Animal Farm
This command will set source "Animal Farm" as active source for default timeout (60 minutes)
/getoutputs will show you your outputs. Outputs are Telegram channels when you send a new quote, your quotes will be forwarded to there. You can activate and deactivate your outputs with this command.
/getlibtoken and /setlibtoken are used to share a quote library between multiple accounts. The owner of the library will use command /getlibtoken to get his library token. The second account will use command /setlibtoken to set library token received by the owner.
```

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
      - ./config/warmlight/config:/app/config
    enviroment:
      -CONFIG_PATH=/app/config/warmlight.toml # path of config file
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



