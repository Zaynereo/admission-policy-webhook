package policy

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ValidatePodSpec applies documented policies to a PodSpec. Returns nil when allowed.
func ValidatePodSpec(spec *corev1.PodSpec) error {
	if spec == nil {
		return fmt.Errorf("pod spec is required")
	}

	var violations []string

	check := func(kind string, containers []corev1.Container, requireResources bool) {
		for i := range containers {
			c := &containers[i]
			if err := ValidateImageReference(c.Image); err != nil {
				violations = append(violations, fmt.Sprintf("%s: %v", kind, err))
			}
			if err := validateSecurityContext(spec, c, kind); err != nil {
				violations = append(violations, err.Error())
			}
			if requireResources {
				if err := validateRegularOrInitResources(c); err != nil {
					violations = append(violations, err.Error())
				}
			} else {
				if err := validateEphemeralResources(c); err != nil {
					violations = append(violations, err.Error())
				}
			}
		}
	}

	check("container", spec.Containers, true)
	check("initContainer", spec.InitContainers, true)
	for i := range spec.EphemeralContainers {
		c := spec.EphemeralContainers[i].EphemeralContainerCommon
		container := corev1.Container{
			Name:            c.Name,
			Image:           c.Image,
			Resources:       c.Resources,
			SecurityContext: c.SecurityContext,
		}
		if err := ValidateImageReference(container.Image); err != nil {
			violations = append(violations, fmt.Sprintf("ephemeralContainer: %v", err))
		}
		if err := validateSecurityContext(spec, &container, "ephemeral"); err != nil {
			violations = append(violations, err.Error())
		}
		if err := validateEphemeralResources(&container); err != nil {
			violations = append(violations, err.Error())
		}
	}

	if len(violations) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(violations, "; "))
}
