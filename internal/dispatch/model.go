package dispatch

import (
	"regexp"

	"github.com/subscan-explorer/alarm-dispatch/conf"
)

type label struct {
	Key         []*regexp.Regexp
	Value       []*regexp.Regexp
	Combination []combination
}

type replaceValue struct {
	regex *regexp.Regexp
	value string
}
type replace struct {
	Key   []replaceValue
	Value []replaceValue
}

func buildLabel(lb conf.LabelKV) *label {
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
		return f
	}
	return nil
}

func buildReplace(rp []conf.ReplaceValue) []replaceValue {
	result := make([]replaceValue, 0, len(rp))
	for _, rv := range rp {
		result = append(result, replaceValue{
			regex: regexp.MustCompile(rv.Regex),
			value: rv.Value,
		})
	}
	return result
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
