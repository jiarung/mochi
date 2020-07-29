package metric

import (
	"errors"
	"strings"
)

var (
	//ValueTypeTable is the value type which stackdriver metric support
	ValueTypeTable = []string{
		"INT64",
		"BOOL",
		"DOUBLE",
		"STRING",
		"DISTRIBUTION",
	}

	//MetricKindTable is the Metric Kind which stackdriver metric support
	MetricKindTable = []string{"GAUGE", "DELTA", "CUMULATIVE"}
)

//Metric is stackdriver metric discript
type Metric struct {
	Type        string
	MetricKind  string
	ValueType   string
	Unit        string
	Description string
	DisplayName string
}

//CheckConfig check Metric type config
func (m *Metric) CheckConfig() error {

	var err error

	//TPYE must start with "custom.googleapis.com/"
	if !strings.HasPrefix(m.Type, "custom.googleapis.com/") {
		return errors.New("metric type error")
	}

	//MetricKind
	err = errors.New("metric kind must be [GAUGE, DELTA, CUMULATIVE]")

	for _, val := range MetricKindTable {
		if m.MetricKind == val {
			err = nil
		}
	}

	if err != nil {
		return err
	}

	//ValueType
	err = errors.New("value type must be [BOOL, INT64, DOUBLE, STRING, DISTRIBUTION, MONEY]")

	for _, val := range ValueTypeTable {
		if m.ValueType == val {
			err = nil
		}
	}

	if err != nil {
		return err
	}

	return err
}
