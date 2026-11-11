# Interview notes — admission-policy-webhook

Independent study artifact for Habib Noor’s portfolio (security / platform / infrastructure). Not employer-affiliated.

## Modules

| Module | Responsibility |
|--------|----------------|
| `cmd/webhook` | Process entry, TLS listener, env-based config |
| `internal/server` | `/healthz`, `/readyz`, `/validate`, bounded HTTP |
| `internal/admission` | `AdmissionReview` parse/emit, UID echo, scope routing |
| `internal/extract` | JSON paths to `PodSpec` per supported kind |
| `internal/policy` | Image digest, security context, resource quantities |

## Tradeoffs

1. **Fail closed (`failurePolicy: Fail`)** — stops risky workloads during outages; can halt deploys.
2. **Allow unsupported kinds** — reduces false positives; CRDs with embedded pods stay ungoverned unless extended.
3. **JSON partial structs vs. scheme runtime** — lighter deps; less validation of irrelevant fields.
4. **Strict lowercase digest hex** — simple parser; rejects uppercase (could normalize later).
5. **Required resources for init** — aligns with batch/startup cost governance; may frustrate minimal demos.

## 30 Q&A

1. **What problem does this solve?** Enforces baseline pod hardening at admission time.
2. **Why admission vs. OPA/Gatekeeper?** Small, readable Go policy without a sidecar template engine; good for demonstrating core K8s hooks.
3. **What API version?** `admission.k8s.io/v1` only for reviews.
4. **Does it mutate objects?** No — validating webhook only.
5. **How is UID preserved?** Copied from `request.uid` to `response.uid`.
6. **Max body size?** 3 MiB (`MaxAdmissionBodyBytes`).
7. **What happens on malformed JSON?** HTTP 400 admission denial with message (no UID if request never parsed).
8. **Unsupported kind?** Allowed without inspection (`extract.errUnsupportedKind`).
9. **Malformed Deployment?** Denied — cannot extract pod template.
10. **CronJob path?** `spec.jobTemplate.spec.template.spec`.
11. **Image rule?** `@sha256:` + 64 hex; no digest denied; `latest` tag denied even with digest.
12. **Does digest prove security?** No — only immutability of content reference.
13. **runAsNonRoot inheritance?** Container setting wins; else pod; unset → false → deny.
14. **runAsUser 0?** Deny when explicitly set to 0 at effective level.
15. **Ephemeral containers resources?** Optional; validate when partially set.
16. **CPU quantities?** `resource.MustParse` / `Quantity.Cmp` semantics in policy tests.
17. **Limits below requests?** Denied for regular/init; same when ephemeral specifies both.
18. **DELETE operations?** Allowed (no policy hook).
19. **Timeout?** Example manifest uses 5s; server has 10s read/write timeouts.
20. **TLS keys in git?** No — `scripts/gen-certs.sh`, `.gitignore` for `certs/`.
21. **caBundle?** Base64 PEM of CA signing serving cert for apiserver trust.
22. **sideEffects?** `None` — webhook does not change cluster state or kubelet.
23. **Health checks?** `/healthz` and `/readyz` return `{"status":"ok"}`.
24. **Docker base?** Distroless nonroot.
25. **Race tests?** CI runs `go test -race ./...`.
26. **Why not deploy here?** Scope control; no paid cloud; reproducible local/CI only.
27. **What would you add next?** PodSecurity standards, verify image signature, dry-run metrics, owner refs.
28. **How to extend kinds?** Add case in `extract.PodSpecFromObject` + webhook rule.
29. **Single replica risk?** Example deployment uses 1; production needs HA and cert rotation.
30. **FailurePolicy alternative?** `Ignore` improves availability but silently skips policy.

## Failure experiments (safe, local)

| Experiment | Command / action | Expected |
|------------|------------------|----------|
| Unpinned image | Send review with `nginx:1.25` | `allowed: false` |
| Latest + digest | `nginx:latest@sha256:…` | Denied |
| Oversized body | POST > 3 MiB to `/validate` | HTTP 400 |
| Bad JSON | POST `{` | HTTP 400 |
| Missing UID | Omit `request.uid` | HTTP 400 denial |
| Wrong review API | `apiVersion: v1beta1` | Denied with message |
| ConfigMap create | Unsupported kind | `allowed: true` |
| Empty Deployment template | Missing `spec.template` | Denied |
| Init missing RAM limit | Policy test case | Denied |
| Pod compliant | Example manifest | Allowed in tests |

## Honesty notes

Generic AI assistance was used for scaffolding and documentation structure; tests and policy behavior were validated with `go test -race ./...` and kubeconform on manifests in CI. No live cluster run is claimed in this repo.
