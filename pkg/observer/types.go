package observer

import (
	"fmt"
	"regexp"
)


type DDMObserver struct {
	DiskStat []*DiskStat
}

//go:generate stringer -type=ActionStatus
type ActionStatus int

const (
	None    ActionStatus = 0
	Running ActionStatus = 1 << iota
	Iddle
	Stopped
	OK
)

type Action struct {
	Name             string
	Cron             string
	Status           ActionStatus
	DisabledByAction bool
	ActionsToStop    []*Action
	HealthCheck      ActionStatus
}

func (act Action) IsInitState() bool {
	return act.Status != Iddle && act.Status != Running
}

func (act Action) IsStopped() bool {
	return act.Status == Stopped
}

func (act Action) IsIddle() bool {
	return act.Status == Iddle
}

func (act Action) IsRunning() bool {
	return act.Status == Running
}

func (act *Action) SetToStop() {
	act.Status = Stopped
}

func (act *Action) SetToRun() {
	act.Status = Running
}

func (act *Action) SetToIddle() {
	act.Status = Iddle
}

func (as Action) Print() string {
	return fmt.Sprintf("Status: %s, DisabledByAction: %t", as.Status.String(), as.DisabledByAction)
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
	return d.Mount.IsIddle() &&
		d.Test.IsIddle()
}

func (d DiskStat) IsActiv() bool {
	return d.Active
}

func (d DiskStat) IsInActive() bool {
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
}
