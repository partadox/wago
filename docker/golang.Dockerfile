############################
# STEP 1 build executable binary
############################
FROM golang:1.24-alpine3.20 AS builder

# Gunakan build-base daripada gcc musl-dev terpisah (lebih aman untuk Alpine)
RUN apk add --no-cache build-base git

WORKDIR /app

# Copy semua isi folder src ke /app
COPY ./src .

RUN go mod download

# SET ENV VARS
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Build dengan flag -v (verbose) agar error muncul di log Coolify
# Tanda titik (.) di akhir sangat penting
RUN go build -v -a -ldflags="-w -s" -o wagoaais .

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