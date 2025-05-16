FROM golang:1.24.2

# Set working directory in the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Expose the port your app runs on (change if needed)
EXPOSE 8080

# Run the Go app
CMD ["go", "run", "."]