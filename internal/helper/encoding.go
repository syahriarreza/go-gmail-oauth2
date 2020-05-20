package helper

import (
	"encoding/base64"
	"strings"

	tk "github.com/eaciit/toolkit"
)

const (
	//CRLF CR+LF
	CRLF = "\r\n"
)

//MessageRFC2822 message struct
type MessageRFC2822 struct {
	Headers tk.M
	Body    string
}

//ToStringRFC2822 creates message string in RFC2822 format
func (m *MessageRFC2822) ToStringRFC2822() string {
	message := ""
	for k, v := range m.Headers {
		message += k + ": " + tk.ToString(v) + CRLF
	}
	message += CRLF + m.Body
	return message
}

// EncodeRFC2822 encode as RFC 2822 format
func (m *MessageRFC2822) EncodeRFC2822() string {
	b := []byte(m.ToStringRFC2822())
	s := base64.StdEncoding.EncodeToString(b)
	s = strings.Replace(s, "/", "_", -1)
	s = strings.Replace(s, "+", "-", -1)
	s = strings.Replace(s, "=", "", -1)
	return s
}
