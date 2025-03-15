# Use the official Golang image as the base image
FROM alpine:3.14


# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY ./bin/go-cache_linux ./

RUN chmod +x go-cache_linux

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./go-cache_linux"]