# Use an official Golang image as the base image
FROM golang:1.23

# Install dependencies for running Chrome
RUN apt-get update && apt-get install -y wget unzip curl && \
    wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
    apt-get install -y ./google-chrome-stable_current_amd64.deb && \
    rm google-chrome-stable_current_amd64.deb

# Set the working directory
WORKDIR /app

# Copy the existing chromedriver from the local directory
COPY ./chromedriver-linux64/chromedriver /usr/local/bin/chromedriver

# Make sure the driver is executable
RUN chmod +x /usr/local/bin/chromedriver

# Copy the Go source code and other files
COPY . .

# Build the Go application
RUN go mod tidy && go build -o /app/bin-collector-symbio-linux main.go

# Expose the web server port
EXPOSE 8080

# Run the Go application
CMD ["/app/bin-collector-symbio-linux"]
