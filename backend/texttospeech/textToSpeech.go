package texttospeech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func FetchAudioFile(ELEVEN_LABS_API_KEY, chatResponse string) []byte {

	//message := "Hello world"

	fmt.Println("chatResponse in texttospeech: ", chatResponse)

	body := map[string]interface{}{
		"text": chatResponse,
		"voice_settings": map[string]interface{}{
			"stability":        0,
			"similarity_boost": 0,
		},
	}

	//voice_shaun := "mTSvIrm2hmcnOvb21nW2"
	voice_rachel := "21m00Tcm4TlvDq8ikWAM"
	//voice_antoni := "ErXwobaYiN019PkySvjV"

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voice_rachel)

	jsonStr, err := json.Marshal(body)
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", ELEVEN_LABS_API_KEY)
	req.Header.Set("accept", "audio/mpeg")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	if resp.StatusCode == 200 {
		// Do something with audio response
		// fmt.Println("Audio response:", respBody)
		return respBody
	} else {
		fmt.Println("Failed to get Eleven Labs audio response")
	}
	return []byte("Error in ElevenLabs request")
}
