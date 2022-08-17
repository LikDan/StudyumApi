package firebase

type IFirebase interface {
	SendNotification(topic string, title string, body string, url string) (string, error)
}
