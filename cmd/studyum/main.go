package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/unrolled/secure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
	"net/http"
	"os"
	applications "studyum/internal/apps"
	"studyum/internal/auth"
	"studyum/internal/codes"
	"studyum/internal/general"
	"studyum/internal/journal"
	"studyum/internal/schedule"
	"studyum/internal/user"
	"studyum/internal/utils/middlewares"
	"studyum/pkg/encryption"
	"studyum/pkg/jwt"
	"studyum/pkg/mail"
	_ "studyum/pkg/validators"
	"time"
)

func main() {
	time.Local = time.FixedZone("GMT", 3*3600)

	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("DB_URL")))
	if err != nil {
		logrus.Fatal(err)
	}

	ctx := context.Background()
	if err = client.Connect(ctx); err != nil {
		logrus.Fatalf("Can't connect to database, error: %s", err.Error())
	}

	id := os.Getenv("GMAIL_CLIENT_ID")
	secret := os.Getenv("GMAIL_CLIENT_SECRET")
	access := os.Getenv("GMAIL_ACCESS_TOKEN")
	refresh := os.Getenv("GMAIL_REFRESH_TOKEN")
	mailer := mail.NewMail(context.Background(), mail.Mode(gin.Mode()), id, secret, access, refresh, "email-templates")

	mailer.ForceSend("likdan.official@gmail.com", "Application started", "Studyum app has been started")
	defer mailer.ForceSend("likdan.official@gmail.com", "Application stopped", "Studyum app has been stopped at"+time.Now().Format("2006-01-02 15:04"))

	defer logrus.Warning("Studyum is stopping at", time.Now().Format("2006-01-02 15:04"))

	//firebaseCredentials := []byte(os.Getenv("FIREBASE_CREDENTIALS"))
	//firebase := fb.NewFirebase(firebaseCredentials)
	encrypt := encryption.NewEncryption(os.Getenv("ENCRYPTION_SECRET"))

	engine := gin.New()
	config := cors.DefaultConfig()
	config.AddAllowHeaders("Access-Control-Allow-Origin", "Access-Control-Allow-Headers")
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	engine.Use(middlewares.ErrorMiddleware(), cors.New(config))

	engine.Any("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "uptime")
	})
	loadSwagger(engine.RouterGroup, "general", "auth", "user", "schedule", "journal")
	setupSSL(engine.RouterGroup)

	db := client.Database("Studyum")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_DB_URL"),
		Password: os.Getenv("REDIS_DB_PASSWORD"),
	})

	j := jwt.NewWithRedis("", time.Minute*15, time.Hour*24*30, time.Second*30, os.Getenv("JWT_SECRET"), redisClient)
	codesController := codes.New(time.Minute*15, time.Minute, mailer, db)

	api := engine.Group("/api")
	api.Use(gin.Logger(), gin.Recovery())

	apps := applications.New(db, encrypt)

	grpcServer := grpc.NewServer()
	authMiddleware, _, _ := auth.New(api.Group("/user"), grpcServer, codesController, encrypt, j, db)

	_, generalController := general.New(api, grpcServer, authMiddleware, db)
	_ = journal.New(api.Group("/journal"), authMiddleware, apps, encrypt, db)
	_ = schedule.New(api.Group("/schedule"), authMiddleware, apps, generalController, db)
	_ = user.New(api.Group("/user"), authMiddleware, encrypt, codesController, j, db)

	if gin.Mode() == gin.DebugMode {
		go launchGRPC(grpcServer)
		logrus.Fatalf("Error launching http server %s", engine.Run().Error())
		return
	}

	redirect := gin.New()
	redirect.Use(redirectFunc())

	go func() {
		logrus.Fatalf("Error launching redirect server %s", redirect.Run().Error())
	}()

	cert := "ssl/api_studyum_net.crt"
	key := "ssl/api_studyum_net.key"
	logrus.Fatalf("Error launching server %s", engine.RunTLS(":443", cert, key).Error())
}

func loadSwagger(e gin.RouterGroup, names ...string) {
	for _, name := range names {
		s := ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.InstanceName(name))
		e.GET("/swagger/"+name+"/*any", s)
	}
}

func setupSSL(e gin.RouterGroup) {
	e.GET(os.Getenv("SSL_DOMAIN_CONFIRMATION_URL"), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, os.Getenv("SSL_DOMAIN_CONFIRMATION_TEXT"))
	})
}

func redirectFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "api.studyum.net",
		})
		if err := secureMiddleware.Process(ctx.Writer, ctx.Request); err != nil {
			return
		}

		ctx.Next()
	}
}

func launchGRPC(server *grpc.Server) {
	listener, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))

	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	logrus.Fatalf("Error launching grpc server %s", server.Serve(listener))

}
