############################
# STEP 1 build executable binary
############################
FROM golang:1.24-alpine3.20 AS builder

# Install build tools
RUN apk add --no-cache build-base git

WORKDIR /app

# Copy source code
COPY ./src .

# 1. DEBUG: Tampilkan struktur file untuk memastikan main.go ada di root /app
# Lihat log ini nanti di Coolify jika masih error
RUN echo "=== FILE STRUCTURE ===" && ls -R /app && echo "======================"

# Download dependencies
RUN go mod download

# Set Environment
ENV CGO_ENABLED=1
ENV GOOS=linux

# 2. BUILD FIX:
# Tambahkan "-p 1" untuk hemat RAM (Mencegah OOM Killer di VPS kecil)
# Hapus "-a" jika ingin memanfaatkan cache build (opsional, tapi hemat resource)
RUN go build -p 1 -v -ldflags="-w -s" -o wagoaais .

#############################
## STEP 2 build a smaller image
#############################
FROM alpine:3.20
RUN apk add --no-cache ffmpeg tzdata ca-certificates

ENV TZ=UTC
WORKDIR /app

COPY --from=builder /app/wagoaais /app/wagoaais

ENTRYPOINT ["/app/wagoaais"]
CMD [ "rest" ]