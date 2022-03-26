package journal

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	h "studyium/src/api"
	"studyium/src/api/schedule"
	userApi "studyium/src/api/user"
	"studyium/src/db"
	"time"
)

func getLessonsDate(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	subject := ctx.Query("subject")
	var teacher string
	var group string

	if user.Type == "teacher" {
		group = ctx.Query("group")
		teacher = user.FullName
	} else {
		group = user.Name
		teacher = ctx.Query("teacher")
	}

	var subjects []schedule.SubjectFull

	find, err := db.SubjectsCollection.Find(nil, bson.M{"teacher": teacher, "group": group, "subject": subject})
	err = find.All(nil, &subjects)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if subjects == nil {
		h.ErrorMessage(ctx, "no such subjects")
		return
	}

	ctx.JSON(200, subjects)
}

func getStudentJournal(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if user.Type != "group" {
		h.ErrorMessage(ctx, "not a student")
		return
	}

	var subjects []FullStudentJournal

	find, err := db.SubjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "_id",
			"foreignField": "subjectId",
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"userId": user.Id}}},
			},
			"as": "marks",
		}}},
		bson.D{{"$match", bson.M{"group": user.Name}}},
		bson.D{{"$sort", bson.M{"lessons.date": 1}}},
		bson.D{{"$group", bson.M{
			"_id":     bson.M{"subject": "$subject", "group": "$group"},
			"subject": bson.M{"$first": "$subject"},
			"group":   bson.M{"$first": "$group"},
			"lessons": bson.M{"$addToSet": bson.M{
				"_id":          "$_id",
				"subject":      "$subject",
				"group":        "$group",
				"teacher":      "$teacher",
				"room":         "$room",
				"marks":        "$marks",
				"date":         "$date",
				"rowIndex":     "$rowIndex",
				"columnIndex":  "$columnIndex",
				"weekIndex":    "$weekIndex",
				"type":         "$type",
				"studyPlaceId": "$educationPlaceId",
			}},
		}}},
	})
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	err = find.All(nil, &subjects)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, subjects)
}

type FullStudentJournal struct {
	Subject string   `json:"subject" bson:"subject"`
	Group   string   `json:"group" bson:"group"`
	Lessons []Lesson `json:"lessons" bson:"lessons"`
}

type Mark struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Mark         string             `json:"mark" bson:"mark"`
	UserId       primitive.ObjectID `json:"userId" bson:"userId"`
	SubjectId    primitive.ObjectID `json:"subjectId" bson:"subjectId"`
	StudyPlaceId int                `json:"studyPlaceId" bson:"studyPlaceId"`
}

type Lesson struct {
	Id               primitive.ObjectID `json:"id" bson:"_id"`
	Subject          string             `json:"subject"`
	Teacher          string             `json:"teacher"`
	Group            string             `json:"group"`
	Room             string             `json:"room"`
	ColumnIndex      int                `json:"columnIndex" bson:"columnIndex"`
	RowIndex         int                `json:"rowIndex" bson:"rowIndex"`
	WeekIndex        int                `json:"weekIndex" bson:"weekIndex"`
	Type_            string             `json:"type" bson:"type"`
	EducationPlaceId int                `json:"educationPlaceId" bson:"educationPlaceId"`
	Date             time.Time          `json:"date"`
	Marks            []Mark             `json:"marks" bson:"marks"`
}

func getMarksViaId(userId primitive.ObjectID, id primitive.ObjectID) []Lesson {
	var marks []Lesson

	lessonsCursor, err := db.SubjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "_id",
			"foreignField": "subjectId",
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"userId": userId}}},
			},
			"as": "marks",
		}}},
		bson.D{{"$match", bson.M{"_id": id}}},
		bson.D{{"$sort", bson.M{"date": 1}}},
	})

	err = lessonsCursor.All(nil, &marks)
	if h.CheckError(err, h.WARNING) {
		return marks
	}

	return marks
}

func getMarks(userId primitive.ObjectID, group, teacher, subject string, studyPlaceId int) []Lesson {
	var marks []Lesson

	match := bson.M{"group": group, "teacher": teacher, "subject": subject, "educationPlaceId": studyPlaceId}

	lessonsCursor, err := db.SubjectsCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "_id",
			"foreignField": "subjectId",
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"userId": userId}}},
			},
			"as": "marks",
		}}},
		bson.D{{"$match", match}},
		bson.D{{"$sort", bson.M{"date": 1}}},
	})

	err = lessonsCursor.All(nil, &marks)
	if h.CheckError(err, h.WARNING) {
		return marks
	}

	return marks
}

func getLesson(lessonId *primitive.ObjectID) (*schedule.SubjectFull, error) {
	var subject schedule.SubjectFull

	err := db.SubjectsCollection.FindOne(nil, bson.M{"_id": lessonId}).Decode(&subject)
	if err != nil {
		return nil, err
	}

	return &subject, nil
}

func BuildRequests(api *gin.RouterGroup) {
	teacherGroup := api.Group("/teachers")

	api.GET("/availableOptions", getAvailableOptions)
	api.GET("/dates", getLessonsDate)
	api.GET("/studentJournal", getStudentJournal)

	teacherGroup.GET("/groupMembers", getGroupMembers)

	teacherGroup.POST("/mark", addMark)
	teacherGroup.GET("/mark", getMark)
	teacherGroup.PUT("/mark", editMark)
	teacherGroup.DELETE("/mark", removeMark)

	teacherGroup.GET("/editInfo", editInfo)
}
