package models

import (
	"bytes"
	htmlTemplate "html/template"
	textTemplate "text/template"
)

type Template struct {
	html *htmlTemplate.Template
	text *textTemplate.Template
}

func Parse(name string, html string, text string) (*Template, error) {
	htmlParse, err := htmlTemplate.New(name).Parse(html)
	if err != nil {
		return nil, err
	}

	textParse, err := textTemplate.New(name).Parse(text)
	if err != nil {
		return nil, err
	}

	return &Template{
		html: htmlParse,
		text: textParse,
	}, nil
}

func (t *Template) Parse(name string, html string, text string) (*Template, error) {
	htmlParse, err := t.html.New(name).Parse(html)
	if err != nil {
		return nil, err
	}

	textParse, err := t.text.New(name).Parse(text)
	if err != nil {
		return nil, err
	}

	return &Template{
		html: htmlParse,
		text: textParse,
	}, nil
}

func (t *Template) Execute(htmlBuf *bytes.Buffer, textBuf *bytes.Buffer, data interface{}) error {
	err := t.html.Execute(htmlBuf, data)
	if err != nil {
		return err
	}
	err = t.text.Execute(textBuf, data)
	if err != nil {
		return err
	}
	return nil
}
