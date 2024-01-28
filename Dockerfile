ARG ARCH=
FROM ${ARCH}golang:1.21-bullseye as build-server

# Initialization
RUN mkdir -p /app
WORKDIR /app

# Dependencies
COPY go.mod ./
COPY go.sum .

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg
COPY assets ./assets

RUN go build -o valetudo-telegram-bot ./cmd/valetudo-telegram-bot/main.go

FROM ${ARCH}debian:bullseye-slim

# Options
ENV TELEGRAM_BOT_TOKEN ""
ENV TELEGRAM_CHAT_IDS ""
ENV VALETUDO_URL ""
ENV TELEGRAM_DEBUG false

# Copy build results
WORKDIR /app

# This has to be done to be able to use https
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update
RUN apt install ca-certificates -y

COPY --from=build-server /app/valetudo-telegram-bot ./valetudo-telegram-bot

# Start the application
ENTRYPOINT [ "./valetudo-telegram-bot" ]
