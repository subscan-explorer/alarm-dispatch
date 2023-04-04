package model

import (
	"strings"
	"time"
)

const (
	AlertStatusFiring   = "firing"
	AlertStatusResolved = "resolved"
)

type AlertType int8

const (
	AlertFiring AlertType = iota
	AlertResolved
)

type Alert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
	Receiver    []string          `json:"-"`
}

func (a Alert) GetTitle() string {
	return "Alert " + strings.Title(a.Status)
}

func (a Alert) AlertType() AlertType {
	if strings.EqualFold(a.Status, AlertStatusResolved) {
		return AlertResolved
	}
	return AlertFiring
}

func (a Alert) IsResolved() bool {
	return a.AlertType() == AlertResolved
}

type Notification struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}
