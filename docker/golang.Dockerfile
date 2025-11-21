############################
# STEP 1 build executable binary
############################
FROM golang:1.24-alpine3.20 AS builder
RUN apk update && apk add --no-cache gcc musl-dev gcompat
WORKDIR /whatsapp
COPY ./src .

# Show Go version
RUN go version

# Fetch dependencies
RUN go mod download

# Verify dependencies
RUN go mod verify

# Build the binary with verbose output (NO optimization flags to see errors clearly)
RUN set -x && go build -v -o /app/whatsapp

# Verify binary was created
RUN ls -lh /app/whatsapp

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