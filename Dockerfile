# Use the official Node.js image as the base image
FROM node:14 AS build

# Set the working directory in the container
WORKDIR /app

# Copy package.json and package-lock.json to the working directory
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy the entire client directory to the working directory
COPY client/ ./

# Build the React app
RUN npm run build

# Use the official Go image as the base image for the production build
FROM golang:1.17 AS production

# Set the working directory in the container
WORKDIR /app

# Copy go.mod and go.sum to the working directory
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

# Copy the entire server directory to the working directory
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o file-convert-stream

# Use a multi-stage build to create a lightweight production image
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates ffmpeg

# Set the working directory in the container
WORKDIR /app

# Copy the built Go binary from the production build stage
COPY --from=production /app/file-convert-stream ./

# Set the command to run the Go binary
CMD ["./file-convert-stream"]
