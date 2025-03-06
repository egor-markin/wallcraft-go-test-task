##############################################################################
# Stage 1: Builder
##############################################################################

# Use the official Golang image as the base image
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

##############################################################################
# Stage 2: Final runtime image
##############################################################################

FROM scratch

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/myapp .

# Expose the port the application will run on
EXPOSE 8080

# Run the application
CMD ["./myapp"]
