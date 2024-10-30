# Use the official Golang image to create a build artifact.
FROM golang:1.17 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o netlify-ddns-script .

# Start a new stage from scratch
FROM alpine:latest

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/netlify-ddns-script /usr/local/bin/netlify-ddns-script

# Command to run the executable
CMD ["netlify-ddns-script"]