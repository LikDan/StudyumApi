package main

import (
	"context"
	"github.com/go-faker/faker/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"log"
	"math/rand"
	"os"
	"studyum/internal/auth/entities"
	entities2 "studyum/internal/general/entities"
	entities3 "studyum/internal/journal/entities"
	entities4 "studyum/internal/schedule/entities"
	"studyum/pkg/encryption"
	"studyum/pkg/hash"
	"time"
)

type ValueWithID struct {
	value string
	id    primitive.ObjectID
}

var (
	random *rand.Rand
	enc    = encryption.NewEncryption(os.Getenv("ENCRYPTION_SECRET"))

	primaryColors   = []string{"#e1e1e1", "#228EFF", "#22FF87"}
	secondaryColors = []string{"#e1e1e1", "#228EFF", "#22FF87"}

	rooms    = []ValueWithID{{"101", primitive.NewObjectID()}, {"102", primitive.NewObjectID()}, {"103", primitive.NewObjectID()}, {"104", primitive.NewObjectID()}, {"201", primitive.NewObjectID()}, {"202", primitive.NewObjectID()}, {"203", primitive.NewObjectID()}, {"204", primitive.NewObjectID()}, {"01a", primitive.NewObjectID()}, {"02a", primitive.NewObjectID()}}
	groups   = []ValueWithID{{"A-01", primitive.NewObjectID()}, {"A-02", primitive.NewObjectID()}, {"A-03", primitive.NewObjectID()}, {"B-01", primitive.NewObjectID()}, {"B-02", primitive.NewObjectID()}, {"C-01", primitive.NewObjectID()}}
	subjects = []ValueWithID{{"Algebra", primitive.NewObjectID()}, {"Computer science", primitive.NewObjectID()}, {"Physics", primitive.NewObjectID()}, {"Geometry", primitive.NewObjectID()}, {"PE", primitive.NewObjectID()}, {"AI", primitive.NewObjectID()}, {"Biology", primitive.NewObjectID()}, {"Geography", primitive.NewObjectID()}, {"Chemistry", primitive.NewObjectID()}, {"Environment science", primitive.NewObjectID()}}

	teachers   []string
	teacherIDs []primitive.ObjectID
	students   = make(map[string][]primitive.ObjectID)

	fromTime              = time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	toTime                = time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
	availableTimesToStart = []time.Duration{time.Hour * 8, time.Hour*9 + time.Minute*55, time.Hour*11 + time.Minute*20, time.Hour * 13, time.Hour*14 + time.Minute*10, time.Hour * 16, time.Hour*17 + time.Minute*40, time.Hour * 19}
	availableDurations    = []time.Duration{time.Minute * 45, time.Minute * 60, time.Minute * 90}

	defaultMarks = []entities2.MarkType{
		{
			Mark:        "A",
			WorkOutTime: 100,
		},
		{
			Mark:        "B",
			WorkOutTime: 100,
		},
		{
			Mark:        "C",
			WorkOutTime: 100,
		},
		{
			Mark:        "D",
			WorkOutTime: 100,
		},
		{
			Mark:        "F",
			WorkOutTime: 100,
		},
	}
)

func main() {
	random = initRandom(len(os.Args) == 0)

	if err := clearDB(); err != nil {
		log.Fatalf("Error clearing database: %v", err)
		return
	}

	usersAmount := 100 + random.Intn(300) //min 100 max 400
	codeUsersAmount := random.Intn(1000)
	lessonsAmount := random.Intn(10000)
	generalLessonsAmount := random.Intn(1000)

	studyPlaceID, err := saveStudyPlace()
	if err != nil {
		log.Fatalf("Error saving studyPlace to database: %v", err)
		return
	}

	saveStatics(studyPlaceID)

	if err := saveUsers(usersAmount, studyPlaceID, "password"); err != nil {
		log.Fatalf("Error saving users to database: %v", err)
		return
	}

	if err := saveCodeUsers(codeUsersAmount, studyPlaceID, "password"); err != nil {
		log.Fatalf("Error saving users to database: %v", err)
		return
	}

	if err := saveLessons(lessonsAmount, studyPlaceID); err != nil {
		log.Fatalf("Error saving lessons to database: %v", err)
		return
	}

	if err := saveGeneralLessons(generalLessonsAmount, studyPlaceID); err != nil {
		log.Fatalf("Error saving general lessons to database: %v", err)
		return
	}
}

func clearDB() error {
	ctx := context.Background()
	return client.Database("Studyum").Drop(ctx)
}

func saveStudyPlace() (primitive.ObjectID, error) {
	studyPlace := entities2.StudyPlace{
		Id:                primitive.NewObjectID(),
		WeeksCount:        1,
		Name:              "Test study place",
		PrimaryColorSet:   primaryColors,
		SecondaryColorSet: secondaryColors,
		JournalColors: entities2.JournalColors{
			General: "#eeeeee",
			Warning: "#ff1010",
			Danger:  "#ff0a0a",
		},
		LessonTypes: []entities2.LessonType{
			{
				Type:               "Lection",
				AbsenceWorkOutTime: 100,
				Marks:              defaultMarks,
				AssignedColor:      "",
				StandaloneMarks:    nil,
			},
			{
				Type:               "Practice",
				AbsenceWorkOutTime: 100,
				Marks:              defaultMarks,
				AssignedColor:      "#ffffaa",
				StandaloneMarks: []entities2.MarkType{
					{
						Mark:        "P",
						WorkOutTime: 100,
					},
					{
						Mark:        "NP",
						WorkOutTime: 100,
					},
				},
			},
		},
		AbsenceMark: "-",
	}

	ctx := context.Background()
	_, err := client.Database("Studyum").Collection("StudyPlaces").InsertOne(ctx, studyPlace)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return studyPlace.Id, nil
}

func saveStatics(studyPlaceID primitive.ObjectID) {
	subjects_ := make([]interface{}, len(subjects))
	for i, subject := range subjects {
		dict := make(map[string]any)

		dict["_id"] = subject.id
		dict["subject"] = subject.value
		dict["studyPlaceID"] = studyPlaceID

		subjects_[i] = dict
	}

	groups_ := make([]interface{}, len(groups))
	for i, group := range groups {
		dict := make(map[string]any)

		dict["_id"] = group.id
		dict["group"] = group.value
		dict["studyPlaceID"] = studyPlaceID

		groups_[i] = dict
	}

	rooms_ := make([]interface{}, len(rooms))
	for i, room := range rooms {
		dict := make(map[string]any)

		dict["_id"] = room.id
		dict["room"] = room.value
		dict["studyPlaceID"] = studyPlaceID

		rooms_[i] = dict
	}

	_, _ = client.Database("Studyum").Collection("Subjects").InsertMany(context.Background(), subjects_)
	_, _ = client.Database("Studyum").Collection("Groups").InsertMany(context.Background(), groups_)
	_, _ = client.Database("Studyum").Collection("Rooms").InsertMany(context.Background(), rooms_)
}

func initRandom(isRandom bool) (random *rand.Rand) {
	var seed int64
	if isRandom {
		seed = time.Now().Unix()
	} else {
		seed = 0
	}

	return rand.New(rand.NewSource(seed))
}

func createStudyPlaceInfo(studyPlaceID primitive.ObjectID) entities.UserStudyPlaceInfo {
	isTeacher := random.Intn(20)%3 == 0
	isTutor := isTeacher && random.Intn(7)%3 != 0
	name := faker.Name()
	var role string
	var roleName string
	var tuitionGroup string
	var permissions []string

	if isTeacher {
		role = "teacher"
		roleName = name
		if isTutor {
			tuitionGroup = groups[random.Intn(len(groups))].value
		}
		permissions = []string{"editJournal"} //viewJournals
		teachers = append(teachers, name)
	} else {
		role = "student"
		roleName = groups[random.Intn(len(groups))].value
	}

	return entities.UserStudyPlaceInfo{
		ID:           studyPlaceID,
		Name:         name,
		Role:         role,
		RoleName:     roleName,
		TuitionGroup: tuitionGroup,
		Permissions:  permissions,
		Accepted:     true,
	}
}

func createUser(password string, studyPlaceID primitive.ObjectID) entities.User {
	var passwordHash, _ = hash.Hash(password)

	id := primitive.NewObjectID()

	info := createStudyPlaceInfo(studyPlaceID)
	if info.Role == "student" {
		students[info.RoleName] = append(students[info.RoleName], id)
	} else {
		teacherIDs = append(teacherIDs, id)
	}

	return entities.User{
		Id:             id,
		Password:       passwordHash,
		Email:          faker.Email(),
		VerifiedEmail:  random.Intn(10)%7 != 0,
		FirebaseToken:  "",
		Login:          faker.Word(),
		PictureUrl:     faker.URL(),
		StudyPlaceInfo: &info,
	}
}

func saveUsers(amount int, studyPlaceID primitive.ObjectID, password string) error {
	users := make([]interface{}, amount)

	for i := 0; i < amount; i++ {
		user := createUser(password, studyPlaceID)
		enc.Encrypt(&user)
		users[i] = user
	}

	ctx := context.Background()
	_, err := client.Database("Studyum").Collection("Users").InsertMany(ctx, users)
	return err
}

func createCodeUser(password string, studyPlaceID primitive.ObjectID) entities.UserCodeData {
	id := primitive.NewObjectID()

	info := createStudyPlaceInfo(studyPlaceID)
	if info.Role == "student" {
		students[info.RoleName] = append(students[info.RoleName], id)
	} else {
		teacherIDs = append(teacherIDs, id)
	}

	var passwordHash, _ = hash.Hash(password)

	return entities.UserCodeData{
		Id:              id,
		Code:            faker.Word(),
		Name:            faker.Name(),
		StudyPlaceID:    info.ID,
		Role:            info.Role,
		RoleName:        info.RoleName,
		TuitionGroup:    info.TuitionGroup,
		Permissions:     info.Permissions,
		DefaultPassword: passwordHash,
	}
}

func saveCodeUsers(amount int, studyPlaceID primitive.ObjectID, password string) error {
	users := make([]interface{}, amount)

	for i := 0; i < amount; i++ {
		user := createCodeUser(password, studyPlaceID)
		enc.Encrypt(&user)
		users[i] = user
	}

	ctx := context.Background()
	_, err := client.Database("Studyum").Collection("CodeUsers").InsertMany(ctx, users)
	return err
}

func createLesson(studyPlaceID primitive.ObjectID) entities3.Lesson {
	lessonTypes := []string{"Practice", "Lection"}
	diff := toTime.Sub(fromTime).Hours() / 24
	addDay := random.Intn(int(diff))
	date := fromTime.Add(time.Duration(addDay*24) * time.Hour)
	index := random.Intn(len(availableTimesToStart))
	start := date.Add(availableTimesToStart[index])

	id := primitive.NewObjectID()

	marksAmount := random.Intn(50)
	marks := make([]entities3.Mark, marksAmount)
	group := groups[random.Intn(len(groups))]
	for i := 0; i < marksAmount; i++ {
		marks[i] = createMark(students[group.value][random.Intn(len(students[group.value]))], id, studyPlaceID)
	}

	return entities3.Lesson{
		Id:             id,
		StudyPlaceId:   studyPlaceID,
		PrimaryColor:   primaryColors[random.Intn(len(primaryColors))],
		SecondaryColor: primaryColors[random.Intn(len(secondaryColors))],
		Type:           lessonTypes[random.Intn(len(lessonTypes))],
		StartDate:      start,
		EndDate:        start.Add(availableDurations[random.Intn(len(availableDurations))]),
		LessonIndex:    index,
		SubjectID:      subjects[random.Intn(len(subjects))].id,
		GroupID:        group.id,
		RoomID:         rooms[random.Intn(len(rooms))].id,
		TeacherID:      teacherIDs[random.Intn(len(teachers))],
		Marks:          marks,
		//todo
		Absences: nil,
	}
}

func saveLessons(amount int, studyPlaceID primitive.ObjectID) error {
	lessons := make([]entities3.Lesson, amount)

	for i := 0; i < amount; i++ {
		lessons[i] = createLesson(studyPlaceID)
	}

	proceededLessons := removeOverflow(lessons)

	ctx := context.Background()
	_, err := client.Database("Studyum").Collection("Lessons").InsertMany(ctx, proceededLessons)
	return err
}

func saveGeneralLessons(amount int, studyPlaceID primitive.ObjectID) error {
	lessons := make([]entities3.Lesson, amount)

	for i := 0; i < amount; i++ {
		lessons[i] = createLesson(studyPlaceID)
	}

	proceededLessons := removeOverflow(lessons)

	proceededGeneralLessons := make([]interface{}, len(proceededLessons))
	for i, l := range proceededLessons {
		lesson := l.(entities3.Lesson)
		proceededGeneralLessons[i] = entities4.GeneralLesson{
			Id:             lesson.Id,
			StudyPlaceId:   lesson.StudyPlaceId,
			PrimaryColor:   lesson.PrimaryColor,
			SecondaryColor: lesson.SecondaryColor,
			StartTime:      lesson.StartDate.Format("15:04"),
			EndTime:        lesson.EndDate.Format("15:04"),
			SubjectID:      lesson.SubjectID,
			GroupID:        lesson.GroupID,
			TeacherID:      lesson.TeacherID,
			RoomID:         lesson.RoomID,
			Type:           lesson.Type,
			LessonIndex:    lesson.LessonIndex,
			DayIndex:       int(lesson.StartDate.Weekday()),
			WeekIndex:      0,
		}
	}

	ctx := context.Background()
	_, err := client.Database("Studyum").Collection("GeneralLessons").InsertMany(ctx, proceededGeneralLessons)
	return err
}

func createMark(studentID, lessonID, studyPlaceID primitive.ObjectID) entities3.Mark {
	return entities3.Mark{
		ID:           primitive.NewObjectID(),
		Mark:         defaultMarks[random.Intn(len(defaultMarks))].Mark,
		StudentID:    studentID,
		LessonID:     lessonID,
		StudyPlaceID: studyPlaceID,
	}
}

func removeOverflow(lessons []entities3.Lesson) []interface{} {
	var proceededLessons []interface{}

	groups := make(map[string][]entities3.Lesson)
	for _, lesson := range lessons {
		groups[lesson.Group] = append(groups[lesson.Group], lesson)
	}

	for _, arr := range groups {
		slices.SortFunc(arr, func(a, b entities3.Lesson) bool {
			return a.StartDate.Compare(b.StartDate) == -1
		})

		proceededLessons = append(proceededLessons, arr[0])
		previousLesson := arr[0]
		for _, lesson := range arr[1:] {
			if previousLesson.EndDate.After(lesson.StartDate) {
				continue
			}

			proceededLessons = append(proceededLessons, lesson)
			previousLesson = lesson
		}
	}

	return proceededLessons
}
