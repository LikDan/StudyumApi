package log

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"
)

type GinWriter struct {
	w    io.Writer
	File *os.File
}

type RequestInfo struct {
	Time    time.Time `json:"time"`
	Code    string    `json:"code"`
	Timeout string    `json:"timeout"`
	Ip      string    `json:"ip"`
	Method  string    `json:"method"`
	Url     string    `json:"url"`
}

func (w GinWriter) Write(b []byte) (int, error) {
	line := string(b)
	if !strings.HasPrefix(line, "[GIN]") {
		return 0, nil
	}

	params := strings.Split(line, " | ")
	if len(params) != 5 {
		return 0, nil
	}

	params2 := strings.Split(params[4], " ")

	time_, _ := time.Parse("2006/01/02 - 15:04:05", strings.TrimPrefix(params[0], "[GIN] "))
	info := RequestInfo{
		Time:    time_,
		Code:    params[1],
		Timeout: strings.Trim(params[2], " "),
		Ip:      strings.Trim(params[3], " "),
		Method:  params2[0],
		Url:     params2[len(params2)-1][1 : len(params2[len(params2)-1])-2],
	}

	bytes, err := json.Marshal(info)
	if err != nil {
		return 0, err
	}

	return w.File.WriteString(string(bytes) + "\n")
}
