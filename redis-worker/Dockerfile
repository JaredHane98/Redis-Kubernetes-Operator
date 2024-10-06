# Start from the latest golang base image
FROM golang:1.23.1 AS builder

# Add Maintainer Info
LABEL maintainer="Jared"

# Set the Current Working Directory inside the container
WORKDIR /workspace

# Download Go modules
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download 


# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o manager ./main.go


# Expose port 8080 to the outside world
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .

EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/manager"]