package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"studyum/internal/parser/entities"
	"studyum/internal/utils"
	"time"
)

type Repository interface {
	GetUsersToParse(ctx context.Context, parserAppName string) ([]entities.JournalUser, error)
	UpdateParseJournalUser(ctx context.Context, user entities.JournalUser) error

	InsertScheduleTypes(ctx context.Context, types []entities.ScheduleTypeInfo) error
	GetScheduleTypesToParse(ctx context.Context, parserAppName string) ([]entities.ScheduleTypeInfo, error)

	UpdateGeneralSchedule(ctx context.Context, lessons []entities.GeneralLesson) error
	GetLessonIDByDateNameAndGroup(ctx context.Context, date time.Time, name string, group string) (primitive.ObjectID, error)
	GetLastLesson(ctx context.Context, studyPlaceId int) (entities.Lesson, error)
	AddLessons(ctx context.Context, lessons []entities.Lesson) error

	AddMarks(ctx context.Context, marks []entities.Mark) error
}

type repository struct {
	generalLessonsCollection *mongo.Collection
	lessonsCollection        *mongo.Collection
	marksCollection          *mongo.Collection

	parseJournalUserCollection   *mongo.Collection
	parseScheduleTypesCollection *mongo.Collection
}

func NewParserRepository(client *mongo.Client) Repository {
	database := client.Database("Schedule")

	return &repository{
		generalLessonsCollection: database.Collection("GeneralLessons"),
		lessonsCollection:        database.Collection("Lessons"),
		marksCollection:          database.Collection("Marks"),

		parseJournalUserCollection:   database.Collection("ParseJournalUsers"),
		parseScheduleTypesCollection: database.Collection("ParseScheduleTypes"),
	}
}

func (p *repository) GetUsersToParse(ctx context.Context, parserAppName string) ([]entities.JournalUser, error) {
	result, err := p.parseJournalUserCollection.Find(ctx, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return nil, err
	}

	var users []entities.JournalUser
	if err = result.All(ctx, users); err != nil {
		return nil, err
	}

	return users, nil
}

func (p *repository) InsertScheduleTypes(ctx context.Context, types []entities.ScheduleTypeInfo) error {
	if len(types) == 0 {
		return errors.New("empty types array")
	}

	if _, err := p.parseScheduleTypesCollection.DeleteMany(ctx, bson.M{"parserAppName": types[0].ParserAppName}); err != nil {
		return err
	}

	for _, type_ := range types {
		type_.Id = primitive.NewObjectID()
	}

	if _, err := p.parseScheduleTypesCollection.InsertMany(ctx, utils.ToInterfaceSlice(types)); err != nil {
		return err
	}

	return nil
}

func (p *repository) GetScheduleTypesToParse(ctx context.Context, parserAppName string) ([]entities.ScheduleTypeInfo, error) {
	result, err := p.parseScheduleTypesCollection.Find(ctx, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return nil, err
	}

	var types []entities.ScheduleTypeInfo
	if err = result.All(ctx, &types); err != nil {
		return nil, err
	}

	return types, nil
}

func (p *repository) UpdateParseJournalUser(ctx context.Context, user entities.JournalUser) error {
	_, err := p.parseJournalUserCollection.UpdateByID(ctx, user.ID, bson.M{"$set": user})
	return err
}

func (p *repository) UpdateGeneralSchedule(ctx context.Context, lessons []entities.GeneralLesson) error {
	if len(lessons) == 0 {
		return errors.New("empty lessons array")
	}

	_, err := p.generalLessonsCollection.DeleteMany(ctx, bson.D{{"studyPlaceId", lessons[0].StudyPlaceId}})
	if err != nil {
		return err
	}

	_, err = p.generalLessonsCollection.InsertMany(ctx, utils.ToInterfaceSlice(lessons))
	return err
}

func (p *repository) GetLessonIDByDateNameAndGroup(ctx context.Context, date time.Time, name string, group string) (primitive.ObjectID, error) {
	var lesson entities.Lesson

	result := p.lessonsCollection.FindOne(ctx, bson.M{"subject": name, "group": group, "startDate": bson.M{"$gte": date, "$lt": date.AddDate(0, 0, 1)}})
	if err := result.Decode(lesson); err != nil {
		return primitive.NilObjectID, err
	}

	return lesson.Id, nil
}

func (p *repository) GetLastLesson(ctx context.Context, studyPlaceId int) (entities.Lesson, error) {
	opt := options.FindOne()
	opt.Sort = bson.M{"startDate": -1}

	var lesson entities.Lesson
	err := p.lessonsCollection.FindOne(ctx, bson.M{"studyPlaceId": studyPlaceId}, opt).Decode(lesson)
	return lesson, err
}

func (p *repository) AddLessons(ctx context.Context, lessons []entities.Lesson) error {
	_, err := p.lessonsCollection.InsertMany(ctx, utils.ToInterfaceSlice(lessons))
	return err
}

func (p *repository) AddMarks(ctx context.Context, marks []entities.Mark) error {
	_, err := p.marksCollection.InsertMany(ctx, utils.ToInterfaceSlice(marks))
	return err
}