package main

import (
	"fmt"

	"github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

var Client expo.PushClient

func init() {
	Client = *expo.NewPushClient(nil)
}

func pushToTokens(tokens []Token, title, body string, data map[string]string) []Response {
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
	responses, err := Client.PublishMultiple(messages)
	if err != nil {
		return []Response{Response{Error: fmt.Sprintf("Error pushing: %s", err)}}
	}

	var result []Response
	for _, response := range responses {
		if response.ValidateResponse() != nil {
			result = append(result, Response{Error: response.Message})
		} else {
			result = append(result, Response{Success: "#"})
		}
	}
	return result
}

func pushToCrsids(crsids []string, title, body string, data map[string]string) []Response {
	var messages []expo.PushMessage
	var result []Response
	for _, crsid := range crsids {
		token, ok := IdToToken[crsid]
		if !ok {
			result = append(result, Response{Error: fmt.Sprintf("\"%s\" not a recognised CRSID", crsid)})
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
	responses, err := Client.PublishMultiple(messages)
	if err != nil {
		return []Response{Response{Error: fmt.Sprintf("Error pushing: %s", err)}}
	}

	for _, response := range responses {
		if response.ValidateResponse() != nil {
			result = append(result, Response{Error: response.Message})
		} else {
			result = append(result, Response{Success: "#"})
		}
	}
	return result
}
