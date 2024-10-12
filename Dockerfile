# Use official Golang image as the base
FROM golang:1.19 as builder

# Set the current working directory inside the container
WORKDIR /app

# Copy everything from the current directory to /app in the container
COPY . .

# Download and install Chrome & ChromeDriver
RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    xvfb \
    && wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt install ./google-chrome-stable_current_amd64.deb -y \
    && rm google-chrome-stable_current_amd64.deb \
    && wget https://chromedriver.storage.googleapis.com/114.0.5735.90/chromedriver_linux64.zip \
    && unzip chromedriver_linux64.zip \
    && mv chromedriver /usr/local/bin/ \
    && rm chromedriver_linux64.zip

# Build Go application
RUN go mod tidy && go build -o /bin-collector-symbio main.go

# Final image for execution
FROM debian:buster-slim

# Install Chrome & ChromeDriver dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libx11-6 \
    libx11-xcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxi6 \
    libxtst6 \
    libnss3 \
    libglib2.0-0 \
    libxrandr2 \
    libasound2 \
    libpangocairo-1.0-0 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libgtk-3-0 \
    libgbm1 \
    libxshmfence1 \
    libdrm2 \
    libnspr4 \
    && apt-get clean

# Copy from the builder image
COPY --from=builder /usr/local/bin/chromedriver /usr/local/bin/chromedriver
COPY --from=builder /app /app

# Set working directory and environment
WORKDIR /app
ENV DISPLAY=:99

# Command to run the app
CMD ["./bin-collector-symbio"]
