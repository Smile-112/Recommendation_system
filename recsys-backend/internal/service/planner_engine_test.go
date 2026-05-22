package service

import (
	"testing"
	"time"
)

func TestBuildPlanMovesFutureTaskToFitNewUrgentTask(t *testing.T) {
	now := mustTime("2026-05-22T09:30:00Z")
	deadlineNew := mustTime("2026-05-22T11:00:00Z")
	deadlineExisting := mustTime("2026-05-22T13:00:00Z")

	out := buildScoreBasedPlan(planningInput{
		Now:       now,
		NewTaskID: 2,
		Tasks: []planningTask{
			{
				ID:               1,
				Duration:         time.Hour,
				Deadline:         &deadlineExisting,
				CharacteristicID: 10,
				Priority:         1,
				PlanStart:        timePtr(mustTime("2026-05-22T10:00:00Z")),
				PlanEnd:          timePtr(mustTime("2026-05-22T11:00:00Z")),
			},
			{
				ID:               2,
				Duration:         time.Hour,
				Deadline:         &deadlineNew,
				CharacteristicID: 10,
				Priority:         10,
			},
		},
		Devices:      []planningDevice{{ID: 100, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true}},
		Operators:    []planningOperator{{ID: 200}},
		Competencies: []planningCompetency{{OperatorID: 200, DeviceTypeID: 50}},
		Busy: []planningBusyInterval{{
			DeviceID: 100,
			Start:    mustTime("2026-05-22T09:30:00Z"),
			End:      mustTime("2026-05-22T10:00:00Z"),
		}},
		Options: planningOptions{DayStartHour: 9, DayEndHour: 22},
	})

	newTask := assignmentByTask(t, out, 2)
	if !newTask.Start.Equal(mustTime("2026-05-22T10:00:00Z")) {
		t.Fatalf("new urgent task starts at %s, want 10:00", newTask.Start)
	}
	if newTask.WarningCode != "" {
		t.Fatalf("new urgent task warning = %q, want none", newTask.WarningCode)
	}

	moved := assignmentByTask(t, out, 1)
	if !moved.Moved {
		t.Fatalf("existing future task was not marked moved")
	}
	if !moved.Start.Equal(mustTime("2026-05-22T11:00:00Z")) {
		t.Fatalf("existing task moved to %s, want 11:00", moved.Start)
	}
}

func TestBuildPlanFallsBackWithDeadlineWarningWhenNoStrictPlanExists(t *testing.T) {
	now := mustTime("2026-05-22T09:30:00Z")
	deadlineNew := mustTime("2026-05-22T11:00:00Z")
	deadlineExisting := mustTime("2026-05-22T11:30:00Z")

	out := buildScoreBasedPlan(planningInput{
		Now:       now,
		NewTaskID: 2,
		Tasks: []planningTask{
			{
				ID:               1,
				Duration:         90 * time.Minute,
				Deadline:         &deadlineExisting,
				CharacteristicID: 10,
				Priority:         1,
				PlanStart:        timePtr(mustTime("2026-05-22T10:00:00Z")),
				PlanEnd:          timePtr(mustTime("2026-05-22T11:30:00Z")),
			},
			{
				ID:               2,
				Duration:         time.Hour,
				Deadline:         &deadlineNew,
				CharacteristicID: 10,
				Priority:         10,
			},
		},
		Devices:      []planningDevice{{ID: 100, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true}},
		Operators:    []planningOperator{{ID: 200}},
		Competencies: []planningCompetency{{OperatorID: 200, DeviceTypeID: 50}},
		Options:      planningOptions{DayStartHour: 9, DayEndHour: 22},
	})

	newTask := assignmentByTask(t, out, 2)
	if newTask.WarningCode != "deadline_missed" {
		t.Fatalf("warning = %q, want deadline_missed", newTask.WarningCode)
	}
	if !newTask.Start.Equal(mustTime("2026-05-22T11:30:00Z")) {
		t.Fatalf("fallback start = %s, want 11:30", newTask.Start)
	}

	existing := assignmentByTask(t, out, 1)
	if !existing.Start.Equal(mustTime("2026-05-22T10:00:00Z")) {
		t.Fatalf("existing task start = %s, want unchanged 10:00", existing.Start)
	}
}

func TestBuildPlanPrefersHigherDeviceCharacteristicScoreWhenDeadlineAllows(t *testing.T) {
	deadline := mustTime("2026-05-22T15:00:00Z")

	out := buildScoreBasedPlan(planningInput{
		Now:       mustTime("2026-05-22T09:00:00Z"),
		NewTaskID: 1,
		Tasks: []planningTask{{
			ID:               1,
			Duration:         time.Hour,
			Deadline:         &deadline,
			CharacteristicID: 10,
			Priority:         1,
		}},
		Devices: []planningDevice{
			{ID: 100, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true},
			{ID: 101, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true},
		},
		Operators:    []planningOperator{{ID: 200}},
		Competencies: []planningCompetency{{OperatorID: 200, DeviceTypeID: 50}},
		DeviceScores: []planningDeviceScore{
			{DeviceID: 100, CharacteristicID: 10, Score: 20},
			{DeviceID: 101, CharacteristicID: 10, Score: 90},
		},
		Options: planningOptions{DayStartHour: 9, DayEndHour: 22},
	})

	got := assignmentByTask(t, out, 1)
	if got.DeviceID != 101 {
		t.Fatalf("device = %d, want higher scoring device 101", got.DeviceID)
	}
}

func TestBuildPlanUsesEarlierDeviceWhenPreferredDeviceMissesDeadline(t *testing.T) {
	deadline := mustTime("2026-05-22T10:30:00Z")

	out := buildScoreBasedPlan(planningInput{
		Now:       mustTime("2026-05-22T09:00:00Z"),
		NewTaskID: 1,
		Tasks: []planningTask{{
			ID:               1,
			Duration:         time.Hour,
			Deadline:         &deadline,
			CharacteristicID: 10,
			Priority:         1,
		}},
		Devices: []planningDevice{
			{ID: 100, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true},
			{ID: 101, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true},
		},
		Operators:    []planningOperator{{ID: 200}},
		Competencies: []planningCompetency{{OperatorID: 200, DeviceTypeID: 50}},
		DeviceScores: []planningDeviceScore{
			{DeviceID: 100, CharacteristicID: 10, Score: 20},
			{DeviceID: 101, CharacteristicID: 10, Score: 90},
		},
		Busy: []planningBusyInterval{{
			DeviceID: 101,
			Start:    mustTime("2026-05-22T09:00:00Z"),
			End:      mustTime("2026-05-22T11:00:00Z"),
		}},
		Options: planningOptions{DayStartHour: 9, DayEndHour: 22},
	})

	got := assignmentByTask(t, out, 1)
	if got.DeviceID != 100 {
		t.Fatalf("device = %d, want earlier feasible device 100", got.DeviceID)
	}
	if got.WarningCode != "" {
		t.Fatalf("warning = %q, want none because lower score device meets deadline", got.WarningCode)
	}
}

func TestBuildPlanLeavesInProgressTaskFixed(t *testing.T) {
	now := mustTime("2026-05-22T09:30:00Z")
	deadlineNew := mustTime("2026-05-22T11:00:00Z")
	deadlineRunning := mustTime("2026-05-22T12:00:00Z")

	out := buildScoreBasedPlan(planningInput{
		Now:       now,
		NewTaskID: 2,
		Tasks: []planningTask{
			{
				ID:               1,
				Duration:         time.Hour,
				Deadline:         &deadlineRunning,
				CharacteristicID: 10,
				PlanStart:        timePtr(mustTime("2026-05-22T09:00:00Z")),
				PlanEnd:          timePtr(mustTime("2026-05-22T10:00:00Z")),
			},
			{
				ID:               2,
				Duration:         time.Hour,
				Deadline:         &deadlineNew,
				CharacteristicID: 10,
				Priority:         10,
			},
		},
		Devices:      []planningDevice{{ID: 100, DeviceTypeID: 50, CharacteristicID: 10, AddInRecSystem: true}},
		Operators:    []planningOperator{{ID: 200}},
		Competencies: []planningCompetency{{OperatorID: 200, DeviceTypeID: 50}},
		Options:      planningOptions{DayStartHour: 9, DayEndHour: 22},
	})

	running := assignmentByTask(t, out, 1)
	if running.Moved {
		t.Fatalf("in-progress task was marked moved")
	}
	if !running.Start.Equal(mustTime("2026-05-22T09:00:00Z")) || !running.End.Equal(mustTime("2026-05-22T10:00:00Z")) {
		t.Fatalf("in-progress task changed to %s-%s", running.Start, running.End)
	}

	newTask := assignmentByTask(t, out, 2)
	if !newTask.Start.Equal(mustTime("2026-05-22T10:00:00Z")) {
		t.Fatalf("new task start = %s, want after fixed interval at 10:00", newTask.Start)
	}
}

func assignmentByTask(t *testing.T, out planningOutput, taskID int64) planAssignment {
	t.Helper()
	for _, assignment := range out.Assignments {
		if assignment.TaskID == taskID {
			return assignment
		}
	}
	t.Fatalf("assignment for task %d not found in %#v", taskID, out.Assignments)
	return planAssignment{}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func timePtr(value time.Time) *time.Time {
	return &value
}
