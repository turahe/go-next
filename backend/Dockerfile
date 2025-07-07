# --- Base builder ---
FROM golang:1.24-alpine AS base
WORKDIR /app
COPY . .
RUN go mod download

# --- Swag docs ---
FROM base AS swaggen
RUN go install github.com/swaggo/swag/cmd/swag@latest && /go/bin/swag init

# --- Development stage ---
FROM swaggen AS dev
RUN go install github.com/cosmtrek/air@latest
RUN /go/bin/swag init
EXPOSE 8080
CMD ["/go/bin/air", "server"]

# --- Staging stage (with debug info, no air) ---
FROM swaggen AS staging
RUN /go/bin/swag init
RUN go build -o main .
EXPOSE 8080
CMD ["./main", "server"]

# --- Production stage (minimal, static binary) ---
FROM swaggen AS prod-builder
RUN /go/bin/swag init
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

FROM alpine:3.20 AS prod
WORKDIR /app
COPY --from=prod-builder /app/main ./main
COPY --from=prod-builder /app/docs ./docs
EXPOSE 8080
CMD ["./main", "server"] 