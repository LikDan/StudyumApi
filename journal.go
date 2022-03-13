package main

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func getTeacherJournalSubjects(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	group := ctx.Query("group")
	subject := ctx.Query("subject")

	var subjects []SubjectFull

	find, err := subjectsCollection.Find(nil, bson.M{"teacher": user.FullName, "group": group, "subject": subject})
	err = find.All(nil, &subjects)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	if subjects == nil {
		errorMessage(ctx, "no such subjects")
		return
	}

	ctx.JSON(200, subjects)
}

func getTeacherJournalTypes(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	find, err := generalSubjectsCollection.Find(nil, bson.M{"teacher": user.FullName})
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	var subjects []SubjectFull
	err = find.All(nil, &subjects)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	var types []JournalTeacherType

	for _, subject := range subjects {
		type_ := JournalTeacherType{
			Teacher: subject.Teacher,
			Subject: subject.Subject,
			Group:   subject.Group,
		}

		if sliceContains(types, type_) {
			continue
		}

		types = append(types, type_)
	}

	ctx.JSON(200, types)
}

func addMark(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}
	if !sliceContains(user.Permissions, "editJournal") {
		errorMessage(ctx, "no permission")
		return
	}

	mark_ := ctx.Query("mark")
	userId := ctx.Query("userId")
	subjectId := ctx.Query("subjectId")

	if mark_ == "" || userId == "" || subjectId == "" {
		errorMessage(ctx, "provide valid params")
		return
	}

	userObjectId, err := primitive.ObjectIDFromHex(userId)

	if checkError(err) {
		errorMessage(ctx, "provide valid params")
		return
	}

	subjectObjectId, err := primitive.ObjectIDFromHex(subjectId)

	if checkError(err) {
		errorMessage(ctx, "provide valid params")
		return
	}

	mark := Mark{
		Id:           primitive.NewObjectID(),
		Mark:         mark_,
		SubjectId:    subjectObjectId,
		UserId:       userObjectId,
		StudyPlaceId: user.StudyPlaceId,
	}

	_, err = marksCollection.InsertOne(nil, mark)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, mark)
}

func getMark(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")
	teacher := user.FullName

	if group == "" || subject == "" || userIdHex == "" {
		errorMessage(ctx, "provide valid params")
		return
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	marks := getMarks(userId, group, teacher, subject, user.StudyPlaceId, nil)

	ctx.JSON(200, marks)
}

func editMark(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}
	if !sliceContains(user.Permissions, "editJournal") {
		errorMessage(ctx, "no permission")
		return
	}

	mark_ := ctx.Query("mark")
	markId := ctx.Query("markId")
	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")
	subjectId := ctx.Query("subjectId")
	teacher := user.FullName

	if mark_ == "" || markId == "" || group == "" || subject == "" || userIdHex == "" || subjectId == "" {
		errorMessage(ctx, "provide valid params")
		return
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	markObjectId, err := primitive.ObjectIDFromHex(markId)

	if checkError(err) {
		errorMessage(ctx, "provide valid params")
		return
	}

	subjectObjectId, err := primitive.ObjectIDFromHex(subjectId)

	if checkError(err) {
		errorMessage(ctx, "provide valid params")
		return
	}

	_, err = marksCollection.UpdateOne(nil, bson.M{"_id": markObjectId, "subjectId": subjectObjectId}, bson.M{"$set": bson.M{"mark": mark_}})
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	marks := getMarks(userId, group, teacher, subject, user.StudyPlaceId, &subjectObjectId)

	if len(marks) != 1 {
		errorMessage(ctx, "wrong response")
	}

	ctx.JSON(200, marks[0])
}

func getGroupMembers(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	if !sliceContains(user.Permissions, "editJournal") {
		errorMessage(ctx, "no permission")
		return
	}

	group := ctx.Query("group")
	var members []User

	groupsCursor, err := usersCollection.Find(nil, bson.M{"studyPlaceId": user.StudyPlaceId, "type": "group", "name": group})
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}
	err = groupsCursor.All(nil, &members)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, members)
}

type JournalTeacherType struct {
	Teacher string `json:"teacher"`
	Subject string `json:"subject"`
	Group   string `json:"group"`
}

type Mark struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Mark         string             `json:"mark" bson:"mark"`
	UserId       primitive.ObjectID `json:"userId" bson:"userId"`
	SubjectId    primitive.ObjectID `json:"subjectId" bson:"subjectId"`
	StudyPlaceId int                `json:"studyPlaceId" bson:"studyPlaceId"`
}

type MarkFull struct {
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

func getMarks(userId primitive.ObjectID, group, teacher, subject string, studyPlaceId int, id *primitive.ObjectID) []MarkFull {
	var marks []MarkFull

	match := bson.M{"group": group, "teacher": teacher, "subject": subject, "educationPlaceId": studyPlaceId}

	if id != nil {
		match["_id"] = id
	}

	lessonsCursor, err := subjectsCollection.Aggregate(nil, mongo.Pipeline{
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
	if checkError(err) {
		return marks
	}

	return marks
}
