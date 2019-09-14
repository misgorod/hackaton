package model

import "time"

type User struct {
	Id   string `validate:"required,gte=1"`
	Name string `validate:"required,gte=1"`
}

type Participant struct {
	UserId   string
	UserName string
	Amount   float64
	Invoice  string
	Status   string
}

type Meeting struct {
	Id      string
	Name    string
	Date    time.Time
	OwnerId string
	Amount  float64
	Status  string
}
