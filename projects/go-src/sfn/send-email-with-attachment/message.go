package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
)

type Message struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        Body
	Attachments map[string][]byte
}

type Body struct {
	Raw         string
	ContentType string
}

func (m *Message) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("From: %s\n", m.From))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(m.CC, ",")))
	}

	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	}

	buf.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n\n", boundary))
	buf.WriteString(fmt.Sprintf("--%s\n", boundary))

	//body
	buf.WriteString(fmt.Sprintf("Content-Type: %s\n\n", m.Body.ContentType))
	buf.WriteString(m.Body.Raw)

	//attachments
	for k, v := range m.Attachments {
		buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
		buf.WriteString("Content-Transfer-Encoding: base64\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n\n", k))

		b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
		base64.StdEncoding.Encode(b, v)
		buf.Write(b)
	}

	buf.WriteString(fmt.Sprintf("\n\n--%s--", boundary))
	return buf.Bytes()
}
