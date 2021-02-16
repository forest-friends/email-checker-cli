FROM golang:alpine as build
RUN apk update && apk add --no-cache git
WORKDIR /build

COPY go.* ./
RUN go mod download

COPY . .
RUN go build -o bin/email-checker-cli cmd/email-checker-cli.go

FROM alpine
WORKDIR /app
COPY --from=build /build/bin/* /app/
ENTRYPOINT ["/app/email-checker-cli"]