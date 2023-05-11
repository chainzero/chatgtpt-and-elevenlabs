package main

import (
	// "lambdaanddynamo/whisper"

	"lambdaanddynamo/restendpoints"
)

const (
	OPEN_AI_ORG         = "change_me"
	OPEN_AI_KEY         = "change_me"
	ELEVEN_LABS_API_KEY = "change_me"
)

func main() {
	restendpoints.StartRestEndpoints()
}
