package restendpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"lambdaanddynamo/chatgpt"
	"lambdaanddynamo/database"
	"lambdaanddynamo/texttospeech"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	router = gin.Default()
)

const (
	filename            = "./database/stored_data.json"
	OPEN_AI_ORG         = "change_me"
	OPEN_AI_KEY         = "change_me"
	ELEVEN_LABS_API_KEY = "change_me"
	endpoint            = "https://api.openai.com/v1/audio/transcriptions"
)

type WhisperText struct {
	Text string `json:"text"`
}

func mapUrls() {
	router.GET("/ping", Ping)
	router.GET("/resetmessages", ResetMessageArray)
	router.POST("/postaudio", PostAudio)
}

func StartRestEndpoints() {
	mapUrls()
	fmt.Println("about to start application")
	router.Use(cors.Default())
	router.Run(":80")
}

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{"Ping": "staus is ok"})
}

func ResetMessageArray(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")

	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Max-Age", "86400")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	c.Header("Access-Control-Allow-Credentials", "true")

	c.JSON(200, gin.H{"ResetMessageArray": "staus is ok"})

	// Truncate JSON file/empty file contents and conversation history
	if err := os.Truncate(filename, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}
}

type Audio struct {
	File multipart.FileHeader `form:"file"`
}

func PostAudio(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")

	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Max-Age", "86400")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	c.Header("Access-Control-Allow-Credentials", "true")

	// Retrieve current message list from database
	db, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer db.Close()

	contents, err := ioutil.ReadAll(db)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	fmt.Println("contents: ", string(contents))
	fmt.Println("length of contents: ", len(contents))

	messageList := database.MessageList{}

	// If length of contents is 0 - DB is empty - load instruction
	// Else read current DB contents and unmarshal into struct
	if len(contents) == 0 {

		// Read the pre-existing entries in the conversation as stored in the database
		messageList = database.ReadDatabaseAndAppendInstruction()
	} else {
		err = json.Unmarshal(contents, &messageList)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
	}

	// Retrieve the uploaded file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save the file to a local temporary location
	dst := "temp.mp3"
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(dst)

	// Open the file
	fileToSend, err := os.Open(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer fileToSend.Close()

	// Create a buffer to store the form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the audio file to the form data
	part, err := writer.CreateFormFile("file", fileToSend.Name())
	if err != nil {
		fmt.Printf("Error creating form file: %s\n", err)
	}
	_, err = io.Copy(part, fileToSend)
	if err != nil {
		fmt.Printf("Error copying file to form data: %s\n", err)
	}

	// Add the model name to the form data
	modelName := "whisper-1"
	_ = writer.WriteField("model", modelName)

	// Set the content type header and close the writer
	err = writer.Close()
	if err != nil {
		fmt.Printf("Error closing form writer: %s\n", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		fmt.Printf("Error creating HTTP request: %s\n", err)
	}
	req.Header.Set("Authorization", "Bearer "+OPEN_AI_KEY)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the HTTP request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %s\n", err)
	}
	defer resp.Body.Close()

	// Read the HTTP response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading HTTP response body: %s\n", err)
	}

	whisperText := WhisperText{}
	json.Unmarshal(responseBody, &whisperText)

	// Store text received from Whisper in Messages struct
	message := database.Message{}
	message.Role = "user"
	message.Content = whisperText.Text

	//messageList := database.MessageList{}
	messageList.Messages = append(messageList.Messages, message)

	// Get chat response
	chatResponse := chatgpt.ChatGptResponse(OPEN_AI_KEY, OPEN_AI_ORG, messageList)

	// Determine the last element in the Messaage array and use that length to get the last message
	messagesLength := len(chatResponse.Messages) - 1

	//Append ChatGPT response to messages list
	message.Role = "assistant"
	message.Content = chatResponse.Messages[messagesLength].Content

	messageList.Messages = append(messageList.Messages, message)

	//Store messages
	fmt.Println("Message list before send to StoreMessagesInDB: ")
	fmt.Printf("%+v\n", messageList)
	database.StoreMessagesInDB(messageList)

	// Convert chat response to audio
	audioOutput := texttospeech.FetchAudioFile(ELEVEN_LABS_API_KEY, chatResponse.Messages[messagesLength].Content)

	// Set the appropriate content type
	c.Header("Content-Type", "video/mpeg")

	// Return the MPEG file data as the response
	c.Data(http.StatusOK, "video/mpeg", audioOutput)
}
