package admission

import (
	"encoding/json"
	"fmt"
	"io"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/zaynereo/admission-policy-webhook/internal/extract"
	"github.com/zaynereo/admission-policy-webhook/internal/policy"
)

// MaxAdmissionBodyBytes bounds request bodies (aligned with common apiserver limits).
const MaxAdmissionBodyBytes = 3 << 20 // 3 MiB

// Review decodes an AdmissionReview, validates when in scope, and returns a response with UID echoed.
func Review(body []byte) ([]byte, int, error) {
	if len(body) == 0 {
		return denyMalformed("empty body")
	}
	if len(body) > MaxAdmissionBodyBytes {
		return denyMalformed(fmt.Sprintf("body exceeds %d bytes", MaxAdmissionBodyBytes))
	}

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		return denyMalformed(fmt.Sprintf("invalid JSON: %v", err))
	}

	if review.Request == nil {
		return denyMalformed("missing request")
	}

	req := review.Request
	if req.UID == "" {
		return denyMalformed("missing request.uid")
	}

	// Only admission.k8s.io/v1 reviews are accepted; wrong API version fails closed.
	if review.APIVersion != "" && review.APIVersion != "admission.k8s.io/v1" {
		return denyReview(req.UID, false, fmt.Sprintf("unsupported admission review apiVersion %q", review.APIVersion))
	}

	allowed, reason, err := evaluate(req)
	if err != nil {
		return denyMalformed(err.Error())
	}

	resp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     req.UID,
			Allowed: allowed,
		},
	}
	if !allowed {
		resp.Response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    403,
			Reason:  metav1.StatusReasonForbidden,
			Message: reason,
		}
	}

	out, err := json.Marshal(resp)
	if err != nil {
		return nil, 500, fmt.Errorf("marshal response: %w", err)
	}
	return out, 200, nil
}

func evaluate(req *admissionv1.AdmissionRequest) (bool, string, error) {
	if req.Object.Raw == nil {
		return false, "missing object", nil
	}

	kind := req.Kind.Kind
	if kind == "" && req.Object.Object != nil {
		// fallback if needed
	}
	if kind == "" {
		return false, "missing resource kind", nil
	}

	if req.Operation != admissionv1.Create && req.Operation != admissionv1.Update {
		return true, "", nil
	}

	spec, err := extract.PodSpecFromObject(kind, req.Object.Raw)
	if extract.IsUnsupportedKind(err) {
		return true, "", nil
	}
	if err != nil {
		return false, err.Error(), nil
	}

	if err := policy.ValidatePodSpec(spec); err != nil {
		return false, err.Error(), nil
	}
	return true, "", nil
}

func denyReview(uid types.UID, allowed bool, msg string) ([]byte, int, error) {
	resp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     uid,
			Allowed: allowed,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Code:    400,
				Reason:  metav1.StatusReasonBadRequest,
				Message: msg,
			},
		},
	}
	out, err := json.Marshal(resp)
	if err != nil {
		return nil, 500, err
	}
	return out, 200, nil
}

// MalformedResponse returns an AdmissionReview denial for bad HTTP payloads (exported for server layer).
func MalformedResponse(msg string) ([]byte, int, error) {
	return denyMalformed(msg)
}

func denyMalformed(msg string) ([]byte, int, error) {
	resp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  metav1.StatusFailure,
				Code:    400,
				Reason:  metav1.StatusReasonBadRequest,
				Message: msg,
			},
		},
	}
	out, err := json.Marshal(resp)
	if err != nil {
		return nil, 500, err
	}
	return out, 400, nil
}

// ReadBoundedBody reads at most MaxAdmissionBodyBytes+1 from r to detect overflow.
func ReadBoundedBody(r io.Reader) ([]byte, error) {
	limited := io.LimitReader(r, MaxAdmissionBodyBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxAdmissionBodyBytes {
		return nil, fmt.Errorf("body exceeds %d bytes", MaxAdmissionBodyBytes)
	}
	return data, nil
}
