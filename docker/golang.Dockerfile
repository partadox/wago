############################
# STEP 1: BUILDER (Gunakan Debian base, bukan Alpine)
############################
FROM golang:1.24 AS builder

WORKDIR /app

# Copy source code
COPY ./src .

# Download dependencies
RUN go mod download

# Setup Environment
# CGO_ENABLED=1 wajib untuk SQLite
ENV CGO_ENABLED=1
ENV GOOS=linux

# Build command
# Gunakan -p 1 untuk hemat RAM
# Hapus flag -s -w sementara untuk melihat jika ada error detail
RUN go build -p 1 -v -o wagoaais .

#############################
## STEP 2: RUNNER (Tetap Alpine supaya kecil)
#############################
FROM alpine:3.20

# Install dependencies runtime
RUN apk add --no-cache ffmpeg tzdata ca-certificates

ENV TZ=UTC
WORKDIR /app

# Copy hasil build dari stage builder
COPY --from=builder /app/wagoaais /app/wagoaais

# Run
ENTRYPOINT ["/app/wagoaais"]
CMD [ "rest" ]