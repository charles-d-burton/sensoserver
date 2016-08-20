package structs

type Token struct {
	Token string `json:"token"`
}

type Reading struct {
	Reading string `json:"reading"`
	Time    string `json:"time"`
	Topic   string `json:"topic"`
}
