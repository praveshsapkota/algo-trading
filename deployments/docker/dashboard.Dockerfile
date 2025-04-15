FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashboard ./cmd/dashboard

# Use a smaller image for the final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/dashboard .

# Copy web assets
COPY --from=builder /app/web/build /root/web/build

# Expose the port
EXPOSE 3000

# Command to run the executable
CMD ["./dashboard"]