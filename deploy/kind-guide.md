# Optional local kind exercise (not performed by CI)

This guide is for **your laptop only**. Nothing here is deployed from this repository.

1. [Install kind](https://kind.sigs.k8s.io/) and kubectl.
2. `kind create cluster`
3. `docker build -t admission-policy-webhook:local .`
4. `kind load docker-image admission-policy-webhook:local`
5. `./scripts/gen-certs.sh` and create a TLS `Secret` from `certs/tls.crt` and `certs/tls.key`.
6. Apply `deploy/namespace.yaml`, `deploy/deployment.yaml`, `deploy/service.yaml`.
7. Patch `deploy/validating-webhook-configuration.yaml` with `caBundle`, then apply it.

Teardown: `kind delete cluster`
