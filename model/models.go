package model

type User struct {
	Id   string `validate:"required,gte=1"`
	Name string `validate:"required,gte=1"`
}

type Participant struct {
	UserName string
	Amount   float64
	Invoice  string
	State    string
}

type Meeting struct {
	Id      string
	Name    string
	Date    string
	OwnerId string
	Amount  string
	State   string
}
