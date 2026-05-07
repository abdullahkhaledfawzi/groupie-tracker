# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o groupie-tracker .

# Run stage — minimal image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/groupie-tracker .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./groupie-tracker"]
