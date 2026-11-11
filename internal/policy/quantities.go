package policy

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

func limitGTERequest(name interface{}, req, lim resource.Quantity) error {
	if lim.Cmp(req) < 0 {
		return fmt.Errorf("%v limit %s must be >= request %s", name, lim.String(), req.String())
	}
	return nil
}
