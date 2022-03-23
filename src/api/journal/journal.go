package journal

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	h "studyium/api"
	"studyium/api/schedule"
	userApi "studyium/api/user"
	"studyium/db"
	"time"
)

func getTeacherJournalSubjects(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	group := ctx.Query("group")
	subject := ctx.Query("subject")

	var subjects []schedule.SubjectFull

	find, err := db.SubjectsCollection.Find(nil, bson.M{"teacher": user.FullName, "group": group, "subject": subject})
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

func getTeacherJournalTypes(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	find, err := db.GeneralSubjectsCollection.Find(nil, bson.M{"teacher": user.FullName})
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	var subjects []schedule.SubjectFull
	err = find.All(nil, &subjects)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	var types []JournalTeacherType

	for _, subject := range subjects {
		type_ := JournalTeacherType{
			Teacher: subject.Teacher,
			Subject: subject.Subject,
			Group:   subject.Group,
		}

		if h.SliceContains(types, type_) {
			continue
		}

		types = append(types, type_)
	}

	ctx.JSON(200, types)
}

func addMark(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}
	if !h.SliceContains(user.Permissions, "editJournal") {
		h.ErrorMessage(ctx, "no permission")
		return
	}

	var mark Mark
	err = ctx.BindJSON(&mark)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if mark.Mark == "" || mark.UserId.IsZero() || mark.SubjectId.IsZero() {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	mark.Id = primitive.NewObjectID()

	_, err = db.MarksCollection.InsertOne(nil, mark)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	marks := getMarksViaId(mark.UserId, mark.SubjectId)

	if len(marks) != 1 {
		h.ErrorMessage(ctx, "wrong response")
	}

	ctx.JSON(200, marks[0])
}

func getMark(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")
	teacher := user.FullName

	if group == "" || subject == "" || userIdHex == "" {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	marks := getMarks(userId, group, teacher, subject, user.StudyPlaceId)

	ctx.JSON(200, marks)
}

func editMark(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}
	if !h.SliceContains(user.Permissions, "editJournal") {
		h.ErrorMessage(ctx, "no permission")
		return
	}

	var mark Mark
	err = ctx.BindJSON(&mark)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if mark.Mark == "" || mark.Id.IsZero() || mark.UserId.IsZero() || mark.SubjectId.IsZero() {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	_, err = db.MarksCollection.UpdateOne(nil, bson.M{"_id": mark.Id, "subjectId": mark.SubjectId}, bson.M{"$set": bson.M{"mark": mark.Mark}})
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	marks := getMarksViaId(mark.UserId, mark.SubjectId)

	if len(marks) != 1 {
		h.ErrorMessage(ctx, "wrong response")
	}

	ctx.JSON(200, marks[0])
}

func removeMark(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}
	if !h.SliceContains(user.Permissions, "editJournal") {
		h.ErrorMessage(ctx, "no permission")
		return
	}

	markId := ctx.Query("markId")
	userIdHex := ctx.Query("userId")
	subjectId := ctx.Query("subjectId")

	if markId == "" || userIdHex == "" || subjectId == "" {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	markObjectId, err := primitive.ObjectIDFromHex(markId)

	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	subjectObjectId, err := primitive.ObjectIDFromHex(subjectId)

	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	_, err = db.MarksCollection.DeleteOne(nil, bson.M{"_id": markObjectId, "subjectId": subjectObjectId})
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	marks := getMarksViaId(userId, subjectObjectId)

	if len(marks) != 1 {
		h.ErrorMessage(ctx, "wrong response")
	}

	ctx.JSON(200, marks[0])
}

func getGroupMembers(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if !h.SliceContains(user.Permissions, "editJournal") {
		h.ErrorMessage(ctx, "no permission")
		return
	}

	group := ctx.Query("group")
	var members []userApi.User

	groupsCursor, err := db.UsersCollection.Find(nil, bson.M{"studyPlaceId": user.StudyPlaceId, "type": "group", "name": group})
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}
	err = groupsCursor.All(nil, &members)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, members)
}

func editInfo(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if !h.SliceContains(user.Permissions, "editJournal") {
		h.ErrorMessage(ctx, "no permission")
		return
	}

	lessonId := h.GetObjectId(ctx, "lessonId")
	homework := ctx.Query("homework")
	smallDescription := ctx.Query("smallDescription")
	description := ctx.Query("description")

	if lessonId == nil {
		h.ErrorMessage(ctx, "provide valid params")
		return
	}

	_, err = db.SubjectsCollection.UpdateOne(nil, bson.M{"_id": lessonId}, bson.M{"$set": bson.M{"homework": homework, "smallDescription": smallDescription, "description": description}})
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	lesson, err := getLesson(lessonId)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, lesson)
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

func getMarksViaId(userId primitive.ObjectID, id primitive.ObjectID) []MarkFull {
	var marks []MarkFull

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

func getMarks(userId primitive.ObjectID, group, teacher, subject string, studyPlaceId int) []MarkFull {
	var marks []MarkFull

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
	api.GET("/types", getTeacherJournalTypes)
	api.GET("/dates", getTeacherJournalSubjects)
	api.GET("/groupMembers", getGroupMembers)

	api.POST("/mark", addMark)
	api.GET("/mark", getMark)
	api.PUT("/mark", editMark)
	api.DELETE("/mark", removeMark)

	api.GET("/editInfo", editInfo)
}
