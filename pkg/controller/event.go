package controller

type eventType string

const (
	addWaitingRoom             eventType = "addWaitingRoom"
)

type event struct {
	eventType      eventType
	oldObj, newObj interface{}
}