package policy

import (
	"fmt"
	"regexp"
	"strings"
)

var sha256Digest = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// ValidateImageReference applies digest-pinning rules documented in doc.go.
func ValidateImageReference(image string) error {
	if image == "" {
		return fmt.Errorf("image must not be empty")
	}

	at := strings.LastIndex(image, "@")
	if at < 0 {
		return fmt.Errorf("image %q must include a SHA256 digest (@sha256:<64-hex>)", image)
	}

	ref := image[:at]
	digestPart := image[at+1:]
	if !sha256Digest.MatchString(digestPart) {
		return fmt.Errorf("image %q must use @sha256:<64-hex> digest form", image)
	}

	if tagIsLatest(ref) {
		return fmt.Errorf("image %q uses forbidden tag latest", image)
	}

	return nil
}

func tagIsLatest(ref string) bool {
	// Strip registry/port with possible single host component; tag is after the last ':' in the repository portion
	// when that ':' is not part of a port (handled by using last ':' before '@' which we already split).
	if ref == "" {
		return false
	}
	colon := strings.LastIndex(ref, ":")
	if colon < 0 {
		return false
	}
	// Digest or algo in path should not appear before '@'; repository may contain ports like localhost:5000/foo
	tag := ref[colon+1:]
	return strings.EqualFold(tag, "latest")
}
