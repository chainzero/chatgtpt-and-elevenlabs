package whisper

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type WhisperText struct {
	Text string `json:"text"`
}

func Whisper(apikey string, file os.File) {
	// Set API endpoint and API key
	//endpoint := "https://api.openai.com/v1/audio/transcriptions"

	// Create a buffer to store the form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the audio file to the form data
	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		fmt.Printf("Error creating form file: %s\n", err)
	}

	fmt.Println("Part: ", part)

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Error copying file to form data: %s\n", err)
	}

}
