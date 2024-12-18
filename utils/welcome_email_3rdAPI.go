package utils

import (
	"log"
	"os"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

func SendWelcomeEmail(email, name string) error {
	client := mailjet.NewMailjetClient(os.Getenv("MAILJET_API_KEY"), os.Getenv("MAILJET_API_SECRET"))

	messages := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "welcome@hotelgardenia.com",
				Name:  "Hotel Gardenia",
			},
			To: &mailjet.RecipientsV31{
				{
					Email: email,
					Name:  name,
				},
			},
			Subject:  "Thank You For Registering",
			TextPart: "Hello " + name + ",\n\nThank you for registering with Hotel Gardenia App!",
			HTMLPart: "<h3>Hello " + name + ",</h3><p>Thank you for registering with Hotel Gardenia App!</p>",
		},
	}

	messagesBody := mailjet.MessagesV31{Info: messages}

	_, err := client.SendMailV31(&messagesBody)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
