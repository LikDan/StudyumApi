package entities

type JournalUser struct {
	ID       string `bson:"_id"`
	Login    string `bson:"login"`
	Password string `bson:"password"`
}
