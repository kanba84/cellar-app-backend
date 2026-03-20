# Buiild stage
FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cellar-app

# Deployment stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/cellar-app .
EXPOSE 8443
CMD ["./cellar-app"]
