package oauth2

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"io/ioutil"
	"net/http"
	h "studyum/src/api"
	"studyum/src/db"
	"studyum/src/models"
	"time"
)

func OAuth2(ctx *gin.Context) {
	config := Configs[ctx.Param("oauth")]

	if config == nil {
		models.BindErrorStr("no such server", 400, h.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	url := config.AuthCodeURL(ctx.Query("redirect"))
	ctx.Redirect(307, url)
}

func CallbackOAuth2(ctx *gin.Context) {
	token, err := googleOAuthConfig.Exchange(context.Background(), ctx.Query("code"))
	if models.BindError(err, 400, h.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if models.BindError(err, 400, h.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		h.CheckError(err, h.WARNING)
	}(response.Body)

	content, err := ioutil.ReadAll(response.Body)
	if models.BindError(err, 418, h.WARNING).CheckAndResponse(ctx) {
		return
	}

	var googleUser models.OAuth2CallbackUser
	err = json.Unmarshal(content, &googleUser)
	if models.BindError(err, 418, h.WARNING).CheckAndResponse(ctx) {
		return
	}

	var user models.User
	if err = db.UsersCollection.FindOne(ctx, bson.M{"email": googleUser.Email}).Decode(&user); err != nil {
		if err.Error() == "mongo: no documents in result" {
			user = models.User{
				Id:            primitive.NewObjectID(),
				Token:         h.GenerateSecureToken(),
				Email:         googleUser.Email,
				VerifiedEmail: googleUser.VerifiedEmail,
				Login:         googleUser.Name,
				Name:          googleUser.Name,
				PictureUrl:    googleUser.PictureUrl,
				Type:          "",
				TypeName:      "",
				StudyPlaceId:  0,
				Permissions:   nil,
				Accepted:      false,
				Blocked:       false,
			}
			if _, err = db.UsersCollection.InsertOne(ctx, user); models.BindError(err, 418, h.WARNING).CheckAndResponse(ctx) {
				return
			}
		} else {
			models.BindError(err, 418, h.WARNING).CheckAndResponse(ctx)
			return
		}
	}

	if user.Token == "" {
		user.Token = h.GenerateSecureToken()
		if _, err = db.UsersCollection.UpdateOne(ctx, bson.M{"email": user.Email}, bson.M{"$set": bson.M{"token": user.Token}}); err != nil {
			models.BindError(err, 418, h.WARNING).CheckAndResponse(ctx)
			return
		}
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   user.Token,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})
	ctx.Redirect(307, "http://"+ctx.Query("state"))
}
