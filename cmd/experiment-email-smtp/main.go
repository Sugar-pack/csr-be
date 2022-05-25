package main

import (
	"log"
	"net/smtp"
)

func main() {
	send("hello there this is my first message")
}

func send(body string) {
	from := "noreply.lkot@yandex.ru" //your from address, don't forget to grant access from mailing apps in yandex mail settings
	pass := ""                       //add here application password generated on https://passport.yandex.ru/profile/access/apppasswords
	to := "any_destination_address@any_domain.com"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\n\n" +
		body

	err := smtp.SendMail("smtp.yandex.ru:587",
		smtp.PlainAuth("", from, pass, "smtp.yandex.ru"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("sent")
}
