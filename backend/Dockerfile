# Build the Go server.
FROM golang:1.27-alpine AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/db-library ./cmd/db-library

# Small runtime image. The UI is the frontend service; this image is the API.
FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 65532 app \
    && adduser -S -D -H -u 65532 -G app app
WORKDIR /app
COPY --from=go-builder /out/db-library /app/db-library
RUN mkdir -p /etc/db-library && chown -R app:app /app /etc/db-library
USER app
ENV DBLIB_CONFIG=/etc/db-library/databases.json
EXPOSE 8080
ENTRYPOINT ["/app/db-library"]
