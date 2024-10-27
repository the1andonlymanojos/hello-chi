FROM golang:1.23-alpine3.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod  \
    go mod download
COPY . .
RUN go build -o hello-chi .


FROM alpine:3.20.3
ENV TEMP_DIR=/app/tmp
ENV STOR_DIR=/app/storage
ENV REDIS_HOST=192.168.68.205:6379
ENV REDIS_PASSWORD=""
ENV REDIS_DB=0
ENV PORT=3000
COPY --from=builder /app/hello-chi /app/hello-chi

RUN mkdir -p $TEMP_DIR $STOR_DIR
EXPOSE 3000
ENTRYPOINT ["/app/hello-chi"]


