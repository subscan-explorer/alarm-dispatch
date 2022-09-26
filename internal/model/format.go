package model

import (
	"bytes"
	"strings"
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

func GetByteBuf() *bytes.Buffer {
	b := pool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func PutByteBuf(b *bytes.Buffer) {
	pool.Put(b)
}

func (a Alert) Markdown() string {
	text := pool.Get().(*bytes.Buffer)
	text.Reset()
	text.WriteByte('*')
	text.WriteString(strings.Title(a.Status))
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

func (a Alert) HTML(lf, space string) string {
	if len(lf) == 0 {
		lf = "\n"
	}
	if len(space) == 0 {
		space = " "
	}
	headSpaces := space + space
	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteString("<b><i>")
	buf.WriteString(strings.Title(a.Status))
	buf.WriteString("</i></b>")
	buf.WriteString(lf)

	buf.WriteString("<b>")
	buf.WriteString("Start:")
	buf.WriteString(space)
	buf.WriteString("</b>")
	buf.WriteString(lf)

	buf.WriteString(headSpaces)
	buf.WriteString(a.StartsAt.Format(time.RFC3339))
	buf.WriteString(lf)

	if !a.EndsAt.IsZero() {
		buf.WriteString("<b>")
		buf.WriteString("End:")
		buf.WriteString(space)
		buf.WriteString("</b>")
		buf.WriteString(lf)

		buf.WriteString(headSpaces)
		buf.WriteString(a.EndsAt.Format(time.RFC3339))
		buf.WriteString(lf)
	}
	for k, v := range a.Annotations {
		buf.WriteString("<b>")
		buf.WriteString(strings.Title(k))
		buf.WriteByte(':')
		buf.WriteString(space)
		buf.WriteString("</b>")
		buf.WriteString(lf)

		buf.WriteString(headSpaces)
		buf.WriteString(v)
		buf.WriteString(lf)
	}
	if len(a.Labels) != 0 {
		buf.WriteString("<b>")
		buf.WriteString("Tag:")
		buf.WriteString(space)
		buf.WriteString("</b>")
		buf.WriteString(lf)
		for k, v := range a.Labels {
			buf.WriteString(headSpaces)
			buf.WriteString("•")
			buf.WriteString(space)
			buf.WriteString("<code>")
			buf.WriteString(k)
			buf.WriteString("</code>")
			buf.WriteString(":")
			buf.WriteString(space)
			buf.WriteString("<code>")
			buf.WriteString(v)
			buf.WriteString("</code>")
			buf.WriteString(lf)
		}
	}
	result := buf.String()
	pool.Put(buf)
	return result
}
