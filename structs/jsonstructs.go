package structs

type Payload struct {
	Reading Reading
}

type Token struct {
	Token string `json:"token"`
}

type Topic struct {
	Token string `json:"token"`
	Topic string `json:"topic"`
}

type Reading struct {
	Reading string `json:"reading"`
	Time    string `json:"time"`
	Topic   string `json:"topic"`
}
type Message struct {
	Token   string `json:"token"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}
