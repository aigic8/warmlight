FROM golang:alpine

RUN mkdir /app

WORKDIR /app

# copying files and dirs
ADD ./cmd /app/cmd
ADD ./internal /app/internal
ADD ./pkg /app/pkg
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

# generate sqlc files
RUN chmod +x /app/internal/db/sqlc-alpine.sh
WORKDIR /app/internal/db
RUN ./sqlc-alpine.sh
WORKDIR /app

RUN go build /app/cmd/warmlight/warmlight.go

# change port if you have changed port in warmlight.toml
EXPOSE 443
# using a raw string instead of array because of https://stackoverflow.com/a/24194340/19217994
CMD /app/warmlight -c ${CONFIG_FILE}
