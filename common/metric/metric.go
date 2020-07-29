package metric

import (
	"time"

	gce "cloud.google.com/go/compute/metadata"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	monitoring "google.golang.org/api/monitoring/v3"

	"github.com/cobinhood/cobinhood-backend/common/logging"
)

//CreateMetric will create a Metric in stackdriver,
//which can be found in web monitor
func (m *Metric) CreateMetric() {
	logger := logging.NewLoggerTag("metric")
	stackdriver, err := getStackDriverService()

	if err != nil {
		logger.Error("get stackdriver err:%v", err)
	}

	err = m.CheckConfig()

	if err != nil {
		logger.Error("check metric config fail: %v", err)
	}

	md := monitoring.MetricDescriptor{
		Type:        m.Type,
		MetricKind:  m.MetricKind,
		ValueType:   m.ValueType,
		Unit:        m.Unit,
		Description: m.Description,
		DisplayName: m.DisplayName,
	}

	_, err = stackdriver.Projects.MetricDescriptors.Create(
		projectResource(), &md).Do()

	if err != nil {
		logger.Error("create metric fail: %v", err)
	}

}

//WriteMetric will send time series metric data to
//the metric has already created in stackdriver
//type is the metric url
func (m *Metric) WriteMetric(value interface{}, label ResourceLabel) {
	logger := logging.NewLoggerTag("metric")
	stackdriver, err := getStackDriverService()

	if err != nil {
		logger.Error("get stackdriver err:%v", err)
		return
	}
	var dataPointValue *monitoring.TypedValue
	switch m.ValueType {
	case "INT64":
		dataPointValue = &monitoring.TypedValue{
			Int64Value: value.(*int64),
		}
	case "BOOL":
		dataPointValue = &monitoring.TypedValue{
			BoolValue: value.(*bool),
		}
	case "DOUBLE":
		dataPointValue = &monitoring.TypedValue{
			DoubleValue: value.(*float64),
		}
	case "STRING":
		dataPointValue = &monitoring.TypedValue{
			StringValue: value.(*string),
		}
	case "DISTRIBUTION":
		dataPointValue = &monitoring.TypedValue{
			DistributionValue: value.(*monitoring.Distribution),
		}

	default:
		return
	}

	dataPoint := &monitoring.Point{
		Interval: &monitoring.TimeInterval{
			EndTime: time.Now().Format(time.RFC3339),
		},
		Value: dataPointValue,
	}

	// Write time series data.
	request := &monitoring.CreateTimeSeriesRequest{
		TimeSeries: []*monitoring.TimeSeries{
			{
				Metric: &monitoring.Metric{
					Type: m.Type,
				},
				Resource: &monitoring.MonitoredResource{
					Type:   label.ResourceType(),
					Labels: label.Label(),
				},
				Points: []*monitoring.Point{
					dataPoint,
				},
			},
		},
	}
	projectName := projectResource()
	_, err = stackdriver.Projects.TimeSeries.Create(projectName, request).Do()
}

func getStackDriverService() (*monitoring.Service, error) {

	oauthClient := oauth2.NewClient(context.Background(),
		google.ComputeTokenSource(""))

	return monitoring.New(oauthClient)

}

func projectResource() string {
	projectID, _ := gce.ProjectID()
	return "projects/" + projectID
}
