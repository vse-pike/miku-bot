FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /bot ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bot /bot
ENTRYPOINT ["/bot"]
