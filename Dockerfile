# syntax=docker/dockerfile:1

FROM golang:1.18-alpine as builder
WORKDIR /app

# Create an unprivileged user
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY base/* ./base/
COPY handlers/ ./handlers/
COPY storage/* ./storage/
COPY wizard/* ./wizard/
COPY settings/* ./settings/

# Build without debugging info
RUN go build -ldflags="-w -s" -o /sadFavBot github.com/kozalosev/SadFavBot

FROM alpine:3
COPY --from=builder sadFavBot /bin/sadFavBot
# Import the user and group files from the builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# Use the unprivileged user
USER appuser:appuser
ENTRYPOINT [ "/bin/sadFavBot" ]

LABEL org.opencontainers.image.source=https://github.com/kozalosev/SadFavBot
LABEL org.opencontainers.image.description="Favorites bot for Telegram"
LABEL org.opencontainers.image.licenses=MIT
