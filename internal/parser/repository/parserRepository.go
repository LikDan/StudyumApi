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

type ParserRepository struct {
	generalLessonsCollection *mongo.Collection
	lessonsCollection        *mongo.Collection
	marksCollection          *mongo.Collection

	parseJournalUserCollection   *mongo.Collection
	parseScheduleTypesCollection *mongo.Collection
}

func NewParserRepository(client *mongo.Client) *ParserRepository {
	database := client.Database("Schedule")

	return &ParserRepository{
		generalLessonsCollection: database.Collection("GeneralLessons"),
		lessonsCollection:        database.Collection("Lessons"),
		marksCollection:          database.Collection("Marks"),

		parseJournalUserCollection:   database.Collection("ParseJournalUsers"),
		parseScheduleTypesCollection: database.Collection("ParseScheduleTypes"),
	}
}

func (p *ParserRepository) GetUsersToParse(ctx context.Context, parserAppName string, users *[]entities.JournalUser) error {
	result, err := p.parseJournalUserCollection.Find(ctx, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return err
	}

	if err = result.All(ctx, users); err != nil {
		return err
	}

	return nil
}

func (p *ParserRepository) InsertScheduleTypes(ctx context.Context, types []*entities.ScheduleTypeInfo) error {
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

func (p *ParserRepository) GetScheduleTypesToParse(ctx context.Context, parserAppName string, types *[]entities.ScheduleTypeInfo) error {
	result, err := p.parseScheduleTypesCollection.Find(ctx, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return err
	}

	if err = result.All(ctx, types); err != nil {
		return err
	}

	return nil
}

func (p *ParserRepository) UpdateParseJournalUser(ctx context.Context, user *entities.JournalUser) error {
	if _, err := p.parseJournalUserCollection.UpdateByID(ctx, user.ID, bson.M{"$set": user}); err != nil {
		return err
	}

	return nil
}

func (p *ParserRepository) UpdateGeneralSchedule(ctx context.Context, lessons []*entities.GeneralLesson) error {
	if len(lessons) == 0 {
		return errors.New("empty lessons array")
	}

	_, err := p.generalLessonsCollection.DeleteMany(ctx, bson.D{{"studyPlaceId", lessons[0].StudyPlaceId}})
	if err != nil {
		return err
	}

	_, err = p.generalLessonsCollection.InsertMany(ctx, utils.ToInterfaceSlice(lessons))
	if err != nil {
		return err
	}

	return nil
}

func (p *ParserRepository) GetLessonByDate(ctx context.Context, date time.Time, name string, group string) (entities.Lesson, error) {
	var lesson entities.Lesson

	result := p.lessonsCollection.FindOne(ctx, bson.M{"subject": name, "group": group, "startDate": bson.M{"$gte": date, "$lt": date.AddDate(0, 0, 1)}})
	if err := result.Decode(lesson); err != nil {
		return entities.Lesson{}, err
	}

	return lesson, nil
}

func (p *ParserRepository) GetLastLesson(ctx context.Context, studyPlaceId int, lesson *entities.Lesson) error {
	opt := options.FindOne()
	opt.Sort = bson.M{"startDate": -1}

	if err := p.lessonsCollection.FindOne(ctx, bson.M{"studyPlaceId": studyPlaceId}, opt).Decode(lesson); err != nil {
		return err
	}

	return nil
}

func (p *ParserRepository) AddLessons(ctx context.Context, lessons []*entities.Lesson) error {
	if _, err := p.lessonsCollection.InsertMany(ctx, utils.ToInterfaceSlice(lessons)); err != nil {
		return err
	}

	return nil
}

func (p *ParserRepository) AddMarks(ctx context.Context, marks []*entities.Mark) error {
	if _, err := p.marksCollection.InsertMany(ctx, utils.ToInterfaceSlice(marks)); err != nil {
		return err
	}

	return nil
}
