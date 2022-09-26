package dispatch

import (
	"regexp"
	"sync"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

var p *Process
var one sync.Once

type Process struct {
	alertPcs AlertProcessor
	labelPcs LabelProcessor
	match    match
}

type combination struct {
	Key   *regexp.Regexp
	Value *regexp.Regexp
}

type match struct {
	LabelExtract []combination
	LabelMatch   []labelMatch
}

type labelMatch struct {
	label
	Receiver []string
}

func InitProcess() *Process {
	one.Do(initialization)
	return p
}

func initialization() {
	p = new(Process)
	p.init()
	p.initDispatch()
}

func (p *Process) init() {
	p.alertPcs = NewAlertProcess(conf.Conf.Alert)
	p.labelPcs = NewLabelProcess(conf.Conf.Label)
}

func (p *Process) initDispatch() {
	for _, dp := range conf.Conf.Dispatch.LabelMatch {
		if lb := buildLabel(dp.LabelKV); lb != nil {
			m := labelMatch{
				label:    *lb,
				Receiver: dp.Receiver,
			}
			p.match.LabelMatch = append(p.match.LabelMatch, m)
		}
	}
	for _, m := range conf.Conf.Dispatch.LabelExtractSender {
		for k, v := range m {
			p.match.LabelExtract = append(p.match.LabelExtract, combination{
				regexp.MustCompile(k),
				regexp.MustCompile(v),
			})
		}
	}
}

func (p *Process) Filter(alerts ...model.Alert) []model.Alert {
	alerts = p.alertPcs.Filter(alerts...)
	alerts = p.labelPcs.Filter(alerts...)
	return alerts
}

func (p *Process) Replace(alerts ...model.Alert) []model.Alert {
	alerts = p.labelPcs.Replace(alerts...)
	return alerts
}

func (p *Process) Dispatch(alerts ...model.Alert) []model.Alert {
	if len(p.match.LabelMatch) == 0 && len(p.match.LabelExtract) == 0 {
		return alerts
	}
	for idx, alert := range alerts {
		var sender []string
		for _, mc := range p.match.LabelMatch {
			for k, v := range alert.Labels {
				if matchRegex(mc.Key, k) ||
					matchRegex(mc.Value, v) ||
					matchCombination(mc.Combination, k, v) {
					sender = append(sender, mc.Receiver...)
					break
				}
			}
		}
		for _, c := range p.match.LabelExtract {
			for k, v := range alert.Labels {
				if c.Key.MatchString(k) {
					if ms := c.Value.FindStringSubmatch(v); len(ms) > 1 {
						for _, s := range ms[1:] {
							if len(s) != 0 {
								sender = append(sender, s)
							}
						}
					}
				}
			}
		}
		if len(sender) != 0 {
			sender = uniq(sender)
			alerts[idx].Receiver = sender
		}
	}
	return alerts
}

func uniq[T comparable](arrs []T) []T {
	if len(arrs) == 0 {
		return arrs
	}
	dict := make(map[T]struct{}, len(arrs))
	for _, arr := range arrs {
		dict[arr] = struct{}{}
	}
	arrs = arrs[:0]
	for k := range dict {
		arrs = append(arrs, k)
	}
	return arrs
}
