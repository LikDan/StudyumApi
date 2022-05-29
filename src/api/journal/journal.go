package journal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	h "studyum/src/api"
	"studyum/src/api/schedule"
	userApi "studyum/src/api/user"
	"studyum/src/db"
	"time"
)

func getLessonsDate(ctx *gin.Context) {
	var user userApi.User
	if err := userApi.GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	subject := ctx.Query("subject")
	var teacher string
	var group string

	if user.Type == "teacher" {
		group = ctx.Query("group")
		teacher = user.Name
	} else {
		group = user.Name
		teacher = ctx.Query("teacher")
	}

	var subjects []schedule.SubjectFull

	find, err := db.LessonsCollection.Find(nil, bson.M{"teacher": teacher, "group": group, "subject": subject})
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

func getJournal(ctx *gin.Context) {
	var user userApi.User
	if err := userApi.GetUserViaGoogle(ctx, &user); h.CheckAndMessage(ctx, 418, err, h.UNDEFINED) {
		return
	}

	var err error
	var journal *Journal

	if user.Type == "teacher" {
		err, journal = getJournalTeacher(user, ctx.Query("group"), ctx.Query("subject"))
	} else {
		err, journal = getJournalStudent(user)
	}

	if err != nil {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	ctx.JSON(200, journal)
}

func getJournalStudent(user userApi.User) (error, *Journal) {
	var journal Journal

	cursor, err := db.LessonsCollection.Aggregate(nil, bson.A{
		bson.M{"$match": bson.M{"group": user.TypeName, "studyPlaceId": user.StudyPlaceId}},
		bson.M{"$group": bson.M{"_id": "$subject"}},
		bson.M{"$lookup": bson.M{
			"from": "Lessons",
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"group": user.TypeName, "studyPlaceId": user.StudyPlaceId}},
				bson.M{"$group": bson.M{"_id": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}}},
				bson.M{"$sort": bson.M{"_id": 1}},
			},
			"as": "date",
		}},
		bson.M{"$unwind": "$date"},
		bson.M{"$addFields": bson.M{"date": "$date._id"}},
		bson.M{"$lookup": bson.M{
			"from": "Lessons",
			"let":  bson.M{"date": "$date", "subject": "$_id"},
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"group": user.TypeName, "studyPlaceId": user.StudyPlaceId}},
				bson.M{"$addFields": bson.M{"date_str": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}}},
				bson.M{"$lookup": bson.M{
					"from": "Marks",
					"let":  bson.M{"subjectId": "$_id"},
					"pipeline": bson.A{
						bson.M{"$match": bson.M{"studyPlaceId": user.StudyPlaceId, "userId": user.Id}},
						bson.M{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$subjectId", "$$subjectId"}}}},
					},
					"as": "marks",
				}},
				bson.M{"$unwind": bson.M{"path": "$marks", "preserveNullAndEmptyArrays": true}},
				bson.M{"$group": bson.M{"_id": bson.M{"date": "$date_str", "subject": "$subject"}, "lessons": bson.M{"$first": "$$ROOT"}, "marks": bson.M{"$push": "$marks"}}},
				bson.M{"$addFields": bson.M{"lessons.marks": "$marks"}},
				bson.M{"$project": bson.M{"marks": 0}},
				bson.M{"$match": bson.M{"$expr": bson.M{"$and": bson.A{bson.M{"$eq": bson.A{"$_id.date", "$$date"}}, bson.M{"$eq": bson.A{"$_id.subject", "$$subject"}}}}}},
			},
			"as": "subjects",
		}},
		bson.M{"$unwind": bson.M{"path": "$subjects", "preserveNullAndEmptyArrays": true}},
		bson.M{"$addFields": bson.M{"lesson": bson.M{"$ifNull": bson.A{"$subjects.lessons", nil}}}},
		bson.M{"$sort": bson.M{"date": 1}},
		bson.M{"$group": bson.M{"_id": "$_id", "title": bson.M{"$first": "$_id"}, "lessons": bson.M{"$push": "$lesson"}}},
		bson.M{"$sort": bson.M{"title": 1}},
		bson.M{"$group": bson.M{"_id": nil, "rows": bson.M{"$push": "$$ROOT"}}},
		bson.M{"$lookup": bson.M{
			"from": "Lessons",
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"group": user.TypeName, "studyPlaceId": user.StudyPlaceId}},
				bson.M{"$group": bson.M{"_id": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}}},
				bson.M{"$addFields": bson.M{"date": bson.M{"$toDate": "$_id"}}},
				bson.M{"$project": bson.M{"_id": 0}},
				bson.M{"$sort": bson.M{"date": 1}},
			},
			"as": "dates",
		}},
		bson.M{"$addFields": bson.M{
			"info": bson.M{
				"editable":     false,
				"studyPlaceId": user.StudyPlaceId,
				"group":        user.TypeName,
				"type":         "Student",
			},
		}},
		bson.M{"$project": bson.M{"_id": 0}},
	})
	if h.CheckError(err, h.WARNING) {
		return err, nil
	}

	cursor.Next(nil)
	err = cursor.Decode(&journal)
	if h.CheckError(err, h.WARNING) {
		return err, nil
	}

	return nil, &journal
}

func getJournalTeacher(user userApi.User, group, subject string) (error, *Journal) {
	if !h.SliceContains(user.Permissions, "editJournal") {
		return fmt.Errorf("no permission"), nil
	}

	if !h.CheckNotEmpty(group, subject) {
		return fmt.Errorf("provide valid params"), nil
	}

	cursor, err := db.UsersCollection.Aggregate(nil, mongo.Pipeline{
		bson.D{{"$match", bson.M{"type": "group", "name": group, "studyPlaceId": user.StudyPlaceId}}},
		bson.D{{"$lookup", bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": user.Name, "group": group, "studyPlaceId": user.StudyPlaceId}}}},
			"as":       "subjects",
		}}},
		bson.D{{"$unwind", "$subjects"}},
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "subjects._id",
			"foreignField": "subjectId",
			"let":          bson.M{"userId": "$_id"},
			"pipeline":     mongo.Pipeline{bson.D{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{"$userId", "$$userId"}}}}}},
			"as":           "subjects.marks",
		}}},
		bson.D{{"$sort", bson.M{"subjects.date": 1}}},
		bson.D{{"$addFields", bson.M{"userType": "student", "subjects.userId": "$_id"}}},
		bson.D{{"$group", bson.M{"_id": "$_id", "title": bson.M{"$first": "$fullName"}, "userType": bson.M{"$first": "$userType"}, "lessons": bson.M{"$push": "$subjects"}}}},
		bson.D{{"$sort", bson.M{"title": 1}}},
		bson.D{{"$group", bson.M{"_id": nil, "rows": bson.M{"$push": "$$ROOT"}}}},
		bson.D{{"$project", bson.M{"_id": 0}}},
		bson.D{{"$lookup", bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": user.Name, "group": group, "studyPlaceId": user.StudyPlaceId}}}},
			"as":       "dates",
		}}},
		bson.D{{"$addFields", bson.M{"info": bson.M{
			"editable":     true,
			"studyPlaceId": user.StudyPlaceId,
			"group":        group,
			"teacher":      user.Name,
			"subject":      subject,
		}}}},
	})
	if h.CheckError(err, h.WARNING) {
		return err, nil
	}

	var journal Journal

	cursor.Next(nil)
	err = cursor.Decode(&journal)
	if h.CheckError(err, h.WARNING) {
		return err, nil
	}

	return nil, &journal
}

type Info struct {
	Editable     bool   `json:"editable" bson:"editable"`
	StudyPlaceId int    `json:"studyPlaceId" bson:"studyPlaceId"`
	Group        string `json:"group" bson:"group"`
	Teacher      string `json:"teacher" bson:"teacher"`
	Subject      string `json:"subject" bson:"subject"`
}

type Row struct {
	Id       string    `json:"id" bson:"_id"`
	Title    string    `json:"title" bson:"title"`
	UserType string    `json:"userType" bson:"userType"`
	Lessons  []*Lesson `json:"lessons" bson:"lessons"`
}

type Journal struct {
	Info  Info     `json:"info" bson:"info"`
	Rows  []Row    `json:"rows" bson:"rows"`
	Dates []Lesson `json:"dates" bson:"dates"`
}

type FullStudentJournal struct {
	Subject string    `json:"subject" bson:"subject"`
	Group   string    `json:"group" bson:"group"`
	Lessons []*Lesson `json:"lessons" bson:"lessons"`
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
	EducationPlaceId int                `json:"educationPlaceId" bson:"studyPlaceId"`
	Date             time.Time          `json:"date"`
	Marks            []Mark             `json:"marks" bson:"marks"`
	Description      string             `json:"description" bson:"description"`
	Homework         string             `json:"homework" bson:"homework"`
	SmallDescription string             `json:"smallDescription" bson:"smallDescription"`
	UserId           primitive.ObjectID `json:"userId" bson:"userId"`
}

func getMarksViaId(userId primitive.ObjectID, id primitive.ObjectID) []Lesson {
	var marks []Lesson

	lessonsCursor, err := db.LessonsCollection.Aggregate(nil, mongo.Pipeline{
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

	match := bson.M{"group": group, "teacher": teacher, "subject": subject, "studyPlaceId": studyPlaceId}

	lessonsCursor, err := db.LessonsCollection.Aggregate(nil, mongo.Pipeline{
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

	err := db.LessonsCollection.FindOne(nil, bson.M{"_id": lessonId}).Decode(&subject)
	if err != nil {
		return nil, err
	}

	return &subject, nil
}

func BuildRequests(api *gin.RouterGroup) {
	teacherGroup := api.Group("/teachers")

	api.GET("/options", getAvailableOptions)
	api.GET("/dates", getLessonsDate)
	api.GET("", getJournal)

	teacherGroup.GET("/groupMembers", getGroupMembers)

	teacherGroup.POST("/mark", addMark)
	teacherGroup.GET("/mark", getMark)
	teacherGroup.PUT("/mark", editMark)
	teacherGroup.DELETE("/mark", removeMark)

	teacherGroup.PUT("/info", editInfo)
}
