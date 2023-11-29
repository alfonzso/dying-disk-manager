package observer

import (
	"fmt"
	"regexp"
)

type DDMObserver struct {
	DiskStat []DiskStat
}

//go:generate stringer -type=ActionStatus
type ActionStatus int

const (
	Running ActionStatus = iota
	Iddle
)

type Action struct {
	Status           ActionStatus
	ThreadIsRunning  bool
	DisabledByAction bool
}

func (as ActionStatus) IsIddle() bool {
	return as == Iddle
}

func (as ActionStatus) IsRunning() bool {
	return as == Running
}

func (as Action) Print() string {
	return fmt.Sprintf("Status: %s, ThreadIsRunning: %t, DisabledByAction: %t", as.Status.String(), as.ThreadIsRunning, as.DisabledByAction)
}

type DiskStat struct {
	Name           string
	UUID           string
	Active         bool
	InactiveReason []string
	Repair         Action
	Mount          Action
	Test           Action
}

func (d DiskStat) IsMountAndTestActionInIddleStatus() bool {
	return d.Mount.Status.IsIddle() &&
		d.Test.Status.IsIddle()
}

func (d DiskStat) IsActiv() bool {
	return d.Active
}

func (d DiskStat) IsNotActiv() bool {
	return !d.Active
}

func (d DiskStat) String() string {
	msg := fmt.Sprintf(`
		|Name: %s
		|UUID: %s
		|Active: %t`,
		d.Name, d.UUID, d.Active,
	)
	regex, _ := regexp.Compile(`\t+[|]`)
	return regex.ReplaceAllString(msg, " ")

	// return msg
	// return strings.Replace(msg, "^\+s|", "", -1)
}
