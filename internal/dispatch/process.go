package dispatch

import (
	"log"
	"regexp"
	"sync"

	"github.com/subscan-explorer/alarm-dispatch/conf"
	"github.com/subscan-explorer/alarm-dispatch/internal/model"
)

var p *Process
var one sync.Once

type Process struct {
	filter *label
	match  match
}

type label struct {
	Key         []*regexp.Regexp
	Value       []*regexp.Regexp
	Combination []combination
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
	p.initFilter()
	p.initDispatch()
	log.Printf("filter: %+v\n", p.filter)
	log.Printf("dispatch: %+v\n", p.match)
}

func buildLabel(lb conf.Label) *label {
	f := new(label)
	for _, k := range lb.Key {
		if len(k) == 0 {
			continue
		}
		f.Key = append(f.Key, regexp.MustCompile(k))
	}
	for _, v := range lb.Value {
		if len(v) == 0 {
			continue
		}
		f.Value = append(f.Value, regexp.MustCompile(v))
	}
	for _, m := range lb.Combination {
		for k, v := range m {
			var c = combination{
				regexp.MustCompile(k),
				regexp.MustCompile(v),
			}
			f.Combination = append(f.Combination, c)
		}
	}
	if len(f.Key) != 0 || len(f.Value) != 0 || len(f.Combination) != 0 {
		p.filter = f
	}
	return f
}

func (p *Process) initFilter() {
	if lb := buildLabel(conf.Conf.Filter.Label); lb != nil {
		p.filter = lb
	}
}

func (p *Process) initDispatch() {
	for _, dp := range conf.Conf.Dispatch.LabelMatch {
		if lb := buildLabel(dp.Label); lb != nil {
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

func matchRegex(re []*regexp.Regexp, str string) bool {
	for _, r := range re {
		if r.MatchString(str) {
			return true
		}
	}
	return false
}

func matchCombination(ct []combination, k, v string) bool {
	for _, c := range ct {
		if c.Key.MatchString(k) && c.Value.MatchString(v) {
			return true
		}
	}
	return false
}

func (p *Process) Filter(alerts ...model.Alert) []model.Alert {
	if p.filter == nil {
		return alerts
	}
	result := make([]model.Alert, 0, len(alerts))
	for _, alert := range alerts {
		skip := false
		for k, v := range alert.Labels {
			if skip = matchRegex(p.filter.Key, k) ||
				matchRegex(p.filter.Value, v) ||
				matchCombination(p.filter.Combination, k, v); skip {
				break
			}
		}
		if !skip {
			result = append(result, alert)
		}
	}
	return result
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
