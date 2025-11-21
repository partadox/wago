############################
# STEP 1 build executable binary
############################
FROM golang:1.24-alpine3.20 AS builder
RUN apk update && apk add --no-cache gcc musl-dev gcompat
WORKDIR /whatsapp
COPY ./src .

# Fetch dependencies.
RUN go mod download
# Verify go.mod and show version
RUN go version && go mod verify || true
# Build the binary with optimizations (show errors)
RUN go build -v -a -ldflags="-w -s" -o /app/whatsapp 2>&1 | tee /tmp/build.log || (cat /tmp/build.log && exit 1)

#############################
## STEP 2 build a smaller image
#############################
FROM alpine:3.20
RUN apk add --no-cache ffmpeg
WORKDIR /app
# Copy compiled from builder.
COPY --from=builder /app/whatsapp /app/whatsapp
# Run the binary.
ENTRYPOINT ["/app/whatsapp"]

CMD [ "rest" ]