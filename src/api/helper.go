package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"reflect"
	"time"
)

type LogType int

const (
	INFO LogType = iota
	WARNING
	ERROR
	UNDEFINED
)

func CheckError(err error, type_ LogType) bool {
	if err != nil {
		if type_ == UNDEFINED {
			return true
		}

		switch type_ {
		case INFO:
			logrus.Info(err.Error())
			break
		case WARNING:
			logrus.Warning(err.Error())
			break
		case ERROR:
			logrus.Error(err.Error())
			break
		}
		return true
	}
	return false
}

func EqualStateInfo[T any](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if !reflect.DeepEqual(v, b[i]) {
			return false
		}
	}
	return true
}

func Message(ctx *gin.Context, name string, value string, code int) {
	ctx.JSON(code, gin.H{name: value})
}

func ErrorMessage(ctx *gin.Context, value string) {
	ctx.Header("error", value)
	ctx.Header("cookie", "")
	ctx.JSON(204, gin.H{})
}

func IsCronRunning(c *cron.Cron) bool {
	return reflect.ValueOf(c).Elem().FieldByName("running").Bool()
}

func SliceContains[T any](slice []T, element T) bool {
	for _, t := range slice {
		if reflect.DeepEqual(element, t) {
			return true
		}
	}
	return false
}

func Date() time.Time {
	return ToDateWithoutTime(time.Now())
}

func ToDateWithoutTime(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func ToInterfaceSlice[T any](slice []T) []interface{} {
	var interface_ []interface{}
	for _, element := range slice {
		interface_ = append(interface_, element)
	}

	return interface_
}

func EqualDateWithoutTime(date time.Time) bson.M {
	return bson.M{"$gte": ToDateWithoutTime(date), "$lt": ToDateWithoutTime(date.AddDate(0, 0, 1))}
}

func GetObjectId(ctx *gin.Context, key string) *primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(ctx.Query(key))
	if err != nil {
		return nil
	}

	return &id
}

func checkNotEmpty(strings ...string) bool {
	for _, s := range strings {
		if s == "" {
			return false
		}
	}
	return true
}

func DecodeJsonLines[T any](r io.Reader, results *[]T) {
	d := json.NewDecoder(r)

	for {
		var l T
		err := d.Decode(&l)
		if err == io.EOF {
			break
		}
		if err != nil {
			print(err.Error())
			continue
		}

		*results = append(*results, l)
	}
}
