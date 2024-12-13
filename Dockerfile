# Start from the latest golang base image
FROM golang:1.23.3-alpine3.20 AS builder
# Add Maintainer Info
LABEL maintainer="Eduardo Alonso <eduardo.alonso@disashop.com>"

#Install upx to reduze binary size
ARG upx_version=4.2.4
RUN apk add xz curl && \
    curl -Ls https://github.com/upx/upx/releases/download/v${upx_version}/upx-${upx_version}-amd64_linux.tar.xz -o - | tar xvJf - -C /tmp  && \
    cp /tmp/upx-${upx_version}-amd64_linux/upx /usr/local/bin/  && \
    chmod +x /usr/local/bin/upx  && \
    apk del xz  && \
    rm -rf /var/lib/apt/lists/*

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies optimizing cache for dependencies dependencies -> https://github.com/montanaflynn/golang-docker-cache
RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

#Add Cache mounts
ENV GOCACHE=/root/.cache/go-build

# Build the Go app
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -gcflags=all="-l -B" -ldflags="-w -s" -installsuffix cgo -o main .

RUN upx --ultra-brute -qq main && \
    upx -t main

# Start from scratch image to optimized disk
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

COPY --from=builder /app/main /app
COPY config.yaml /app/

# Command to run the executable
ENTRYPOINT ["/app/main"]
