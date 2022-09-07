package dispatch

import (
	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

type LabelProcessor interface {
	Filter(...model.Alert) []model.Alert
	Replace(...model.Alert) []model.Alert
}
type LabelProcess struct {
	exclude *label
	keep    *label
	rp      replace
}

func NewLabelProcess(conf conf.Label) LabelProcessor {
	l := new(LabelProcess)
	if l.keep = buildLabel(conf.Keep.Label); l.keep == nil {
		l.exclude = buildLabel(conf.Exclude.Label)
	}

	l.rp.Key = buildReplace(conf.Replace.Label.Key)
	l.rp.Value = buildReplace(conf.Replace.Label.Value)
	return l
}

func (l *LabelProcess) Filter(alerts ...model.Alert) []model.Alert {
	var filterMap func(map[string]string) map[string]string
	if l.exclude != nil {
		filterMap = l.excludeLabel
	}
	if l.keep != nil {
		filterMap = l.keepLabel
	}
	if filterMap == nil {
		return alerts
	}
	for i := 0; i < len(alerts); i++ {
		alerts[i].Labels = filterMap(alerts[i].Labels)
	}
	return alerts
}

func (l *LabelProcess) Replace(alerts ...model.Alert) []model.Alert {
	if len(l.rp.Key) == 0 && len(l.rp.Value) == 0 {
		return alerts
	}
	for i := 0; i < len(alerts); i++ {
		alerts[i].Labels = l.replace(alerts[i].Labels)
	}
	return alerts
}

func (l *LabelProcess) excludeLabel(m map[string]string) map[string]string {
	for k, v := range m {
		if matchRegex(l.exclude.Key, k) ||
			matchRegex(l.exclude.Value, v) ||
			matchCombination(l.exclude.Combination, k, v) {
			delete(m, k)
		}
	}
	return m
}

func (l *LabelProcess) keepLabel(m map[string]string) map[string]string {
	for k, v := range m {
		if matchRegex(l.keep.Key, k) ||
			matchRegex(l.keep.Value, v) ||
			matchCombination(l.keep.Combination, k, v) {
		} else {
			delete(m, k)
		}
	}
	return m
}

func (l *LabelProcess) replace(m map[string]string) map[string]string {
	for k, v := range m {
		for _, re := range l.rp.Value {
			if str := re.regex.FindStringSubmatch(v); len(str) > 0 {
				if len(re.value) != 0 {
					m[k] = re.value
				} else if len(str) > 1 {
					m[k] = str[1]
				}
			}
		}
	}
	if len(l.rp.Key) != 0 {
		result := make(map[string]string)
		for k, v := range m {
			mc := false
			for _, re := range l.rp.Key {
				if str := re.regex.FindStringSubmatch(k); len(str) > 0 {
					if len(re.value) != 0 {
						mc = true
						result[re.value] = v
					} else if len(str) > 1 {
						mc = true
						result[str[1]] = v
					}
				}
			}
			if !mc {
				result[k] = v
			}
		}
		m = result
	}
	return m
}
