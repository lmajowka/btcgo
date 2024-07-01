package mail

import (
    "fmt"
    "log"
    "gopkg.in/mail.v2"
    "os"
)
// Configurações do servidor SMTP
smtpHost := os.Getenv("SMTP_HOST")
smtpPort := os.Getenv("SMTP_PORT")
smtpUser := os.Getenv("SMTP_USER")
smtpPass := os.Getenv("SMTP_PASS")
destinationEmail := os.Getenv("DESTINATION_EMAIL")

if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || destinationEmail == "" {
	log.Fatal("Variáveis de ambiente SMTP não estão completamente definidas")
}
func sendEmail(from string, password string, to []string, subject string, body string) error {
    m := mail.NewMessage()
    m.SetHeader("From", from)
    m.SetHeader("To", to...)
    m.SetHeader("Subject", subject)
    m.SetBody("text/plain", body)

    d := mail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

    // Enviar o email
    if err := d.DialAndSend(m); err != nil {
        return err
    }
    return nil
}

func main() {
    from := "seuemail@example.com"
    password := "suasenha"
    to := []string{"destinatario@example.com"}
    subject := "Assunto do Email"
    body := "Corpo do email"

    err := sendEmail(from, password, to, subject, body)
    if err != nil {
        log.Println("Erro ao enviar email:", err)
    } else {
        log.Println("Email enviado com sucesso!")
    }
}
