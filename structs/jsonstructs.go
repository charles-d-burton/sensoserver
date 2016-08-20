package structs

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
