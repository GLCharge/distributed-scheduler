# Use an official GoLang runtime as the base image
FROM golang:1.20-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Verify dependencies
RUN go mod verify

# Copy the source code from the current directory to the working directory inside the container
COPY . .

# install go swaggo, used to generate swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12
# build docs
RUN swag init -g handlers/doc.go

# Build the Go application
RUN go build -o bin/manager cmd/manager/main.go

# Use an official Alpine Linux runtime as a base image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage to the current stage
COPY --from=builder /app/bin/manager /app/manager

# set command to run when starting the container
CMD ["/app/manager"]