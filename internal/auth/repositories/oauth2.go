package repositories

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"studyum/internal/auth/entities"
)

type OAuth2 interface {
	GetService(ctx context.Context, service string) (entities.OAuth2ServiceRaw, error)
	GetCallbackUser(ctx context.Context, url string) (entities.OAuth2CallbackUser, error)

	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
	SignUp(ctx context.Context, user entities.User) error
}

type oauth2 struct {
	services *mongo.Collection
	users    *mongo.Collection
}

func NewOAuth2(services *mongo.Collection, users *mongo.Collection) OAuth2 {
	return &oauth2{services: services, users: users}
}

func (r *oauth2) GetService(ctx context.Context, serviceName string) (service entities.OAuth2ServiceRaw, err error) {
	err = r.services.FindOne(ctx, bson.M{"_id": serviceName}).Decode(&service)
	return
}

func (r *oauth2) GetCallbackUser(ctx context.Context, url string) (user entities.OAuth2CallbackUser, err error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}

	err = json.NewDecoder(response.Body).Decode(&user)
	return
}

func (r *oauth2) GetUserByEmail(ctx context.Context, email string) (user entities.User, err error) {
	err = r.users.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	return
}

func (r *oauth2) SignUp(ctx context.Context, user entities.User) error {
	_, err := r.users.InsertOne(ctx, user)
	return err
}
