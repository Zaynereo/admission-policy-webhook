package policy

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func boolPtr(b bool) *bool { return &b }
func int64Ptr(i int64) *int64 { return &i }

const goodImage = "app:v1@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func compliantContainer(name string) corev1.Container {
	return corev1.Container{
		Name:  name,
		Image: goodImage,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},
	}
}

func TestValidatePodSpec_Table(t *testing.T) {
	tests := []struct {
		name    string
		spec    corev1.PodSpec
		wantErr bool
	}{
		{
			name: "allow compliant",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers:      []corev1.Container{compliantContainer("app")},
			},
		},
		{
			name: "deny missing digest",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers: []corev1.Container{{
					Name:  "app",
					Image: "app:v1",
					Resources: compliantContainer("app").Resources,
				}},
			},
			wantErr: true,
		},
		{
			name: "pod runAsNonRoot inherited",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers:      []corev1.Container{compliantContainer("app")},
			},
		},
		{
			name: "container overrides pod runAsNonRoot false",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers: []corev1.Container{{
					Name:      "app",
					Image:     goodImage,
					Resources: compliantContainer("app").Resources,
					SecurityContext: &corev1.SecurityContext{
						RunAsNonRoot: boolPtr(false),
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "container runAsUser 0 overrides",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: boolPtr(true),
					RunAsUser:    int64Ptr(1000),
				},
				Containers: []corev1.Container{{
					Name:      "app",
					Image:     goodImage,
					Resources: compliantContainer("app").Resources,
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: int64Ptr(0),
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "init container needs resources",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				InitContainers: []corev1.Container{{
					Name:  "init",
					Image: goodImage,
				}},
				Containers: []corev1.Container{compliantContainer("app")},
			},
			wantErr: true,
		},
		{
			name: "limits below requests",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers: []corev1.Container{{
					Name:  "app",
					Image: goodImage,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "ephemeral without resources allowed",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers:      []corev1.Container{compliantContainer("app")},
				EphemeralContainers: []corev1.EphemeralContainer{{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name:  "debug",
						Image: goodImage,
					},
				}},
			},
		},
		{
			name: "ephemeral partial resources denied",
			spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
				Containers:      []corev1.Container{compliantContainer("app")},
				EphemeralContainers: []corev1.EphemeralContainer{{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name:  "debug",
						Image: goodImage,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("10m"),
							},
						},
					},
				}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePodSpec(&tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePodSpec() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
