package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/entities"
)

type JournalRepository interface {
	AddMark(ctx context.Context, mark entities.Mark) (primitive.ObjectID, error)
	UpdateMark(ctx context.Context, mark entities.Mark) error
	GetMarkById(ctx context.Context, id primitive.ObjectID) (entities.Mark, error)
	DeleteMarkByID(ctx context.Context, id primitive.ObjectID) error

	GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]entities.JournalAvailableOption, error)

	GetStudentJournal(ctx context.Context, userId primitive.ObjectID, group string, studyPlaceId primitive.ObjectID) (entities.Journal, error)
	GetJournal(ctx context.Context, group string, subject string, typeName string, studyPlaceId primitive.ObjectID) (entities.Journal, error)

	GetLessonByID(ctx context.Context, id primitive.ObjectID) (entities.Lesson, error)
	GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId primitive.ObjectID) ([]entities.Lesson, error)
}

type journalRepository struct {
	*Repository
}

func NewJournalRepository(repository *Repository) JournalRepository {
	return &journalRepository{Repository: repository}
}

func (j *journalRepository) AddMark(ctx context.Context, mark entities.Mark) (primitive.ObjectID, error) {
	mark.Id = primitive.NewObjectID()
	if _, err := j.marksCollection.InsertOne(ctx, mark); err != nil {
		return primitive.NilObjectID, err
	}

	return mark.Id, nil
}

func (j *journalRepository) UpdateMark(ctx context.Context, mark entities.Mark) error {
	_, err := j.marksCollection.UpdateOne(ctx, bson.M{"_id": mark.Id, "lessonId": mark.LessonId}, bson.M{"$set": bson.M{"mark": mark.Mark}})
	return err
}

func (j *journalRepository) GetMarkById(ctx context.Context, id primitive.ObjectID) (entities.Mark, error) {
	var mark entities.Mark
	err := j.marksCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&mark)
	return mark, err
}

func (j *journalRepository) DeleteMarkByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := j.marksCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (j *journalRepository) GetAvailableOptions(ctx context.Context, teacher string, editable bool) ([]entities.JournalAvailableOption, error) {
	aggregate, err := j.lessonsCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"teacher": teacher}},
		bson.M{"$group": bson.M{
			"_id": bson.M{
				"teacher": "$teacher",
				"subject": "$subject",
				"group":   "$group",
			},
			"teacher": bson.M{"$first": "$teacher"},
			"subject": bson.M{"$first": "$subject"},
			"group":   bson.M{"$first": "$group"}},
		},
		bson.M{"$addFields": bson.M{"editable": editable}},
	})
	if err != nil {
		return nil, err
	}

	var options []entities.JournalAvailableOption
	if err = aggregate.All(ctx, &options); err != nil {
		return nil, err
	}

	return options, nil
}

func (j *journalRepository) GetStudentJournal(ctx context.Context, userId primitive.ObjectID, group string, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	cursor, err := j.lessonsCollection.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"group": group, "studyPlaceId": studyPlaceId}},
		bson.M{"$group": bson.M{"_id": "$subject"}},
		bson.M{"$lookup": bson.M{
			"from": "Lessons",
			"pipeline": bson.A{
				bson.M{"$match": bson.M{"group": group, "studyPlaceId": studyPlaceId}},
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
				bson.M{"$match": bson.M{"group": group, "studyPlaceId": studyPlaceId}},
				bson.M{"$addFields": bson.M{"date_str": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}}},
				bson.M{"$lookup": bson.M{
					"from": "Marks",
					"let":  bson.M{"subjectId": "$_id"},
					"pipeline": bson.A{
						bson.M{"$match": bson.M{"studyPlaceId": studyPlaceId, "userId": userId}},
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
				bson.M{"$match": bson.M{"group": group, "studyPlaceId": studyPlaceId}},
				bson.M{"$group": bson.M{"_id": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date"}}}},
				bson.M{"$addFields": bson.M{"lesson": bson.M{"endDate": bson.M{"$toDate": "$_id"}}, "startDate": bson.M{"$toDate": "$_id"}}},
				bson.M{"$project": bson.M{"_id": 0}},
				bson.M{"$sort": bson.M{"date": 1}},
			},
			"as": "dates",
		}},
		bson.M{"$addFields": bson.M{
			"info": bson.M{
				"editable":     false,
				"studyPlaceId": studyPlaceId,
				"group":        group,
				"type":         "Student",
			},
		}},
		bson.M{"$project": bson.M{"_id": 0}},
	})
	if err != nil {
		return entities.Journal{}, err
	}

	cursor.Next(ctx)
	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *journalRepository) GetJournal(ctx context.Context, group string, subject string, typeName string, studyPlaceId primitive.ObjectID) (entities.Journal, error) {
	cursor, err := j.usersCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"type": "group", "typeName": group, "studyPlaceId": studyPlaceId}}},
		bson.D{{"$lookup", bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": typeName, "group": group, "studyPlaceId": studyPlaceId}}}},
			"as":       "subjects",
		}}},
		bson.D{{"$unwind", "$subjects"}},
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "subjects._id",
			"foreignField": "lessonId",
			"let":          bson.M{"userId": "$_id"},
			"pipeline":     mongo.Pipeline{bson.D{{"$match", bson.M{"$expr": bson.M{"$eq": bson.A{"$userId", "$$userId"}}}}}},
			"as":           "subjects.marks",
		}}},
		bson.D{{"$sort", bson.M{"subjects.date": 1}}},
		bson.D{{"$addFields", bson.M{"userType": "student", "subjects.userId": "$_id"}}},
		bson.D{{"$group", bson.M{"_id": "$_id", "title": bson.M{"$first": "$name"}, "userType": bson.M{"$first": "$userType"}, "lessons": bson.M{"$push": "$subjects"}}}},
		bson.D{{"$sort", bson.M{"title": 1}}},
		bson.D{{"$group", bson.M{"_id": nil, "rows": bson.M{"$push": "$$ROOT"}}}},
		bson.D{{"$project", bson.M{"_id": 0}}},
		bson.D{{"$lookup", bson.M{
			"from":     "Lessons",
			"pipeline": mongo.Pipeline{bson.D{{"$match", bson.M{"subject": subject, "teacher": typeName, "group": group, "studyPlaceId": studyPlaceId}}}},
			"as":       "dates",
		}}},
		bson.D{{"$addFields", bson.M{"info": bson.M{
			"editable":     true,
			"studyPlaceId": studyPlaceId,
			"group":        group,
			"teacher":      typeName,
			"subject":      subject,
		}}}},
	})
	if err != nil {
		return entities.Journal{}, err
	}

	cursor.Next(ctx)
	var journal entities.Journal
	if err = cursor.Decode(&journal); err != nil {
		return entities.Journal{}, err
	}

	return journal, nil
}

func (j *journalRepository) GetLessonByID(ctx context.Context, id primitive.ObjectID) (lesson entities.Lesson, err error) {
	err = j.lessonsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&lesson)
	return
}

func (j *journalRepository) GetLessons(ctx context.Context, userId primitive.ObjectID, group, teacher, subject string, studyPlaceId primitive.ObjectID) ([]entities.Lesson, error) {
	lessonsCursor, err := j.lessonsCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$lookup", bson.M{
			"from":         "Marks",
			"localField":   "_id",
			"foreignField": "lessonId",
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"userId": userId}}},
			},
			"as": "marks",
		}}},
		bson.D{{"$match", bson.M{"group": group, "teacher": teacher, "subject": subject, "studyPlaceId": studyPlaceId}}},
		bson.D{{"$sort", bson.M{"date": 1}}},
	})

	var marks []entities.Lesson
	if err = lessonsCursor.All(ctx, &marks); err != nil {
		return nil, err
	}

	return marks, nil
}
