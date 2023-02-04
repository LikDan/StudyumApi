package mail

type Data map[string]string

type Mode string

const (
	DebugMode   Mode = "debug"
	ReleaseMode Mode = "release"
)
