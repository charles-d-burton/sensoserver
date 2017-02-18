package workers

import (
	"log"

	"github.com/tidwall/gjson"
)

func ProcessIntentRequest(message string) {
	token := gjson.Get(message, "session.user.accessToken")
	log.Println(token)
}
