FROM golang:1.25 AS builder
WORKDIR /app

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
RUN apt update && apt install -y ca-certificates

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
  echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
  echo 'nobody:x:65534:' > /user/group

# Goreleaser uses the already built binaries, so we just need to copy
COPY ttn-exporter /ttn-exporter

# Final stage: the running container.
FROM scratch AS final

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the first stage.
COPY --from=builder /ttn-exporter /ttn-exporter

EXPOSE      9981
USER        65534:65534
ENTRYPOINT  [ "/ttn-exporter" ]
