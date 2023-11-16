package observer

import (
	"fmt"
	"regexp"
)

type DDMObserver struct {
	DiskStat []DiskStat
}

type ActionStatusType int

const (
	Running ActionStatusType = iota
	Iddle
)

func (as ActionStatusType) IsIddle() bool {
	return as == Iddle
}

func (as ActionStatusType) IsRunning() bool {
	return as == Running
}

type DiskStat struct {
	Name                  string
	UUID                  string
	Active                bool
	InactiveReason        []string
	RepairThreadIsRunning bool
	TestThreadIsRunning   bool
	MountThreadIsRunning  bool
	ActionStatus          struct {
		Repair ActionStatusType
		Mount  ActionStatusType
		Test   ActionStatusType
	}
}

func (d DiskStat) IsMountAndTestActionInIddleStatus() bool {
	return d.ActionStatus.Mount.IsIddle() &&
		d.ActionStatus.Test.IsIddle()
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
