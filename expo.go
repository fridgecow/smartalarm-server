package main

import (
	"encoding/json"
	"log"

	"github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

var Client expo.PushClient

func init() {
	Client = *expo.NewPushClient(nil)
}

func pushToTokens(tokens []Token, title, body string, data map[string]string) []error {
	notificationData, _ := json.Marshal(map[string]string{"title": title, "body": body})
	data["notification"] = string(notificationData)
	var messages []expo.PushMessage
	for _, token := range tokens {
		messages = append(messages, expo.PushMessage{
			To:       token,
			Title:    title,
			Body:     body,
			Data:     data,
			Priority: expo.HighPriority,
		})
	}
	log.Printf("Pushing messages:\n%+v\n", messages)
	responses, err := Client.PublishMultiple(messages)
	if err != nil {
		return []error{Err("Error pushing: %s", err)}
	}

	var errs []error
	for _, response := range responses {
		if response.ValidateResponse() != nil {
			errs = append(errs, Err(response.Message))
		}
	}
	log.Printf("Messages pushed. Errors returned: %+v\n", errs)
	return errs
}

func pushToCrsids(crsids []string, title, body string, data map[string]string) []error {
	notificationData, _ := json.Marshal(map[string]string{"title": title, "body": body})
	if data == nil {
		data = make(map[string]string)
	}
	data["notification"] = string(notificationData)
	var messages []expo.PushMessage
	var errs []error
	for _, crsid := range crsids {
		token, ok := IdToToken[crsid]
		if !ok {
			errs = append(errs, Err("\"%s\" not a recognised CRSID", crsid))
		} else {
			messages = append(messages, expo.PushMessage{
				To:       token,
				Title:    title,
				Body:     body,
				Data:     data,
				Priority: expo.HighPriority,
			})
		}
	}
	log.Printf("Pushing messages:\n%+v\n", messages)
	responses, err := Client.PublishMultiple(messages)
	if err != nil {
		return []error{Err("Error pushing: %s", err)}
	}

	for _, response := range responses {
		if response.ValidateResponse() != nil {
			errs = append(errs, Err(response.Message))
		}
	}
	log.Printf("Messages pushed. Errors returned: %+v\n", errs)
	return errs
}
