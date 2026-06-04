# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go ./
COPY json2swagger.go ./
COPY schemaParser.go ./
RUN go get gopkg.in/yaml.v3 github.com/go-faster/jx
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /swaggerscripts .

# Runtime stage
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /swaggerscripts /usr/local/bin/swaggerscripts
RUN mkdir output && mkdir input
ENTRYPOINT ["swaggerscripts"]
CMD ["-help"]
