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
	RunID           int64                     `json:"run_id"`
	Updated         int                       `json:"updated"`
	Warnings        []RecomputeWarning        `json:"warnings"`
	Recommendations []RecomputeRecommendation `json:"recommendations"`
	UnscheduledIDs  []int64                   `json:"unscheduled_ids"`
}

type RecomputeWarning struct {
	TaskID  int64  `json:"task_id"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RecomputeRecommendation struct {
	TaskID      int64     `json:"task_id"`
	DeviceID    int64     `json:"device_id"`
	OperatorID  int64     `json:"operator_id"`
	PlanStart   time.Time `json:"plan_start"`
	PlanEnd     time.Time `json:"plan_end"`
	Score       float64   `json:"score"`
	Selected    bool      `json:"selected"`
	WarningCode string    `json:"warning_code,omitempty"`
	Explanation string    `json:"explanation"`
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
	devices, err := p.repos.ListDevices(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	deviceTypes, err := p.repos.ListDeviceTypes(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	operators, err := p.repos.ListOperators(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	competencies, err := p.repos.ListOperatorCompetencies(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	operatorDevices, err := p.repos.ListOperatorDevices(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	userBusy, err := p.repos.ListOperatorBusy(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	deviceScores, err := p.repos.ListDeviceCharacteristicScores(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	weights, err := p.repos.ListPlanningWeights(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}

	input := planningInput{
		Now:             time.Now(),
		NewTaskID:       newestTaskID(tasks),
		Tasks:           mapPlanningTasks(tasks),
		Devices:         mapPlanningDevices(devices, deviceTypes),
		Operators:       mapPlanningOperators(operators),
		Competencies:    mapPlanningCompetencies(competencies),
		OperatorDevices: mapPlanningOperatorDevices(operatorDevices),
		DeviceScores:    mapPlanningDeviceScores(deviceScores),
		Weights:         mapPlanningWeights(weights),
		Busy:            mapPlanningBusy(allTasks, tasks, userBusy),
		Options:         planningOptions{DayStartHour: workDayStartHour, DayEndHour: workDayEndHour},
	}
	output := buildScoreBasedPlan(input)

	runID, err := p.repos.CreatePlanningRun(ctx, workspaceID, "completed")
	if err != nil {
		return RecomputeResult{}, err
	}

	updated := 0
	recommendations := make([]RecomputeRecommendation, 0, len(output.Assignments))
	for _, assignment := range output.Assignments {
		if assignment.DeviceID <= 0 || assignment.Start.IsZero() || assignment.End.IsZero() {
			continue
		}
		if err := p.repos.UpdateDeviceTaskPlan(ctx, assignment.TaskID, assignment.DeviceID, assignment.OperatorID, assignment.Start, assignment.End); err != nil {
			return RecomputeResult{}, err
		}
		updated++
		start := assignment.Start
		end := assignment.End
		warningText := warningTextForCode(assignment.WarningCode)
		if _, err := p.repos.CreatePlanningRecommendation(ctx, storage.PlanningRecommendation{
			RunID:       runID,
			TaskID:      assignment.TaskID,
			DeviceID:    assignment.DeviceID,
			OperatorID:  assignment.OperatorID,
			Start:       &start,
			End:         &end,
			Score:       assignment.Score,
			Selected:    true,
			WarningCode: assignment.WarningCode,
			WarningText: warningText,
			Explanation: assignment.Explanation,
		}); err != nil {
			return RecomputeResult{}, err
		}
		recommendations = append(recommendations, RecomputeRecommendation{
			TaskID:      assignment.TaskID,
			DeviceID:    assignment.DeviceID,
			OperatorID:  assignment.OperatorID,
			PlanStart:   assignment.Start,
			PlanEnd:     assignment.End,
			Score:       assignment.Score,
			Selected:    true,
			WarningCode: assignment.WarningCode,
			Explanation: assignment.Explanation,
		})
	}

	warnings := make([]RecomputeWarning, 0, len(output.Warnings))
	for _, warning := range output.Warnings {
		warnings = append(warnings, RecomputeWarning{TaskID: warning.TaskID, Code: warning.Code, Message: warning.Message})
	}

	unscheduled := output.Unscheduled
	if unscheduled == nil {
		unscheduled = []int64{}
	}
	return RecomputeResult{
		RunID:           runID,
		Updated:         updated,
		Warnings:        warnings,
		Recommendations: recommendations,
		UnscheduledIDs:  unscheduled,
	}, nil
}

func coalesceDeadline(t *time.Time, fallback time.Time) time.Time {
	if t == nil {
		return fallback
	}
	return *t
}

func newestTaskID(tasks []storage.DeviceTaskRow) int64 {
	var id int64
	for _, task := range tasks {
		if task.ID > id {
			id = task.ID
		}
	}
	return id
}

func mapPlanningTasks(tasks []storage.DeviceTaskRow) []planningTask {
	result := make([]planningTask, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, planningTask{
			ID:               task.ID,
			Name:             task.Name,
			Deadline:         task.Deadline,
			Duration:         task.Duration,
			SetupTime:        task.SetupTime,
			UnloadTime:       task.UnloadTime,
			CharacteristicID: task.EquipmentCharacteristicID,
			Priority:         int(task.PriorityID),
			NeedOperator:     task.NeedOperator,
			DeviceID:         task.DeviceID,
			OperatorID:       task.OperatorID,
			PlanStart:        task.PlanStart,
			PlanEnd:          task.PlanEnd,
			CompletionMark:   task.CompletionMark,
		})
	}
	return result
}

func mapPlanningDevices(devices []storage.Device, deviceTypes []storage.DeviceType) []planningDevice {
	characteristicsByType := map[int64]int64{}
	for _, deviceType := range deviceTypes {
		characteristicsByType[deviceType.ID] = deviceType.EquipmentCharacteristicID
	}
	result := make([]planningDevice, 0, len(devices))
	for _, device := range devices {
		addInRecSystem := true
		if device.AddInRecSystem != nil {
			addInRecSystem = *device.AddInRecSystem
		}
		result = append(result, planningDevice{
			ID:               device.ID,
			DeviceTypeID:     device.DeviceTypeID,
			CharacteristicID: characteristicsByType[device.DeviceTypeID],
			AddInRecSystem:   addInRecSystem,
		})
	}
	return result
}

func mapPlanningOperators(operators []storage.Operator) []planningOperator {
	result := make([]planningOperator, 0, len(operators))
	for _, operator := range operators {
		result = append(result, planningOperator{ID: operator.ID})
	}
	return result
}

func mapPlanningCompetencies(items []storage.OperatorCompetency) []planningCompetency {
	result := make([]planningCompetency, 0, len(items))
	for _, item := range items {
		result = append(result, planningCompetency{OperatorID: item.OperatorID, DeviceTypeID: item.DeviceTypeID})
	}
	return result
}

func mapPlanningOperatorDevices(items []storage.OperatorDevice) []planningOperatorDevice {
	result := make([]planningOperatorDevice, 0, len(items))
	for _, item := range items {
		result = append(result, planningOperatorDevice{OperatorID: item.OperatorID, DeviceID: item.DeviceID})
	}
	return result
}

func mapPlanningDeviceScores(items []storage.DeviceCharacteristicScore) []planningDeviceScore {
	result := make([]planningDeviceScore, 0, len(items))
	for _, item := range items {
		result = append(result, planningDeviceScore{
			DeviceID:         item.DeviceID,
			CharacteristicID: item.EquipmentCharacteristicID,
			Score:            item.Score,
		})
	}
	return result
}

func mapPlanningWeights(items []storage.PlanningWeight) map[string]float64 {
	result := map[string]float64{}
	for _, item := range items {
		result[item.CriterionCode] = item.Weight
	}
	return result
}

func mapPlanningBusy(allTasks []storage.DeviceTaskRow, planningTasks []storage.DeviceTaskRow, userBusy []storage.UserTaskBusy) []planningBusyInterval {
	planningIDs := map[int64]struct{}{}
	for _, task := range planningTasks {
		planningIDs[task.ID] = struct{}{}
	}
	result := make([]planningBusyInterval, 0, len(allTasks)+len(userBusy))
	for _, task := range allTasks {
		if _, ok := planningIDs[task.ID]; ok {
			continue
		}
		if task.PlanStart == nil || task.PlanEnd == nil {
			continue
		}
		result = append(result, planningBusyInterval{
			DeviceID:   task.DeviceID,
			OperatorID: task.OperatorID,
			Start:      *task.PlanStart,
			End:        *task.PlanEnd,
		})
	}
	for _, busy := range userBusy {
		result = append(result, planningBusyInterval{OperatorID: busy.OperatorID, Start: busy.Start, End: busy.End})
	}
	return result
}

func warningTextForCode(code string) string {
	if code == deadlineMissedWarning {
		return "Невозможно выполнить до дедлайна без нарушения сроков других задач."
	}
	return ""
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
