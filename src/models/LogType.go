package models

type LogType int

const (
	INFO LogType = iota
	WARNING
	ERROR
	UNDEFINED
)
