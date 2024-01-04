package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"github.com/artkescha/mailer/message"
)

type sender struct {
	server *EmailServer
}

type Sender interface {
	Send(recipients []string, message *message.Message) error
}

func (s *sender) Send(recipients []string, message *message.Message) error {
	return send(s.server, recipients, message)
}

func NewSender(server *EmailServer) Sender {
	return &sender{
		server: server,
	}
}

func send(server *EmailServer, recipients []string, message *message.Message) error {
	conn, err := net.Dial("tcp", server.Address())
	if err != nil {
		return fmt.Errorf("net dial server %s error: %s", server.Address(), err)
	}
	// TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         server.Server,
		MinVersion:         tls.VersionTLS10,
	}
	if server.SSLTLS {
		conn = tls.Client(conn, tlsConfig)
	}
	client, err := smtp.NewClient(conn, server.Server)
	if err != nil {
		return fmt.Errorf("create smtp client error: %s", err)
	}
	defer func() {
		_ = client.Quit()
	}()
	hasStartTLS, _ := client.Extension("STARTTLS")
	if hasStartTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("client startTLS error: %s", err)
		}
	}

	var auth smtp.Auth
	if server.Auth {
		// Exchange
		auth = LoginAuth(server.Username, server.Password)
	} else {
		// others
		auth = unencryptedAuth{
			smtp.PlainAuth("", server.From, server.Password, server.Server),
		}
	}

	if server.Password != "" {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentification smtp client error: %s", err)
		}
	}

	if err = client.Mail(server.From); err != nil {
		return fmt.Errorf("send email failed: %s", err)
	}
	for _, k := range recipients {
		if err = client.Rcpt(k); err != nil {
			return fmt.Errorf("rcpt failed: %s", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("start writer error: %s", err)
	}

	messageBody := message.BuildMessage(server.From, recipients)
	if _, err := w.Write(messageBody); err != nil {
		return fmt.Errorf("write body error: %s", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close writer error: %s", err)
	}

	return nil
}
