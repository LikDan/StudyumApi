package entities

import "time"

type VerificationCode struct {
	Code      string    `json:"code" bson:"code"`
	Email     string    `json:"email" bson:"email"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
