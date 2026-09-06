FROM golang:1.23-alpine AS build
WORKDIR /src
COPY miku-bot/go.mod miku-bot/go.sum ./
RUN go mod download
COPY miku-bot/cmd ./cmd
COPY miku-bot/internal ./internal
RUN CGO_ENABLED=0 go build -o /bot ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bot /bot
ENTRYPOINT ["/bot"]
