package models

type OAuth2CallbackUser struct {
	Id            string `json:"id" bson:"_id"`
	Email         string `json:"email" bson:"email"`
	VerifiedEmail bool   `json:"verified_email" bson:"verifiedEmail"`
	Name          string `json:"name" bson:"login"`
	PictureUrl    string `json:"picture" bson:"picture"`
}
