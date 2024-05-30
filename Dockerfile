ARG GO_VERSION=1.22.1

FROM golang:${GO_VERSION}-alpine AS builder
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

# install git since it's not in the alpine image by default
RUN apk add --no-cache ca-certificates git

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /workspace

COPY . .

RUN go mod download

# Build
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN GOOS=$(echo $TARGETPLATFORM | cut -f1 -d '/') \
    GOARCH=$(echo $TARGETPLATFORM | cut -f2 -d '/') \
    CGO_ENABLED=0 go build -o /go/bin/exporter

# Final stage
FROM scratch AS final
WORKDIR /
COPY --from=builder /etc/passwd /etc/passwd
USER appuser

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/exporter /exporter

EXPOSE 8080
ENTRYPOINT ["/exporter"]
