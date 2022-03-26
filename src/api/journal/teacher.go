package journal

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	h "studyium/api"
	userApi "studyium/api/user"
	"studyium/db"
)

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
