// Package policy implements fail-closed validation for Pod workloads.
//
// # Image policy (regular, init, and ephemeral containers)
//
//   - Every container image MUST include a SHA256 content digest: @sha256: followed by exactly 64 lowercase hexadecimal characters.
//   - Images without a digest are denied (including untagged references and tag-only references).
//   - The tag latest is always denied, including forms like nginx:latest and nginx:latest@sha256:<digest>.
//   - Digest pinning is a supply-chain hygiene control; it does not prove image authenticity or trust in the registry.
//
// # Security context policy (regular, init, and ephemeral containers)
//
//   - Effective runAsNonRoot must be true. Container-level securityContext.runAsNonRoot overrides the pod-level value when set; otherwise the pod-level value applies.
//   - Effective runAsUser must not be 0. Container-level runAsUser overrides pod-level when set.
//   - When runAsNonRoot/runAsUser are unset, Kubernetes defaults apply (runAsNonRoot defaults to false); such workloads are denied.
//
// # Resources policy
//
//   - Every regular and init container MUST set cpu and memory in both requests and limits.
//   - For each resource, the limit quantity must be greater than or equal to the request (parsed with Kubernetes resource.Quantity semantics).
//   - Ephemeral containers: requests and limits are not required (per API semantics). When either requests or limits are present, quantities are validated and limits must be >= requests for cpu/memory that are specified.
package policy
