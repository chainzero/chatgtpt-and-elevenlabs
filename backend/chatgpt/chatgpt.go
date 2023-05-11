package chatgpt

import (
	"lambdaanddynamo/database"
)

// Get response for message posted to ChatGPT
func ChatGptResponse(apiKey, organization string, messageList database.MessageList) database.MessageList {
	client := NewClient(apiKey, organization)

	r := CreateCompletionsRequest{
		Model:       "gpt-3.5-turbo",
		Messages:    messageList.Messages,
		Temperature: 0.7,
		N:           1,
	}

	completions, err := client.CreateCompletions(r)
	if err != nil {
		panic(err)
	}

	assistantMessage := database.Message{
		Role:    completions.Choices[0].Message.Role,
		Content: completions.Choices[0].Message.Content,
	}

	messageList.Messages = append(messageList.Messages, assistantMessage)

	return messageList

}
