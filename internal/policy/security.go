package policy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func validateSecurityContext(pod *corev1.PodSpec, c *corev1.Container, kind string) error {
	name := c.Name
	if name == "" {
		name = "<unnamed>"
	}

	runAsNonRoot := effectiveRunAsNonRoot(pod, c)
	if !runAsNonRoot {
		return fmt.Errorf("%s container %q: effective runAsNonRoot must be true (set on pod or container securityContext)", kind, name)
	}

	if uid, ok := effectiveRunAsUser(pod, c); ok && uid == 0 {
		return fmt.Errorf("%s container %q: effective runAsUser must not be 0", kind, name)
	}

	return nil
}

func effectiveRunAsNonRoot(pod *corev1.PodSpec, c *corev1.Container) bool {
	if c.SecurityContext != nil && c.SecurityContext.RunAsNonRoot != nil {
		return *c.SecurityContext.RunAsNonRoot
	}
	if pod.SecurityContext != nil && pod.SecurityContext.RunAsNonRoot != nil {
		return *pod.SecurityContext.RunAsNonRoot
	}
	return false
}

func effectiveRunAsUser(pod *corev1.PodSpec, c *corev1.Container) (int64, bool) {
	if c.SecurityContext != nil && c.SecurityContext.RunAsUser != nil {
		return *c.SecurityContext.RunAsUser, true
	}
	if pod.SecurityContext != nil && pod.SecurityContext.RunAsUser != nil {
		return *pod.SecurityContext.RunAsUser, true
	}
	return 0, false
}
