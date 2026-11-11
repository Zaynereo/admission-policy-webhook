package policy

import "testing"

const testDigest = "@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestValidateImageReference(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		wantErr bool
	}{
		{"pinned ok", "registry.example/app:1.0" + testDigest, false},
		{"untagged digest ok", "registry.example/app" + testDigest, false},
		{"no digest", "nginx:1.25", true},
		{"untagged no digest", "nginx", true},
		{"latest tag", "nginx:latest", true},
		{"latest with digest", "nginx:latest" + testDigest, true},
		{"short digest", "nginx@sha256:abc", true},
		{"upper digest rejected", "nginx@sha256:" + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageReference(tt.image)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateImageReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
