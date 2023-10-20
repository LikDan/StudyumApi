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
		Description:       "Test study place description. This is university in Country with some specializations. It focusing on educate dumb guys to make them smart :)",
		PictureUrl:        "data:image/png;base64,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",
		BannerUrl:         "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSHd7r0VtowAK_ZkE_LpBGgvYWB7uUmDd9OBzR0Wix16lQUZR9Fhsh7mI2dPAxZzosFMA&usqp=CAU",
		Address:           "Country, Region, City, District, Street, Building",
		Phone:             "+1234567890",
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
