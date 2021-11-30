package core

import (
	"errors"
	"sort"
	"testing"
	"time"
)

type SetEvent struct {
	flag bool
}

func (event *SetEvent) Trigger() {
	event.flag = true
}

func TestNegSimDur(t *testing.T) {
	_, err := NewScheduler(-1 * time.Second)
	if !errors.Is(err, NegSimDurErr) {
		t.Error("Cannot run simulations for a negative duration!")
	}
}

func TestEndTime(t *testing.T) {
	sched, err := NewScheduler(time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	event := &SetEvent{
		flag: false,
	}
	sched.Schedule(2*time.Second, event)
	sched.Run()
	if event.flag {
		t.Error("Triggered an event scheduled after the end time")
	}
	if !sched.IsStopped() {
		t.Error("Scheduler incorrectly reported as running")
	}
}

type ChronoEvent struct {
	order *[]int
	seqno int
}

func (event *ChronoEvent) Trigger() {
	*event.order = append(*event.order, event.seqno)
}

func TestChronologicalSchedule(t *testing.T) {
	order := []int{}
	sched, _ := NewScheduler(time.Minute)
	for _, seqno := range []int{7, 22, 11} {
		sched.Schedule(time.Duration(seqno)*time.Second, &ChronoEvent{
			order: &order,
			seqno: seqno,
		})
	}
	sched.Run()
	if len(order) != 3 || !sort.IntsAreSorted(order) {
		t.Error("Events were scheduled in the incorrect order")
	}
	if !sched.IsStopped() {
		t.Error("Scheduler incorrectly reported as running")
	}
}

type Generator struct {
	sched *Scheduler
}

func (event *Generator) Trigger() {
	event.sched.Schedule(time.Second, &Generator{
		sched: event.sched,
	})
}

func TestGenerativeEvents(t *testing.T) {
	sched, _ := NewScheduler(3 * time.Second)
	sched.Schedule(time.Duration(0), &Generator{
		sched: sched,
	})
	sched.Run()
	if sched.NumTriggered != 3 {
		t.Error("Generated the event an incorrect number of times!")
	}
}
