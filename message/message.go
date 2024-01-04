package message

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

type Message struct {
	Id          int
	Date        time.Time
	From        []string
	To          []string
	Sender      string
	Subject     string
	ContentType string
	Body        string
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Content  []byte
}

func New() *Message {
	return &Message{
		From:        make([]string, 0),
		To:          make([]string, 0),
		Attachments: make([]Attachment, 0),
	}
}

func NewMessage(subject string, body string) *Message {
	msg := New()
	msg.Subject = subject
	msg.Body = body
	return msg
}

func (m *Message) GoString() string {
	return fmt.Sprintf("[message: id=%d, date=%s, sender=%s]", m.Id, m.Date, m.Sender)
}

func (m *Message) AddAttachment(fileName string) (err error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("reading file attachments failed: %s", err.Error())
	}
	_, fil := filepath.Split(fileName)

	m.Attachments = append(m.Attachments, Attachment{
		Filename: fil,
		Content:  data,
	})
	return nil
}

func (m *Message) BuildMessage(sender string, recipients []string) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("From: " + sender + "\r\n")
	buf.WriteString("To: " + strings.Join(recipients, ";") + "\r\n")
	buf.WriteString("Subject: " + m.Subject + "\r\n")
	var coder = base64.StdEncoding
	buf.WriteString("MIME-Version: 1.0\r\n")
	boundary := "f46d043c813270fc6b04c2d223da"
	if len(m.Attachments) > 0 {
		buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
		buf.WriteString("\r\n--" + boundary + "\r\n")
	}
	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n\r\n", "text/plain"))
	buf.WriteString(m.Body + "\r\n")
	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			buf.WriteString("\r\n\r\n--" + boundary + "\r\n")
			ext := filepath.Ext(attachment.Filename)
			mimetype := mime.TypeByExtension(ext)
			if mimetype != "" {
				mime := fmt.Sprintf("Content-Type: %s\r\n", mimetype)
				buf.WriteString(mime)
			} else {
				buf.WriteString("Content-Type: application/octet-stream\r\n")
			}
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")

			buf.WriteString("Content-Disposition: attachment; filename=\"=?UTF-8?B?")
			buf.WriteString(coder.EncodeToString([]byte(attachment.Filename)))
			buf.WriteString("?=\"\r\n\r\n")

			b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Content)))
			base64.StdEncoding.Encode(b, attachment.Content)

			// write base64 content in lines of up to 76 chars
			for i, l := 0, len(b); i < l; i++ {
				buf.WriteByte(b[i])
				if (i+1)%76 == 0 {
					buf.WriteString("\r\n")
				}
			}
			buf.WriteString("\r\n--" + boundary)
		}
		buf.WriteString("--")
	}
	return buf.Bytes()
}
