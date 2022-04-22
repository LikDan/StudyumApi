package api

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
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

func ErrorMessage(ctx *gin.Context, value string) {
	ctx.JSON(418, value)
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

func DateEqual(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
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

func CheckNotEmpty(strings ...string) bool {
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

func MapSlice[T any, V any](s []*V, f func(int, *V) T) []T {
	var t []T

	for i, v := range s {
		t = append(t, f(i, v))
	}

	return t
}

func CheckAndMessage(ctx *gin.Context, code int, err error, logType LogType) bool {
	if CheckError(err, logType) {
		Message(ctx, code, err.Error())
		return true
	}
	return false
}

func Message(ctx *gin.Context, code int, msg interface{}) {
	ctx.JSON(code, msg)
}

func GetTimeDuration(hour, minute int) time.Duration {
	return time.Duration(hour*60*60+minute*60) * time.Second
}

type Shift struct {
	Start time.Duration
	End   time.Duration
}

func BindShift(sHour, sMinute, eHour, eMinute int) Shift {
	return Shift{
		Start: GetTimeDuration(sHour, sMinute),
		End:   GetTimeDuration(eHour, eMinute),
	}
}

func GenerateSecureToken() string {
	b := make([]byte, 128)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
