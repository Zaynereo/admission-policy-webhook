package policy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func validateRegularOrInitResources(c *corev1.Container) error {
	return validateRequiredResources(c.Name, c.Resources)
}

func validateEphemeralResources(c *corev1.Container) error {
	req := c.Resources.Requests
	lim := c.Resources.Limits
	if len(req) == 0 && len(lim) == 0 {
		return nil
	}
	return validateOptionalResourcePair(c.Name, "ephemeral", req, lim)
}

func validateRequiredResources(name string, res corev1.ResourceRequirements) error {
	if name == "" {
		name = "<unnamed>"
	}
	for _, resName := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
		req, reqOK := res.Requests[resName]
		lim, limOK := res.Limits[resName]
		if !reqOK || !limOK {
			return fmt.Errorf("container %q: %s requests and limits are required", name, resName)
		}
		if err := limitGTERequest(resName, req, lim); err != nil {
			return fmt.Errorf("container %q: %w", name, err)
		}
	}
	return nil
}

func validateOptionalResourcePair(name, kind string, req, lim corev1.ResourceList) error {
	for _, resName := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
		reqQ, reqOK := req[resName]
		limQ, limOK := lim[resName]
		if !reqOK && !limOK {
			continue
		}
		if reqOK != limOK {
			return fmt.Errorf("%s container %q: %s must have both request and limit when either is set", kind, name, resName)
		}
		if err := limitGTERequest(resName, reqQ, limQ); err != nil {
			return fmt.Errorf("%s container %q: %w", kind, name, err)
		}
	}
	return nil
}
