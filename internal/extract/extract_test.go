package extract

import (
	"testing"
)

const podTemplate = `{
  "spec": {
    "template": {
      "spec": {
        "containers": [{"name":"c","image":"x"}]
      }
    }
  }
}`

const cronJobBody = `{
  "spec": {
    "jobTemplate": {
      "spec": {
        "template": {
          "spec": {
            "containers": [{"name":"c","image":"x"}]
          }
        }
      }
    }
  }
}`

func TestPodSpecFromObject_Paths(t *testing.T) {
	tests := []struct {
		kind    string
		raw     string
		wantErr bool
		skip    bool
	}{
		{"Deployment", podTemplate, false, false},
		{"CronJob", cronJobBody, false, false},
		{"Deployment", `{"spec":{}}`, true, false},
		{"ConfigMap", `{"data":{}}`, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			spec, err := PodSpecFromObject(tt.kind, []byte(tt.raw))
			if tt.skip {
				if !IsUnsupportedKind(err) || spec != nil {
					t.Fatalf("expected unsupported kind, err=%v spec=%v", err, spec)
				}
				return
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && spec == nil {
				t.Fatal("expected spec")
			}
		})
	}
}
