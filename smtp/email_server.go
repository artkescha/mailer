package smtp

import (
	"fmt"
	"strconv"
)

type EmailServer struct {
	Server   string
	Port     int
	Username string
	Password string
	Auth     bool
	SSLTLS   bool
	From     string
}

func (s *EmailServer) Address() string {
	if s == nil {
		return ""
	}

	return s.Server + ":" + strconv.Itoa(s.Port)
}

func (s *EmailServer) String() string {
	return fmt.Sprintf("[email server: host: %s, user: %s]", s.Address(), s.Username)
}
