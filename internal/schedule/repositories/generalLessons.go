package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/schedule/entities"
	"studyum/internal/utils"
)

type GeneralLessonsRepository = utils.ICRUDRepositoryWithStudyPlaceID[entities.GeneralLesson, primitive.ObjectID]

type generalLessonsRepository struct {
	generalLessons *mongo.Collection
}

func NewGeneralLessonsRepository(generalLessons *mongo.Collection) GeneralLessonsRepository {
	return &generalLessonsRepository{generalLessons: generalLessons}
}

func (s *generalLessonsRepository) GetByID(ctx context.Context, studyPlaceID primitive.ObjectID, id primitive.ObjectID) (lesson entities.GeneralLesson, err error) {
	cursor, err := s.generalLessons.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"_id": id, "studyPlaceID": studyPlaceID}},
		bson.M{"$lookup": bson.M{"from": "StudyPlaceUsers", "localField": "teacherID", "foreignField": "_id", "as": "teacher"}},
		bson.M{"$lookup": bson.M{"from": "Groups", "localField": "groupID", "foreignField": "_id", "as": "group"}},
		bson.M{"$lookup": bson.M{"from": "Subjects", "localField": "subjectID", "foreignField": "_id", "as": "subject"}},
		bson.M{"$lookup": bson.M{"from": "Rooms", "localField": "roomID", "foreignField": "_id", "as": "room"}},
		bson.M{
			"$addFields": bson.M{
				"subject": bson.M{"$first": "$subject.subject"},
				"room":    bson.M{"$first": "$room.room"},
				"teacher": bson.M{"$first": "$teacher.roleName"},
				"group":   bson.M{"$first": "$group.group"},
			},
		},
	})

	if err != nil {
		return entities.GeneralLesson{}, err
	}

	cursor.Next(ctx)
	err = cursor.Decode(&lesson)
	return
}

func (s *generalLessonsRepository) Add(ctx context.Context, lesson entities.GeneralLesson) error {
	_, err := s.generalLessons.InsertOne(ctx, lesson)
	return err
}

func (s *generalLessonsRepository) Update(ctx context.Context, studyPlaceID primitive.ObjectID, lesson entities.GeneralLesson) error {
	_, err := s.generalLessons.UpdateOne(ctx, bson.M{"_id": lesson.Id, "studyPlaceID": studyPlaceID}, bson.M{"$set": bson.M{
		"primaryColor":     lesson.PrimaryColor,
		"secondaryColor":   lesson.SecondaryColor,
		"startTimeMinutes": lesson.StartTimeMinutes,
		"endTimeMinutes":   lesson.EndTimeMinutes,
		"subjectID":        lesson.SubjectID,
		"groupID":          lesson.GroupID,
		"teacherID":        lesson.TeacherID,
		"roomID":           lesson.RoomID,
		"weekIndex":        lesson.WeekIndex,
		"dayIndex":         lesson.DayIndex,
		"lessonIndex":      lesson.LessonIndex,
	}})
	return err
}

func (s *generalLessonsRepository) DeleteByID(ctx context.Context, studyPlaceID primitive.ObjectID, id primitive.ObjectID) error {
	_, err := s.generalLessons.DeleteMany(ctx, bson.M{"_id": id, "studyPlaceID": studyPlaceID})
	return err
}
