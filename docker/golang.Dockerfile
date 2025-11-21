############################
# STEP 1 build executable binary
############################
# Pastikan tag versi go valid (misal 1.24 atau 1.23)
FROM golang:1.24-alpine3.20 AS builder

# Install dependencies untuk CGO
RUN apk update && apk add --no-cache gcc musl-dev gcompat

# Ubah WORKDIR ke /app agar konsisten dengan output dan stage final
WORKDIR /app

# Copy source code
# Asumsi: folder 'src' berisi go.mod dan main.go
COPY ./src .

# Fetch dependencies.
RUN go mod download

# Build the binary
# Hapus path absolut '/app/' karena kita sudah di dalam '/app'
# Tambahkan CGO_ENABLED=1 jika aplikasi butuh gcc (karena ada musl-dev)
ENV CGO_ENABLED=1
RUN go build -a -ldflags="-w -s" -o wagoaais

#############################
## STEP 2 build a smaller image
#############################
FROM alpine:3.20

RUN apk add --no-cache ffmpeg tzdata

ENV TZ=UTC

WORKDIR /app

# Copy compiled from builder.
# Karena di builder workdir sudah /app, copy dari /app/wagoaais
COPY --from=builder /app/wagoaais /app/wagoaais

# Run the binary.
ENTRYPOINT ["/app/wagoaais"]

CMD [ "rest" ]