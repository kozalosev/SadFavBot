# syntax=docker/dockerfile:1

FROM golang:1.18-alpine as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY base/* ./base/
COPY handlers/* ./handlers/
COPY storage/* ./storage/
COPY wizard/* ./wizard/

RUN go build -o /sadFavBot github.com/kozalosev/SadFavBot

FROM alpine:3
COPY --from=builder sadFavBot /bin/sadFavBot
ENTRYPOINT [ "/bin/sadFavBot" ]

LABEL org.opencontainers.image.source=https://github.com/kozalosev/SadFavBot
LABEL org.opencontainers.image.description="Favorites bot for Telegram"
LABEL org.opencontainers.image.licenses=MIT
