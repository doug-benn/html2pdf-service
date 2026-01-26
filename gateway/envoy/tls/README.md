# Envoy TLS certificates

This directory is mounted into the Envoy container at `/etc/envoy/tls`.

For local development, generate a self-signed certificate and private key here.
The `.gitignore` in this folder excludes generated certificates and keys so
private material is not committed.
