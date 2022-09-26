package dispatch

import (
	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/metrics"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

type AlertProcessor interface {
	Filter(...model.Alert) []model.Alert
	Replace(...model.Alert) []model.Alert
}

type AlertFilter struct {
	labelFilter *label
}

func NewAlertProcess(c conf.Alert) AlertProcessor {
	f := new(AlertFilter)
	f.labelFilter = buildLabel(c.Filter.Label)
	return f
}

func (a *AlertFilter) Filter(alerts ...model.Alert) []model.Alert {
	if a.labelFilter == nil {
		return alerts
	}
	result := make([]model.Alert, 0, len(alerts))
	for _, alert := range alerts {
		skip := false
		for k, v := range alert.Labels {
			if skip = matchRegex(a.labelFilter.Key, k) ||
				matchRegex(a.labelFilter.Value, v) ||
				matchCombination(a.labelFilter.Combination, k, v); skip {
				break
			}
		}
		if skip {
			metrics.IncAlertStatusCount("filter")
		} else {
			result = append(result, alert)
		}
	}
	return result
}

func (a *AlertFilter) Replace(al ...model.Alert) []model.Alert {
	return al
}
