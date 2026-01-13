FROM golang:1.24-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/recsys ./cmd/app

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=build /out/recsys /app/recsys
COPY web /app/web

EXPOSE 8080
USER app
ENTRYPOINT ["/app/recsys"]
