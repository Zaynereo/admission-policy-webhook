package admission

import (
	"encoding/json"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const goodImage = "app:v1@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func compliantPod() *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: ptr(true)},
			Containers: []corev1.Container{{
				Name:  "app",
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
			}},
		},
	}
}

func ptr[T any](v T) *T { return &v }

func buildReview(t *testing.T, uid string, gvk metav1.GroupVersionKind, obj any, op admissionv1.Operation, apiVer string) []byte {
	t.Helper()
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	ar := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: apiVer, Kind: "AdmissionReview"},
		Request: &admissionv1.AdmissionRequest{
			UID:       types.UID(uid),
			Operation: op,
			Kind:      gvk,
			Object: runtime.RawExtension{
				Raw: raw,
			},
		},
	}
	out, err := json.Marshal(ar)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestReview_Table(t *testing.T) {
	uid := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	tests := []struct {
		name       string
		body       []byte
		wantCode   int
		wantAllow  bool
		wantUID    string
		wantSubstr string
	}{
		{
			name:      "allow compliant pod",
			body:      buildReview(t, uid, metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, compliantPod(), admissionv1.Create, "admission.k8s.io/v1"),
			wantCode:  200,
			wantAllow: true,
			wantUID:   uid,
		},
		{
			name:      "deny bad image",
			body:      buildReview(t, uid, metav1.GroupVersionKind{Version: "v1", Kind: "Pod"}, func() *corev1.Pod {
				p := compliantPod()
				p.Spec.Containers[0].Image = "nginx:latest"
				return p
			}(), admissionv1.Create, "admission.k8s.io/v1"),
			wantCode:    200,
			wantAllow:   false,
			wantUID:     uid,
			wantSubstr:  "latest",
		},
		{
			name:     "echo uid",
			body:     buildReview(t, uid, metav1.GroupVersionKind{Version: "v1", Kind: "Pod"}, compliantPod(), admissionv1.Create, "admission.k8s.io/v1"),
			wantCode: 200,
			wantUID:  uid,
			wantAllow: true,
		},
		{
			name:     "malformed json",
			body:     []byte("{"),
			wantCode: 400,
		},
		{
			name:     "wrong api version",
			body:     buildReview(t, uid, metav1.GroupVersionKind{Version: "v1", Kind: "Pod"}, compliantPod(), admissionv1.Create, "admission.k8s.io/v1beta1"),
			wantCode: 200,
			wantAllow: false,
			wantUID:  uid,
		},
		{
			name: "unsupported kind allowed",
			body: buildReview(t, uid, metav1.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"data":       map[string]string{"k": "v"},
			}, admissionv1.Create, "admission.k8s.io/v1"),
			wantCode:  200,
			wantAllow: true,
			wantUID:   uid,
		},
		{
			name: "cronjob path",
			body: buildReview(t, uid, metav1.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"}, map[string]any{
				"apiVersion": "batch/v1",
				"kind":       "CronJob",
				"spec": map[string]any{
					"jobTemplate": map[string]any{
						"spec": map[string]any{
							"template": map[string]any{
								"spec": compliantPod().Spec,
							},
						},
					},
				},
			}, admissionv1.Create, "admission.k8s.io/v1"),
			wantCode:  200,
			wantAllow: true,
			wantUID:   uid,
		},
		{
			name: "deployment path",
			body: buildReview(t, uid, metav1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"spec": map[string]any{
					"template": map[string]any{
						"spec": compliantPod().Spec,
					},
				},
			}, admissionv1.Create, "admission.k8s.io/v1"),
			wantCode:  200,
			wantAllow: true,
			wantUID:   uid,
		},
		{
			name: "malformed deployment fail closed",
			body: buildReview(t, uid, metav1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"spec":       map[string]any{},
			}, admissionv1.Create, "admission.k8s.io/v1"),
			wantCode:   200,
			wantAllow:  false,
			wantUID:    uid,
			wantSubstr: "template",
		},
		{
			name:      "missing object",
			body:      mustReviewNoObject(t, uid),
			wantCode:  200,
			wantAllow: false,
			wantUID:   uid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code, err := Review(tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if code != tt.wantCode {
				t.Fatalf("code=%d want %d body=%s", code, tt.wantCode, out)
			}
			if tt.wantUID == "" && tt.wantCode == 400 {
				return
			}
			var resp admissionv1.AdmissionReview
			if err := json.Unmarshal(out, &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Response == nil {
				t.Fatal("missing response")
			}
			if string(resp.Response.UID) != tt.wantUID && tt.wantUID != "" {
				t.Fatalf("uid=%q want %q", resp.Response.UID, tt.wantUID)
			}
			if resp.Response.Allowed != tt.wantAllow {
				t.Fatalf("allowed=%v want %v msg=%v", resp.Response.Allowed, tt.wantAllow, resp.Response.Result)
			}
			if tt.wantSubstr != "" && resp.Response.Result != nil {
				if !strings.Contains(resp.Response.Result.Message, tt.wantSubstr) {
					t.Fatalf("message=%q want substring %q", resp.Response.Result.Message, tt.wantSubstr)
				}
			}
		})
	}
}

func mustReviewNoObject(t *testing.T, uid string) []byte {
	ar := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request: &admissionv1.AdmissionRequest{
			UID:       types.UID(uid),
			Operation: admissionv1.Create,
			Kind:      metav1.GroupVersionKind{Version: "v1", Kind: "Pod"},
		},
	}
	b, err := json.Marshal(ar)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
