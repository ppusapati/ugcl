# Start from official Go image
FROM golang:1.24.1

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./
RUN go mod tidy

# Copy the rest of the app
COPY . .

# Build the Go app
RUN go build -o server .

# Tell Cloud Run which port to expose
ENV PORT=8080

# Start the server
CMD ["./server"]
