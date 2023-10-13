package entities

type Preferences struct {
	Theme    string `json:"theme" bson:"theme"`
	Language string `json:"language" bson:"language"`
	Timezone string `json:"timezone" bson:"timezone"`
}
