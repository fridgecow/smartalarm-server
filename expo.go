package main

import (
	"fmt"

	"github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

var Client expo.PushClient

func init() {
	Client = *expo.NewPushClient(nil)
}

func pushToTokens(tokens []Token, title, body string, data map[string]string) []error {
	var messages []expo.PushMessage
	for _, token := range tokens {
		messages = append(messages, expo.PushMessage{
			To: token,
			Title: title,
			Body: body,
			Data: data,
		})
	}
	responses, err := Client.PublishMultiple(messages)
	if err != nil {
		return []error{fmt.Errorf("Error pushing: %s", err)}
	}

	var errs []error
	for _, response := range responses {
		if response.ValidateResponse() != nil {
			errs = append(errs, fmt.Errorf(response.Message))
		}
	}
	return errs
}

func pushToCrsids(crsids []string, title, body string, data map[string]string) (error, []string, []Token) {
	var messages []expo.PushMessage
	var nocrsid []string
	for _, crsid := range crsids {
		token, ok := IdToToken[crsid]
		if !ok {
			nocrsid = append(nocrsid, crsid)
			continue
		}
		messages = append(messages, expo.PushMessage{
			To: token,
			Title: title,
			Body: body,
			Data: data,
		})
	}
	responses, err := Client.PublishMultiple(messages)
	if err != nil {
		return fmt.Errorf("Error pushing: %s", err), nil, nil
	}

	var failed []Token
	for _, response := range responses {
		if response.ValidateResponse() != nil {
			failed = append(failed, response.PushMessage.To)
		}
	}
	return nil, nocrsid, failed
}
