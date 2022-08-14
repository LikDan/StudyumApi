package oauth2

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"studyum/src/controllers"
	"studyum/src/db"
	"studyum/src/models"
	"studyum/src/utils"
	"time"
)

func OAuth2(ctx *gin.Context) {
	config := Configs[ctx.Param("oauth")]

	if config == nil {
		models.BindErrorStr("no such server", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	url := config.AuthCodeURL(ctx.Query("host"))
	ctx.Redirect(307, url)
}

func PutAuthToken(ctx *gin.Context) {
	bytes, _ := ctx.GetRawData()
	token := string(bytes)

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:    "authToken",
		Value:   token,
		Path:    "/",
		Expires: time.Now().AddDate(1, 0, 0),
	})

	var user models.User
	if err := controllers.AuthUserViaToken(token, &user); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, user)
}

func CallbackOAuth2(ctx *gin.Context) {
	token, err := googleOAuthConfig.Exchange(context.Background(), ctx.Query("code"))
	if models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		models.BindError(err, 500, models.WARNING).CheckAndResponse(ctx)
	}(response.Body)

	content, err := io.ReadAll(response.Body)
	if models.BindError(err, 418, models.WARNING).CheckAndResponse(ctx) {
		return
	}

	var googleUser models.OAuth2CallbackUser
	err = json.Unmarshal(content, &googleUser)
	if models.BindError(err, 418, models.WARNING).CheckAndResponse(ctx) {
		return
	}

	var user models.User

	if err = db.GetUserByEmail(ctx, googleUser.Email, &user).Error; err != nil {
		if err.Error() != "mongo: no documents in result" {
			models.BindError(err, 418, models.WARNING).CheckAndResponse(ctx)
			return
		}
		user = models.User{
			Id:            primitive.NewObjectID(),
			Token:         utils.GenerateSecureToken(),
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

		if db.SignUp(&user).CheckAndResponse(ctx) {
			return
		}
	}

	if user.Token == "" {
		user.Token = utils.GenerateSecureToken()

		if db.UpdateUserTokenByEmail(ctx, user.Email, user.Token).CheckAndResponse(ctx) {
			return
		}
	}

	ctx.Redirect(307, "http://"+ctx.Query("state")+"/user/receiveToken?token="+user.Token)
}
