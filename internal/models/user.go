package models

type User struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
	Email string `json:"email"`
}
