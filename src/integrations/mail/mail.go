package mail

import (
    "fmt"
    "log"
    "net/smtp"
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

func main() {
    for update := range updates {
        if update.Message != nil {
            msg := update.Message.Text
            chatID := update.Message.Chat.ID

            // Enviar resposta no Telegram
            response := tgbotapi.NewMessage(chatID, "Mensagem recebida: "+msg)
            bot.Send(response)

            // Enviar e-mail
            err := sendEmail("destination_email@example.com", "Mensagem do Telegram", msg)
            if err != nil {
                log.Printf("Erro ao enviar e-mail: %v", err)
            } else {
                log.Printf("E-mail enviado com sucesso!")
            }
        }
    }
}

// Função para enviar e-mails
func sendEmail(to, subject, body string) error {
    from := smtpUser
    pass := smtpPass

    msg := "From: " + from + "\n" +
        "To: " + to + "\n" +
        "Subject: " + subject + "\n\n" +
        body

    err := smtp.SendMail(smtpHost+":"+smtpPort,
        smtp.PlainAuth("", from, pass, smtpHost),
        from, []string{to}, []byte(msg))

    return err
}
