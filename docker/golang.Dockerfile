# Gunakan image Golang lengkap (Debian based) bukan Alpine
# Ini lebih stabil dan tidak perlu install compiler C manual
FROM golang:1.24

# Set folder kerja
WORKDIR /app

# Copy source code dari folder src lokal ke dalam container
COPY ./src .

# Download library
RUN go mod download

# Setup Environment
# Kita tetap butuh CGO untuk SQLite
ENV CGO_ENABLED=1
ENV GOOS=linux

# Build paling standar (tanpa flag aneh-aneh)
# Kita build langsung file main di root
RUN go build -o wagoaais .

# Buka port (dokumentasi saja)
EXPOSE 3000

# Jalankan aplikasi langsung
CMD ["./wagoaais", "rest"]