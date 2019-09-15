package notification_client

import (
	"encoding/json"

	"gitlab.faza.io/go-framework/kafkaadapter"
)

type email struct {
	From       string
	To         string
	Subject    string
	Body       string
	Attachment []string
}

type sms struct {
	To   string
	Body string
}

func SendSms(brokers []string, to, body string) error {
	k := kafkaadapter.NewKafka(brokers, "notification-sms-single")
	k.Config.Producer.Return.Successes = true
	s := sms{
		To:   to,
		Body: body,
	}
	smsByte, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, _, err = k.SendOne("", smsByte)
	if err != nil {
		return err
	}
	return nil
}

func SendEmail(brokers []string, from, to, subject, body string, attachment []string) error {
	k := kafkaadapter.NewKafka(brokers, "notification-email-single")
	k.Config.Producer.Return.Successes = true
	s := email{
		From:       from,
		To:         to,
		Subject:    subject,
		Body:       body,
		Attachment: attachment,
	}
	smsByte, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, _, err = k.SendOne("", smsByte)
	if err != nil {
		return err
	}
	return nil
}
