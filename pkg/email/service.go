package email

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type Service interface {
	SendVerificationEmail(to, firstName, verificationToken string) error
}

type service struct {
	logger   *zap.SugaredLogger
	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
	appURL   string
}

func GetService(logger *zap.SugaredLogger) Service {
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	return &service{
		logger:   logger,
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: smtpPort,
		smtpUser: os.Getenv("SMTP_USER"),
		smtpPass: os.Getenv("SMTP_PASSWORD"),
		appURL:   os.Getenv("APP_URL"),
	}
}

func (s *service) SendVerificationEmail(to, firstName, verificationToken string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@todoapp.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Verify Your Email Address")

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, verificationToken)

	body := fmt.Sprintf(`
<html>
	<body>
		<h2>Welcome to Todo App, %s!</h2>
		<p><a href="%s">Verify Email</a></p>
		<p>This link will expire in 24 hours.</p>
	</body>
</html>
	`, firstName, verificationURL)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUser, s.smtpPass)

	if err := d.DialAndSend(m); err != nil {
		s.logger.Errorw("failed to send verification email", "error", err, "to", to)
		return err
	}

	s.logger.Infow("verification email sent successfully", "to", to)
	return nil
}
