package core

import (
	"container/heap"
	"errors"
	"time"

	"github.com/emirpasic/gods/maps/hashmap"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
)

// Intro
// -----
//
// The current simulation application is implemented as a single-threaded dispatch of events.
// See `discrete event simulator` for more information.
// The app uses the below simulator to schedule these necessary events to be executed in the chronological order.
// The scheduler schedules these events by running for a total of the specified duration
//
// Usecases
// --------
//
// - simulate network latency by scheduling a message receive after the latency duration has expired
// - simulate heartbeats by firing a beat event after the heartbeat interval
// - simulate block generation by scheduling generation events
// - cancel timers (message received within time via an eager push)
//
// Implementation
// --------------
//
// Scheduler uses a red-black tree to store events using the event's schedule time as the sorting key
// This allows O(logn) event insert and popping the earliest event(s)
// To remove an event by id, the impl. uses a separate map data structure that maps the id to event time
//   removal by id is used for cancelling futures

var (
	NegSimDurErr = errors.New("Simulation duration cannot be negative!")
)

type Scheduler struct {
	// taskQ stores scheduled events and orders them in the chronologically
	// taskIDs maps event IDs to trigger times to allow easy cancellations
	// multiple events may share the same trigger time
	// => taskQ maps trigger times to slices of tasks
	taskQ   *rbt.Tree
	taskIDs *hashmap.Map

	// endTime is constant across the whole run and stops the simulator when CurTime crosses endTime
	// endTime is not inclusive
	// CurTime is set to the time at which the current event is being dispatched
	//   the event observes the current time as the dispatch time
	// CurTime should only be used for interval calculations (and not for its absolute value)
	endTime time.Time
	CurTime time.Time

	// The total number of events triggered. Doesn't include incomplete events
	NumTriggered int64
}

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
		taskQ:        rbt.NewWith(utils.TimeComparator),
		taskIDs:      hasmap.New(),
		endTime:      endTime,
		CurTime:      epoch,
		NumTriggered: 0,
	}
	return sched, nil
}

func (sched *Scheduler) Run() {
	// loop until no more tasks remain
	// new tasks are added via the schedule function
	for !sched.taskQ.Empty() {
		tasks := sched.taskQ.Left().([])
	}

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
