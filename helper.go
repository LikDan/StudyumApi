package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"log"
	"reflect"
	"time"
)

func checkError(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}

func EqualStateInfo(a, b []StateInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func message(ctx *gin.Context, name string, value string, code int) {
	ctx.JSON(code, gin.H{name: value})
}

func errorMessage(ctx *gin.Context, value string) {
	ctx.Header("error", value)
	ctx.Header("cookie", "")
	ctx.JSON(204, gin.H{})
}

func isCronRunning(c *cron.Cron) bool {
	return reflect.ValueOf(c).Elem().FieldByName("running").Bool()
}

func sliceContains[T any](slice []T, element T) bool {
	for _, t := range slice {
		if reflect.DeepEqual(element, t) {
			return true
		}
	}
	return false
}

func Date() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func ToInterfaceSlice[T any](slice []T) []interface{} {
	var interface_ []interface{}
	for _, element := range slice {
		interface_ = append(interface_, element)
	}

	return interface_
}
