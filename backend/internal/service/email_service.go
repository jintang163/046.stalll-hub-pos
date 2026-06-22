package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"stalll-hub-pos/backend/config"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

type EmailMessage struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments map[string][]byte
}

func (s *EmailService) SendEmail(msg *EmailMessage) error {
	if !config.AppConfig.Email.Enabled {
		log.Printf("[EmailService] Email disabled in config, skipping send to %v: %s", msg.To, msg.Subject)
		return nil
	}
	if config.AppConfig.Email.Host == "" || config.AppConfig.Email.Username == "" {
		log.Printf("[EmailService] Email SMTP not configured, skipping send to %v: %s", msg.To, msg.Subject)
		return nil
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("no recipient email address")
	}

	go func() {
		if err := s.sendSMTP(msg); err != nil {
			log.Printf("[EmailService] Send email failed (to=%v subject=%s): %v", msg.To, msg.Subject, err)
		} else {
			log.Printf("[EmailService] Email sent successfully to %v: %s", msg.To, msg.Subject)
		}
	}()
	return nil
}

func (s *EmailService) sendSMTP(msg *EmailMessage) error {
	cfg := config.AppConfig.Email
	fromAddr := cfg.FromEmail
	if fromAddr == "" {
		fromAddr = cfg.Username
	}
	fromName := cfg.FromName
	if fromName == "" {
		fromName = "大排档POS系统"
	}

	boundary := "----=_Part_" + fmt.Sprintf("%d", time.Now().UnixNano())

	var headers strings.Builder
	headers.WriteString(fmt.Sprintf("From: %s <%s>\r\n", encodeRFC2047(fromName), fromAddr))
	headers.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.CC) > 0 {
		headers.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	headers.WriteString(fmt.Sprintf("Subject: %s\r\n", encodeRFC2047(msg.Subject)))
	headers.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	headers.WriteString("MIME-Version: 1.0\r\n")

	var body strings.Builder
	if len(msg.Attachments) > 0 {
		headers.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		headers.WriteString("\r\n")

		body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		if msg.IsHTML {
			body.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		} else {
			body.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		}
		body.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		body.WriteString(encodeQuotedPrintable(msg.Body))
		body.WriteString("\r\n")

		for filename, data := range msg.Attachments {
			body.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
			body.WriteString(fmt.Sprintf("Content-Type: application/octet-stream; name=\"%s\"\r\n", filename))
			body.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
			body.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			body.WriteString(encodeBase64Wrap(data, 76))
			body.WriteString("\r\n")
		}
		body.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else {
		if msg.IsHTML {
			headers.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		} else {
			headers.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		}
		headers.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		body.WriteString(encodeQuotedPrintable(msg.Body))
	}

	fullMsg := headers.String() + body.String()

	recipients := make([]string, 0, len(msg.To)+len(msg.CC)+len(msg.BCC))
	recipients = append(recipients, msg.To...)
	recipients = append(recipients, msg.CC...)
	recipients = append(recipients, msg.BCC...)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	if cfg.UseSSL || cfg.Port == 465 {
		return s.sendWithSSL(addr, cfg.Host, auth, fromAddr, recipients, []byte(fullMsg))
	}
	if cfg.UseTLS || cfg.Port == 587 {
		return s.sendWithStartTLS(addr, cfg.Host, auth, fromAddr, recipients, []byte(fullMsg))
	}

	return smtp.SendMail(addr, auth, fromAddr, recipients, []byte(fullMsg))
}

func (s *EmailService) sendWithSSL(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

func (s *EmailService) sendWithStartTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Quit()

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}
	if err := c.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}

func encodeRFC2047(s string) string {
	b := make([]byte, 0, len(s)*3)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x20 && c < 0x7f && c != '=' && c != '?' && c != '_' {
			b = append(b, c)
		} else {
			b = append(b, '=', byteToHex(c>>4), byteToHex(c&0x0f))
		}
	}
	if len(b) == len(s) {
		return s
	}
	return "=?utf-8?Q?" + string(b) + "?="
}

func byteToHex(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'A' + (b - 10)
}

func encodeQuotedPrintable(s string) string {
	var b strings.Builder
	lineLen := 0
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\n' {
			b.WriteString("\r\n")
			lineLen = 0
			continue
		}
		if r == '\r' {
			continue
		}
		if r < 0x80 && r != '=' && r != '?' && r != '_' && r != '.' {
			if lineLen >= 74 {
				b.WriteString("=\r\n")
				lineLen = 0
			}
			b.WriteRune(r)
			lineLen++
		} else {
			bytes := []byte(string(r))
			for _, by := range bytes {
				if lineLen >= 72 {
					b.WriteString("=\r\n")
					lineLen = 0
				}
				b.WriteString(fmt.Sprintf("=%02X", by))
				lineLen += 3
			}
		}
	}
	return b.String()
}

func encodeBase64Wrap(data []byte, lineLen int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder
	var buffer [3]byte
	n := len(data)
	written := 0

	for i := 0; i < n; i += 3 {
		count := 3
		if i+3 > n {
			count = n - i
		}
		for j := 0; j < 3; j++ {
			if j < count {
				buffer[j] = data[i+j]
			} else {
				buffer[j] = 0
			}
		}

		b1 := buffer[0] >> 2
		b2 := ((buffer[0] & 0x03) << 4) | (buffer[1] >> 4)
		b3 := ((buffer[1] & 0x0f) << 2) | (buffer[2] >> 6)
		b4 := buffer[2] & 0x3f

		result.WriteByte(alphabet[b1])
		result.WriteByte(alphabet[b2])
		written += 2
		if count > 1 {
			result.WriteByte(alphabet[b3])
		} else {
			result.WriteByte('=')
		}
		if count > 2 {
			result.WriteByte(alphabet[b4])
		} else {
			result.WriteByte('=')
		}
		written += 2

		if written >= lineLen && i+3 < n {
			result.WriteString("\r\n")
			written = 0
		}
	}
	return result.String()
}
