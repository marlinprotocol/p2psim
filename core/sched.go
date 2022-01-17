package core

import (
	"container/heap"
	"errors"
	"time"
)

// The current simulation application is implemented as a single-threaded dispatch of events
// The scheduler schedules these events by running for a total of the specified duration
// Currently the scheduler only supports scheduling events after a specific time duration which is sufficent for
//   our current purposes
// The events are triggerred in the chronological order assuming that the events are being scheduled into the future
//
// Usecases
// - simulate network latency by scheduling a message receive after the latency duration has expired
// - simulate heartbeats by firing a beat event after the heartbeat interval
// - simulate block generation by scheduling generation events
//
// TODO: Support cancelling schedules

var (
	NegSimDurErr = errors.New("Simulation duration cannot be negative!")
)

type Scheduler struct {
	taskQ        TaskQueue
	endTime      time.Time
	CurTime      time.Time // useful for interval calculations (do not use for absolute time)
	NumTriggered int64     // doesn't include incomplete events
}

type TaskQueue []*Task

type Task struct {
	triggerTime time.Time
	event       Event
}

type Event interface {
	Trigger()
}

// dur represents the total duration for which simulation is run
func NewScheduler(dur time.Duration) (*Scheduler, error) {
	if dur < 0 {
		return nil, NegSimDurErr
	}
	epoch := time.Time{}
	endTime := epoch.Add(dur)
	sched := &Scheduler{
		taskQ:        TaskQueue{},
		CurTime:      epoch,
		endTime:      endTime,
		NumTriggered: 0,
	}
	return sched, nil
}

func (sched *Scheduler) Run() {
	// loop until no more tasks remain
	// new tasks are added via the schedule function
	for len(sched.taskQ) > 0 {
		task := heap.Pop(&sched.taskQ).(*Task)
		sched.CurTime = task.triggerTime
		if !sched.CurTime.Before(sched.endTime) {
			// end simulation because we are past our time
			// endTime is not inclusive
			break
		}
		task.event.Trigger()
		sched.NumTriggered++
	}
}

func (sched *Scheduler) Schedule(after time.Duration, event Event) {
	// schedule for later execution
	heap.Push(&sched.taskQ, &Task{
		triggerTime: sched.CurTime.Add(after),
		event:       event,
	})
}

func (sched *Scheduler) IsStopped() bool {
	return len(sched.taskQ) == 0 || !sched.CurTime.Before(sched.endTime)
}

/////////////
// Implement heap.Interface methods

func (taskQ TaskQueue) Len() int {
	return len(taskQ)
}

func (taskQ TaskQueue) Less(i, j int) bool {
	// task with lower time is popped first
	return taskQ[i].triggerTime.Before(taskQ[j].triggerTime)
}

func (taskQ TaskQueue) Swap(i, j int) {
	taskQ[i], taskQ[j] = taskQ[j], taskQ[i]
}

func (taskQP *TaskQueue) Push(x interface{}) {
	(*taskQP) = append((*taskQP), x.(*Task))
}

func (taskQP *TaskQueue) Pop() interface{} {
	taskQ := *taskQP
	numTasks := len(taskQ)
	task := taskQ[numTasks-1]    // take last element before popping
	*taskQP = taskQ[:numTasks-1] // pop the last element
	return task
}
