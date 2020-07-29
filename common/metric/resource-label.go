package metric

import (
	"os"
	"strings"

	gce "cloud.google.com/go/compute/metadata"
	"github.com/cobinhood/cobinhood-backend/common/logging"
)

//ResourceLabel interface of
//resource label of stackdriver metric
type ResourceLabel interface {
	Label() map[string]string
	ResourceType() string
}

//GkeContainerLabel this is resource label for GKE container
type GkeContainerLabel struct {
}

//Label will return GKE container proporties for groupping metric
func (gke *GkeContainerLabel) Label() map[string]string {
	projectID, _ := gce.ProjectID()
	zone, _ := gce.Zone()
	clusterName, _ := gce.InstanceAttributeValue("cluster-name")
	clusterName = strings.TrimSpace(clusterName)

	if len(os.Getenv("POD_ID")) < 0 {
		logger := logging.NewLoggerTag("metric")
		logger.Error("env: POD_ID not found")
	}
	proporties := map[string]string{
		"project_id":   projectID,
		"zone":         zone,
		"cluster_name": clusterName,
		// container name doesn't matter here,
		// because the metric is exported for
		// the pod, not the container
		"container_name": "",
		"pod_id":         os.Getenv("POD_ID"),
		// namespace_id and instance_id don't matter
		"namespace_id": "default",
		"instance_id":  "",
	}
	return proporties
}

//ResourceType will return the type of this label
func (gke *GkeContainerLabel) ResourceType() string {
	return "gke_container"
}
