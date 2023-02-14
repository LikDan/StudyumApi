package controllers

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"strings"
	"studyum/internal/apps/entities"
	"studyum/internal/apps/repositories"
)

type Controller interface {
	EventWithContext(ctx context.Context, studyPlaceID primitive.ObjectID, name string, data ...any)
	AsyncEventWithContext(ctx context.Context, studyPlaceID primitive.ObjectID, name string, data ...any)
	Event(studyPlaceID primitive.ObjectID, name string, data ...any)
	AsyncEvent(studyPlaceID primitive.ObjectID, name string, data ...any)
}

type controller struct {
	apps repositories.Apps
	data repositories.Data
}

func NewController(apps repositories.Apps, data repositories.Data) Controller {
	return &controller{apps: apps, data: data}
}

func (c *controller) EventWithContext(ctx context.Context, studyPlaceID primitive.ObjectID, name string, data ...any) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Panic recieved: %s", recover())
		}
	}()

	app, err := c.apps.GetByStudyPlaceID(ctx, studyPlaceID)
	if err != nil {
		return
	}

	i, ok := c.findInterface(name)
	if !ok {
		logrus.Warningln("No interface contains function with name " + name)
		return
	}

	appType := reflect.TypeOf(app)
	if !appType.Implements(i) {
		logrus.Debugln("This app does not implements suitable interface")
		return
	}

	method, ok := appType.MethodByName(name)
	if !ok {
		logrus.Errorln("No method with name " + name)
		return
	}

	if method.Func.Type().NumIn()-3 != len(data) {
		logrus.Errorf("No enough params. Passed %d, required %d\n", len(data), method.Func.Type().NumIn()-3)
		return
	}
	values := c.toReflect(app, ctx, data)

	trackable := c.getTrackable(values[2:])

	var record map[string]any
	switch trackable.Type {
	case entities.Field:
		record, err = c.data.Get(ctx, trackable.Collection, trackable.Property, trackable.Value)
		if err != nil {
			logrus.Errorln("Error getting data object: " + err.Error())
			return
		}
	case entities.Array:
		record, err = c.data.GetNested(ctx, trackable.Collection, trackable.Nested, trackable.Property, trackable.Value)
		if err != nil {
			logrus.Errorln("Error getting data object: " + err.Error())
			return
		}
	default:
		logrus.Errorln("No such trackable type")
		return
	}

	dataField := reflect.ValueOf(entities.Data{})
	if dataMap, ok := record[trackable.DataProperty]; ok {
		if data, ok := dataMap.(bson.M); ok {
			dataField = reflect.ValueOf(entities.Data(data))
		}
	}

	values = append(values, values[len(values)-1])
	copy(values[2:], values[2:])
	values[2] = dataField

	for index := 0; index < method.Func.Type().NumIn(); index++ {
		if t := method.Func.Type().In(index); !values[index].Type().AssignableTo(t) {
			logrus.Errorf("Cannot call method with param. Passed %s, required %s\n", values[index].Type().String(), t.String())
			return
		}
	}

	result := method.Func.Call(values)
	if len(result) == 0 {
		return
	}

	var resultData entities.Data
	for _, value := range result {
		if value.Type().AssignableTo(reflect.TypeOf(entities.Data{})) {
			resultData = value.Interface().(entities.Data)
		}
	}

	if len(resultData) == 0 {
		return
	}

	switch trackable.Type {
	case entities.Field:
		if err = c.data.Insert(ctx, trackable.Collection, trackable.Property, trackable.Value, trackable.DataProperty, resultData); err != nil {
			logrus.Errorln("Error getting data object: " + err.Error())
			return
		}
	case entities.Array:
		if err = c.data.InsertNested(ctx, trackable.Collection, trackable.Nested, trackable.Property, trackable.Value, trackable.DataProperty, resultData); err != nil {
			logrus.Errorln("Error getting data object: " + err.Error())
			return
		}
	}
}

func (c *controller) AsyncEventWithContext(ctx context.Context, studyPlaceID primitive.ObjectID, name string, data ...any) {
	go c.EventWithContext(ctx, studyPlaceID, name, data...)
}

func (c *controller) Event(studyPlaceID primitive.ObjectID, name string, data ...any) {
	c.EventWithContext(context.Background(), studyPlaceID, name, data...)
}

func (c *controller) AsyncEvent(studyPlaceID primitive.ObjectID, name string, data ...any) {
	c.AsyncEventWithContext(context.Background(), studyPlaceID, name, data...)
}

func (c *controller) toReflect(app entities.App, ctx context.Context, data []any) []reflect.Value {
	values := make([]reflect.Value, len(data)+2)
	values[0] = reflect.ValueOf(app)
	values[1] = reflect.ValueOf(ctx)
	for i, el := range data {
		values[i+2] = reflect.ValueOf(el)
	}

	return values
}

func (c *controller) findInterface(f string) (reflect.Type, bool) {
	types := []reflect.Type{
		reflect.TypeOf((*entities.LessonsManageInterface)(nil)).Elem(),
		reflect.TypeOf((*entities.MarksManageInterface)(nil)).Elem(),
		reflect.TypeOf((*entities.AbsencesManageInterface)(nil)).Elem(),
	}

	for _, t := range types {
		_, ok := t.MethodByName(f)
		if ok {
			return t, true
		}
	}

	return nil, false
}

func (c *controller) parseTag(tag string) (entities.Trackable, error) {
	err := errors.New("Wrong struct tag: " + tag)

	i := strings.Index(tag, ",")
	if i == -1 || tag[:i] != "trackable" {
		return entities.Trackable{}, err
	}

	values := map[string]string{}
	for _, entry := range strings.Split(tag[i:], ",") {
		i := strings.Index(entry, "=")
		if i == -1 {
			continue
		}

		pair := strings.Split(entry, "=")
		values[pair[0]] = pair[1]
	}

	if values["collection"] == "" {
		return entities.Trackable{}, err
	}

	trackable := entities.DefaultTrackable(values["collection"])
	if values["type"] != "" {
		trackable.Type = entities.TrackableType(values["type"])
	}
	if values["property"] != "" {
		trackable.Property = values["property"]
	}
	if values["dataProperty"] != "" {
		trackable.DataProperty = values["dataProperty"]
	}
	if values["nested"] != "" {
		trackable.Nested = values["nested"]
	}

	return trackable, nil
}

func (c *controller) findPropWithTag(value reflect.Value, tag string) (entities.Trackable, bool) {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)

			tag, _ := field.Tag.Lookup("apps")
			trackable, ok := c.findPropWithTag(value.Field(i), tag)
			if ok {
				trackable.Field = value.Field(i)
				return trackable, true
			}
		}
	case reflect.Array: //primitive.ObjectID -> byte array
		if !value.Type().AssignableTo(reflect.TypeOf(primitive.NilObjectID)) || tag == "" {
			return entities.Trackable{}, false
		}

		trackable, err := c.parseTag(tag)
		if err != nil {
			logrus.Errorln("Error parsing tag " + err.Error())
			return entities.Trackable{}, false
		}

		return trackable, true
	}

	return entities.Trackable{}, false
}

func (c *controller) getTrackable(values []reflect.Value) entities.Trackable {
	var trackable entities.Trackable
	for _, value := range values {
		if t, ok := c.findPropWithTag(value, ""); ok {
			trackable = t
			break
		}
	}

	if trackable.Collection == "" {
		return trackable
	}

	trackable.Value = trackable.Field.Interface().(primitive.ObjectID)
	return trackable
}
