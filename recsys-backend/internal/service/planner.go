package service

import (
	"context"
	"sort"
	"time"

	"recsys-backend/internal/storage"
)

type Planner struct {
	repos *storage.Repos
}

func NewPlanner(repos *storage.Repos) *Planner {
	return &Planner{repos: repos}
}

type RecomputeRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type RecomputeResult struct {
	Updated        int     `json:"updated"`
	UnscheduledIDs []int64 `json:"unscheduled_ids"`
}

const (
	workDayStartHour = 9
	workDayEndHour   = 22
	// maxScheduleAhead limits how far into the future the planner looks,
	// preventing infinite loops when there is no deadline.
	maxScheduleAhead = 365 * 24 * time.Hour
)

type interval struct {
	start time.Time
	end   time.Time
}

func (p *Planner) Recompute(ctx context.Context, workspaceID int64) (RecomputeResult, error) {
	tasks, err := p.repos.ListTasksForPlanning(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	allTasks, err := p.repos.ListDeviceTasksForWorkspace(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	busy, err := p.repos.ListOperatorBusy(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}

	plannedIDs := make(map[int64]struct{}, len(tasks))
	for _, t := range tasks {
		plannedIDs[t.ID] = struct{}{}
	}

	operatorBusy := map[int64][]interval{}
	for _, b := range busy {
		operatorBusy[b.OperatorID] = append(operatorBusy[b.OperatorID], interval{start: b.Start, end: b.End})
	}

	deviceBusy := map[int64][]interval{}
	for _, t := range allTasks {
		if t.PlanStart == nil || t.PlanEnd == nil {
			continue
		}
		if _, ok := plannedIDs[t.ID]; ok {
			continue
		}
		if t.DeviceID > 0 {
			deviceBusy[t.DeviceID] = append(deviceBusy[t.DeviceID], interval{start: *t.PlanStart, end: *t.PlanEnd})
		}
		if t.NeedOperator && t.OperatorID > 0 {
			operatorBusy[t.OperatorID] = append(
				operatorBusy[t.OperatorID],
				interval{start: *t.PlanStart, end: *t.PlanEnd},
			)
		}
	}

	// farFuture is computed once so the sort comparator is deterministic.
	farFuture := time.Now().Add(maxScheduleAhead)
	sort.Slice(tasks, func(i, j int) bool {
		di := coalesceDeadline(tasks[i].Deadline, farFuture)
		dj := coalesceDeadline(tasks[j].Deadline, farFuture)
		if !di.Equal(dj) {
			return di.Before(dj)
		}
		return tasks[i].PriorityID < tasks[j].PriorityID
	})

	startAnchor := time.Now()
	updated := 0
	var unscheduled []int64

	for _, t := range tasks {
		if t.DeviceID <= 0 || (t.NeedOperator && t.OperatorID <= 0) {
			unscheduled = append(unscheduled, t.ID)
			continue
		}

		total := t.SetupTime + t.Duration + t.UnloadTime

		start, end, ok := findNextAvailableSlot(
			startAnchor,
			total,
			deviceBusy[t.DeviceID],
			operatorBusy[t.OperatorID],
			t.Deadline,
		)
		if !ok {
			unscheduled = append(unscheduled, t.ID)
			continue
		}

		if err := p.repos.UpdateDeviceTaskPlan(ctx, t.ID, start, end); err != nil {
			return RecomputeResult{}, err
		}
		deviceBusy[t.DeviceID] = append(deviceBusy[t.DeviceID], interval{start: start, end: end})
		if t.NeedOperator {
			operatorBusy[t.OperatorID] = append(operatorBusy[t.OperatorID], interval{start: start, end: end})
		}
		updated++
	}

	return RecomputeResult{Updated: updated, UnscheduledIDs: unscheduled}, nil
}

func coalesceDeadline(t *time.Time, fallback time.Time) time.Time {
	if t == nil {
		return fallback
	}
	return *t
}

// findNextAvailableSlot finds the earliest window of length dur starting at or after
// start where neither deviceBusy nor operatorBusy is occupied, within work hours.
// Returns false when no such window exists before the effective deadline (or
// maxScheduleAhead if no deadline is set).
func findNextAvailableSlot(
	start time.Time,
	dur time.Duration,
	deviceBusy []interval,
	operatorBusy []interval,
	deadline *time.Time,
) (time.Time, time.Time, bool) {
	cur := alignToWorkday(start)

	maxDate := start.Add(maxScheduleAhead)
	if deadline != nil && deadline.Before(maxDate) {
		maxDate = *deadline
	}

	busy := append([]interval{}, deviceBusy...)
	busy = append(busy, operatorBusy...)
	sort.Slice(busy, func(i, j int) bool {
		return busy[i].start.Before(busy[j].start)
	})

	for {
		if cur.After(maxDate) {
			return time.Time{}, time.Time{}, false
		}
		cur = alignToWorkday(cur)
		dayEnd := time.Date(cur.Year(), cur.Month(), cur.Day(), workDayEndHour, 0, 0, 0, cur.Location())
		end := cur.Add(dur)
		if end.After(dayEnd) {
			cur = nextWorkdayStart(cur)
			continue
		}
		conflict := false
		for _, iv := range busy {
			if intersects(cur, end, iv.start, iv.end) {
				cur = iv.end
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		if end.After(maxDate) {
			return time.Time{}, time.Time{}, false
		}
		return cur, end, true
	}
}

func alignToWorkday(t time.Time) time.Time {
	dayStart := time.Date(t.Year(), t.Month(), t.Day(), workDayStartHour, 0, 0, 0, t.Location())
	dayEnd := time.Date(t.Year(), t.Month(), t.Day(), workDayEndHour, 0, 0, 0, t.Location())
	if t.Before(dayStart) {
		return dayStart
	}
	if !t.Before(dayEnd) {
		return nextWorkdayStart(t)
	}
	return t
}

func nextWorkdayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, workDayStartHour, 0, 0, 0, t.Location())
}

func intersects(a1, a2, b1, b2 time.Time) bool {
	return a1.Before(b2) && b1.Before(a2)
}
