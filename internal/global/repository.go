package global

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	GeneralLessonsCollection *mongo.Collection
	LessonsCollection        *mongo.Collection
	StudyPlacesCollection    *mongo.Collection
	UsersCollection          *mongo.Collection
	absencesCollection       *mongo.Collection
	SignUpCodesCollection    *mongo.Collection
}

func NewRepository(client *mongo.Client) *Repository {
	database := client.Database("Studyum")

	return &Repository{
		GeneralLessonsCollection: database.Collection("GeneralLessons"),
		LessonsCollection:        database.Collection("Lessons"),
		StudyPlacesCollection:    database.Collection("StudyPlaces"),
		UsersCollection:          database.Collection("Users"),
		absencesCollection:       database.Collection("Absences"),
		SignUpCodesCollection:    database.Collection("SignUpCodes"),
	}
}

func (r *Repository) GetUserByID(ctx context.Context, id primitive.ObjectID) (user User, err error) {
	err = r.UsersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return
}

func (r *Repository) SetRefreshToken(ctx context.Context, old string, session Session) error {
	_, err := r.UsersCollection.UpdateOne(ctx, bson.M{"sessions.refreshToken": old}, bson.M{"$set": bson.M{"sessions.$": session}})
	return err
}

func (r *Repository) GetStudyPlaceByApiToken(ctx context.Context, token string) (err error, studyPlace StudyPlace) {
	err = r.StudyPlacesCollection.FindOne(ctx, bson.M{"apiToken": token}).Decode(&studyPlace)
	return
}

func (r *Repository) GetUserViaRefreshToken(ctx context.Context, refreshToken string) (user User, err error) {
	err = r.UsersCollection.FindOne(ctx, bson.M{"sessions.refreshToken": refreshToken}).Decode(&user)
	return
}
