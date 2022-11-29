FROM --platform=$BUILDPLATFORM golang:1.19 AS builder
WORKDIR /app

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
RUN apt update && apt install -y ca-certificates

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
  echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
  echo 'nobody:x:65534:' > /user/group

# Set the environment variables for the go command:
# * CGO_ENABLED=0 to build a statically-linked executable
ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH

COPY ./ ./

RUN go build -o /ttn_exporter ttn_exporter.go

# Final stage: the running container.
FROM scratch AS final
LABEL org.opencontainers.image.source=https://github.com/juusujanar/ttn-exporter
LABEL org.opencontainers.image.description="Prometheus exporter for The Things Network"
LABEL org.opencontainers.image.licenses=MIT

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the first stage.
COPY --from=builder /ttn_exporter /ttn_exporter

EXPOSE      9981
USER        65534:65534
ENTRYPOINT  [ "/ttn_exporter" ]
