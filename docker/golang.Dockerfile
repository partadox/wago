# Ganti ke versi 1.23 yang sudah pasti stabil
FROM golang:1.23

WORKDIR /app

# Copy source
COPY ./src .

# DEBUG 1: Tampilkan isi folder dan isi go.mod di log
# Supaya kita tahu file-nya benar-benar ada
RUN echo "=== LIST FILES ===" && ls -R && echo "=== CONTENT GO.MOD ===" && cat go.mod

# Download modules
RUN go mod download

# DEBUG 2: Coba build TANPA CGO (Pure Go).
# Jika ini BERHASIL, berarti masalah Anda adalah library SQLite/CGO yang berat.
# Jika ini GAGAL, berarti masalah struktur folder/code.
ENV CGO_ENABLED=0
ENV GOOS=linux

# Kita pakai flag -x untuk melihat detail step compiler
RUN go build -x -o wagoaais .

EXPOSE 3000

CMD ["./wagoaais", "rest"]