FROM golang:alpine

RUN mkdir /app

WORKDIR /app

ADD ./cmd /app/cmd
ADD ./internal /app/internal
ADD ./migrations /app/migrations
ADD ./pkg /app/pkg

COPY ./migrate.sh /app/migrate.sh
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY ./warmlight.toml /app/warmlight.toml

RUN go build /app/cmd/warmlight/warmlight.go

RUN chmod +x /app/migrate.sh
RUN /app/migrate.sh

CMD [ "/app/cmd/warmlight/warmlight" ]