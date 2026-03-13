# Use an official Go runtime as a parent image
FROM golang:1.24.4
# Set the working directory to /go/src/app
WORKDIR /go/src/app

# Install Orchestrion
# RUN go install github.com/DataDog/orchestrion@v1.5.0

# Copy the Go module files
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN orchestrion pin && go mod download

# Copy the current directory contents into the container at /go/src/app
COPY . .

# Build the Go app
RUN go mod tidy
RUN go install -v ./...
# Build with Orchestrion
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 orchestrion go build -tags appsec -o main cmd/server/main.go

#Build without Orchestrion
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/server/main.go

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
