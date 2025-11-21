############################
# STEP 1 build executable binary
############################
FROM golang:1.24-alpine3.20 AS builder
RUN apk update && apk add --no-cache gcc musl-dev gcompat
WORKDIR /app
COPY ./src .

# Fetch dependencies
RUN go mod download

# Build the binary (simple version - no optimization flags)
RUN go build -o whatsapp

#############################
## STEP 2 build a smaller image
#############################
FROM alpine:3.20
RUN apk add --no-cache ffmpeg
WORKDIR /app
# Copy compiled binary from builder
COPY --from=builder /app/whatsapp /app/whatsapp
# Run the binary
ENTRYPOINT ["/app/whatsapp"]

CMD [ "rest" ]