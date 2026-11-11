package extract

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// Supported workload kinds validated by this webhook (Pod templates).
var SupportedKinds = map[string]bool{
	"Pod":         true,
	"Deployment":  true,
	"ReplicaSet":  true,
	"StatefulSet": true,
	"DaemonSet":   true,
	"Job":         true,
	"CronJob":     true,
}

// ScopeNote documents kinds outside policy enforcement.
const ScopeNote = "Only Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, and CronJob are validated; all other kinds are allowed without inspection."

// PodSpecFromObject unmarshals the submitted object and returns a PodSpec for supported kinds.
// Malformed JSON or unsupported structure for a supported kind returns an error (fail closed).
func PodSpecFromObject(kind string, raw []byte) (*corev1.PodSpec, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("object raw is empty")
	}

	switch kind {
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(raw, &pod); err != nil {
			return nil, fmt.Errorf("decode Pod: %w", err)
		}
		return &pod.Spec, nil
	case "Deployment":
		var d deploymentShape
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, fmt.Errorf("decode Deployment: %w", err)
		}
		if d.Spec.Template.Spec == nil {
			return nil, fmt.Errorf("deployment missing spec.template.spec")
		}
		return d.Spec.Template.Spec, nil
	case "ReplicaSet":
		var rs replicaSetShape
		if err := json.Unmarshal(raw, &rs); err != nil {
			return nil, fmt.Errorf("decode ReplicaSet: %w", err)
		}
		if rs.Spec.Template.Spec == nil {
			return nil, fmt.Errorf("replicaset missing spec.template.spec")
		}
		return rs.Spec.Template.Spec, nil
	case "StatefulSet":
		var ss statefulSetShape
		if err := json.Unmarshal(raw, &ss); err != nil {
			return nil, fmt.Errorf("decode StatefulSet: %w", err)
		}
		if ss.Spec.Template.Spec == nil {
			return nil, fmt.Errorf("statefulset missing spec.template.spec")
		}
		return ss.Spec.Template.Spec, nil
	case "DaemonSet":
		var ds daemonSetShape
		if err := json.Unmarshal(raw, &ds); err != nil {
			return nil, fmt.Errorf("decode DaemonSet: %w", err)
		}
		if ds.Spec.Template.Spec == nil {
			return nil, fmt.Errorf("daemonset missing spec.template.spec")
		}
		return ds.Spec.Template.Spec, nil
	case "Job":
		var job jobShape
		if err := json.Unmarshal(raw, &job); err != nil {
			return nil, fmt.Errorf("decode Job: %w", err)
		}
		if job.Spec.Template.Spec == nil {
			return nil, fmt.Errorf("job missing spec.template.spec")
		}
		return job.Spec.Template.Spec, nil
	case "CronJob":
		var cj cronJobShape
		if err := json.Unmarshal(raw, &cj); err != nil {
			return nil, fmt.Errorf("decode CronJob: %w", err)
		}
		if cj.Spec.JobTemplate.Spec.Template.Spec == nil {
			return nil, fmt.Errorf("cronjob missing spec.jobTemplate.spec.template.spec")
		}
		return cj.Spec.JobTemplate.Spec.Template.Spec, nil
	default:
		return nil, errUnsupportedKind{Kind: kind}
	}
}

type errUnsupportedKind struct {
	Kind string
}

func (e errUnsupportedKind) Error() string {
	return fmt.Sprintf("kind %q is outside validation scope", e.Kind)
}

// IsUnsupportedKind reports whether err indicates an out-of-scope kind (allow).
func IsUnsupportedKind(err error) bool {
	_, ok := err.(errUnsupportedKind)
	return ok
}

type podTemplateSpec struct {
	Spec *corev1.PodSpec `json:"spec"`
}

type deploymentShape struct {
	Spec struct {
		Template podTemplateSpec `json:"template"`
	} `json:"spec"`
}

type replicaSetShape struct {
	Spec struct {
		Template podTemplateSpec `json:"template"`
	} `json:"spec"`
}

type statefulSetShape struct {
	Spec struct {
		Template podTemplateSpec `json:"template"`
	} `json:"spec"`
}

type daemonSetShape struct {
	Spec struct {
		Template podTemplateSpec `json:"template"`
	} `json:"spec"`
}

type jobShape struct {
	Spec struct {
		Template podTemplateSpec `json:"template"`
	} `json:"spec"`
}

type cronJobShape struct {
	Spec struct {
		JobTemplate struct {
			Spec struct {
				Template podTemplateSpec `json:"template"`
			} `json:"spec"`
		} `json:"jobTemplate"`
	} `json:"spec"`
}
