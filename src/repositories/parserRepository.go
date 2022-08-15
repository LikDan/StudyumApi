package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"studyum/src/models"
	"studyum/src/utils"
	"time"
)

type ParserRepository struct {
	*Repository

	parseJournalUserCollection   *mongo.Collection
	parseScheduleTypesCollection *mongo.Collection
}

func NewParserRepository(repository *Repository) *ParserRepository {
	return &ParserRepository{
		Repository:                   repository,
		parseJournalUserCollection:   repository.database.Collection("ParseJournalUsers"),
		parseScheduleTypesCollection: repository.database.Collection("ParseScheduleTypes"),
	}
}

func (p *ParserRepository) GetUsersToParse(ctx context.Context, parserAppName string, users *[]models.ParseJournalUser) *models.Error {
	result, err := p.parseJournalUserCollection.Find(ctx, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	if err := result.All(ctx, users); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) InsertScheduleTypes(ctx context.Context, types []*models.ScheduleTypeInfo) *models.Error {
	if len(types) == 0 {
		return models.BindErrorStr("Provided empty array", 418, models.UNDEFINED)
	}

	if _, err := p.parseScheduleTypesCollection.DeleteMany(ctx, bson.M{"parserAppName": types[0].ParserAppName}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	for _, type_ := range types {
		type_.Id = primitive.NewObjectID()
	}

	if _, err := p.parseScheduleTypesCollection.InsertMany(ctx, utils.ToInterfaceSlice(types)); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) GetScheduleTypesToParse(ctx context.Context, parserAppName string, types *[]models.ScheduleTypeInfo) *models.Error {
	result, err := p.parseScheduleTypesCollection.Find(ctx, bson.M{"parserAppName": parserAppName})
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	if err := result.All(ctx, types); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) UpdateParseJournalUser(ctx context.Context, user *models.ParseJournalUser) *models.Error {
	if _, err := p.parseJournalUserCollection.UpdateByID(ctx, user.ID, bson.M{"$set": user}); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) UpdateGeneralSchedule(ctx context.Context, lessons []*models.GeneralLesson) *models.Error {
	if len(lessons) == 0 {
		return models.BindErrorStr("Provided empty array", 418, models.UNDEFINED)
	}

	_, err := p.generalLessonsCollection.DeleteMany(ctx, bson.D{{"studyPlaceId", lessons[0].StudyPlaceId}})
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	_, err = p.generalLessonsCollection.InsertMany(ctx, utils.ToInterfaceSlice(lessons))
	if err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) GetLessonByDate(ctx context.Context, date time.Time, name string, group string, lesson *models.Lesson) {
	result := p.lessonsCollection.FindOne(ctx, bson.M{"subject": name, "group": group, "startDate": bson.M{"$gte": date, "$lt": date.AddDate(0, 0, 1)}})
	_ = result.Decode(lesson)
}

func (p *ParserRepository) GetLastLesson(ctx context.Context, studyPlaceId int, lesson *models.Lesson) *models.Error {
	opt := options.FindOne()
	opt.Sort = bson.M{"startDate": -1}

	if err := p.lessonsCollection.FindOne(ctx, bson.M{"studyPlaceId": studyPlaceId}, opt).Decode(lesson); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) AddLessons(ctx context.Context, lessons []*models.Lesson) *models.Error {
	if _, err := p.lessonsCollection.InsertMany(ctx, utils.ToInterfaceSlice(lessons)); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}

func (p *ParserRepository) AddMarks(ctx context.Context, marks []*models.Mark) *models.Error {
	if _, err := p.marksCollection.InsertMany(ctx, utils.ToInterfaceSlice(marks)); err != nil {
		return models.BindError(err, 418, models.WARNING)
	}

	return models.EmptyError()
}
