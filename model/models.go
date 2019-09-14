package model

type User struct {
	Id   string
	Name string
}

type Participant struct {
	Id      string
	OwnerId string
	UserId  string
	Amount  float64
	Invoice string
	Status  string
}

type Meeting struct {
	Id           string
	OwnerId      string
	Amount       float64
	Status       string
}
