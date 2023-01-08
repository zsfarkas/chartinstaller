package releases

import (
	"errors"

	"helm.sh/helm/v3/pkg/release"
)

type ReleaseStatus struct {
	Name      string `json:"name" example:"release1"`
	Namespace string `json:"namespace" example:"namespace1"`
	Chart     string `json:"chart" example:"chart1"`
	Revision  int    `json:"revision" example:"1"`
	Status    string `json:"status" example:"deployed"`
}

func NewReleaseStatusFrom(release *release.Release) *ReleaseStatus {
	return &ReleaseStatus{
		Name:      release.Name,
		Namespace: release.Namespace,
		Chart:     release.Chart.Name() + "-" + release.Chart.Metadata.Version,
		Revision:  release.Version,
		Status:    release.Info.Status.String(),
	}
}

type ReleaseRequest struct {
	Chart  string                 `json:"chart" example:"chart1"`
	Values map[string]interface{} `json:"values" example:"{}" format:"any"`
}

func (releaseRequest ReleaseRequest) Validation() error {
	switch {
	case len(releaseRequest.Chart) == 0:
		return errors.New("chart name is empty")
	default:
		return nil
	}
}
