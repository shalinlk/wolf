package models
import "fmt"

type Error struct {
	Code int
	Message string
}

func (e Error)Error() string {
	return fmt.Sprintf("Error : %d : %s", e.Code, e.Message)
}

var ExpiredStartTimeError Error = Error{Code : 1007, Message : "Start time is already expired"}
var InvalidPublisherError Error = Error{Code : 1007, Message : "Invalid publisher"}