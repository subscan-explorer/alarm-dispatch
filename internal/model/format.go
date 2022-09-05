package model

import (
	"bytes"
	"sync"
	"time"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func (a Alert) Markdown() string {
	text := pool.Get().(*bytes.Buffer)
	text.Reset()
	switch a.Status {
	case AlertStatusFiring:
		text.WriteString(":trumpet: ")
	case AlertStatusResolved:
		text.WriteString(":+1: ")
	}
	text.WriteByte('*')
	text.WriteString(a.Status)
	text.WriteByte('*')
	text.WriteByte('\n')
	text.WriteString("*Time*: ")
	text.WriteString(a.StartsAt.Format(time.RFC3339))
	if !a.EndsAt.IsZero() {
		text.WriteString(" - ")
		text.WriteString(a.EndsAt.Format(time.RFC3339))
	}
	text.WriteByte('\n')
	for k, v := range a.Labels {
		text.WriteByte('>')
		text.WriteByte('`')
		text.WriteString(k)
		text.WriteByte('`')
		text.WriteString(": ")
		text.WriteString(v)
		text.WriteByte('\n')
	}
	for _, v := range a.Annotations {
		text.WriteString(v)
		text.WriteByte('\n')
	}
	result := text.String()
	pool.Put(text)
	return result
}

func (a Alert) MarkdownV2() string {
	text := pool.Get().(*bytes.Buffer)
	text.Reset()
	text.WriteByte('*')
	text.WriteString(a.Status)
	text.WriteByte('*')
	text.WriteByte('\n')
	text.WriteString("*Time*: ")
	text.WriteString(a.StartsAt.Format(time.RFC3339))
	if !a.EndsAt.IsZero() {
		text.WriteString(" - ")
		text.WriteString(a.EndsAt.Format(time.RFC3339))
	}
	text.WriteByte('\n')
	for k, v := range a.Labels {
		text.WriteByte('>')
		text.WriteByte('`')
		text.WriteString(k)
		text.WriteByte('`')
		text.WriteString(": ")
		text.WriteString(v)
		text.WriteByte('\n')
	}
	for _, v := range a.Annotations {
		text.WriteString(v)
		text.WriteByte('\n')
	}
	result := text.String()
	pool.Put(text)
	return result
}

func (a Alert) HTML() string {
	text := pool.Get().(*bytes.Buffer)
	text.Reset()
	text.WriteString("<b>")
	text.WriteString(a.Status)
	text.WriteString("</b>")
	text.WriteByte('\n')
	text.WriteString("<b>Time</b>: ")
	text.WriteString(a.StartsAt.Format(time.RFC3339))
	if !a.EndsAt.IsZero() {
		text.WriteString(" - ")
		text.WriteString(a.EndsAt.Format(time.RFC3339))
	}
	text.WriteByte('\n')
	for k, v := range a.Labels {
		text.WriteString(" ● ")
		text.WriteString("<code>")
		text.WriteString(k)
		text.WriteString("</code>")
		text.WriteString(": ")
		text.WriteString("<code>")
		text.WriteString(v)
		text.WriteString("</code>")
		text.WriteByte('\n')
	}
	for _, v := range a.Annotations {
		text.WriteString(v)
		text.WriteByte('\n')
	}
	result := text.String()
	pool.Put(text)
	return result
}

func (a Alert) EmailHTML() string {
	text := pool.Get().(*bytes.Buffer)
	text.Reset()

	text.WriteString(`<!DOCTYPE html>
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <link rel="stylesheet" type="text/css" id="u0"
        href="https://zh.rakko.tools/tools/129/lib/tinymce/skins/ui/oxide/content.min.css">
    <link rel="stylesheet" type="text/css" id="u1"
        href="https://zh.rakko.tools/tools/129/lib/tinymce/skins/content/default/content.min.css">
</head>

<body id="tinymce" class="mce-content-body " data-id="content" contenteditable="true" spellcheck="false">`)

	switch a.Status {
	case AlertStatusFiring:
		text.WriteString("<h3><span style=\"color: rgb(224, 62, 45);\" data-mce-style=\"color: #e03e2d;\">")
	case AlertStatusResolved:
		text.WriteString("<h3><span>")
	}
	text.WriteString(a.Status)
	text.WriteString("</span></h3>")
	text.WriteString("<b>Time: ")
	text.WriteString(a.StartsAt.Format(time.RFC3339))
	if !a.EndsAt.IsZero() {
		text.WriteString(" - ")
		text.WriteString(a.EndsAt.Format(time.RFC3339))
	}
	text.WriteString("</b><br>")
	if len(a.Labels) > 0 {
		text.WriteString("<ul>")
		for k, v := range a.Labels {
			text.WriteString("<li>")
			text.WriteString(k)
			text.WriteString(": <code data-mce-selected=\"inline-boundary\">")
			text.WriteString(v)
			text.WriteString("</code></li>")
		}
		text.WriteString("</ul>")
	}
	if len(a.Annotations) > 0 {
		for _, v := range a.Annotations {
			text.WriteString("<p>")
			text.WriteString(v)
			text.WriteString("</p>")
		}
	}
	result := text.String()
	pool.Put(text)
	return result
}
