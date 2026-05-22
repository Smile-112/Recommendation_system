package service

import (
	"sort"
	"time"
)

const deadlineMissedWarning = "deadline_missed"

type planningInput struct {
	Now             time.Time
	NewTaskID       int64
	Tasks           []planningTask
	Devices         []planningDevice
	Operators       []planningOperator
	Competencies    []planningCompetency
	OperatorDevices []planningOperatorDevice
	DeviceScores    []planningDeviceScore
	Weights         map[string]float64
	Busy            []planningBusyInterval
	Options         planningOptions
}

type planningOutput struct {
	Assignments []planAssignment
	Warnings    []planWarning
	Unscheduled []int64
}

type planningTask struct {
	ID               int64
	Name             string
	Deadline         *time.Time
	Duration         time.Duration
	SetupTime        time.Duration
	UnloadTime       time.Duration
	CharacteristicID int64
	Priority         int
	NeedOperator     bool
	DeviceID         int64
	OperatorID       int64
	PlanStart        *time.Time
	PlanEnd          *time.Time
	CompletionMark   string
}

type planningDevice struct {
	ID               int64
	DeviceTypeID     int64
	CharacteristicID int64
	AddInRecSystem   bool
}

type planningOperator struct {
	ID int64
}

type planningCompetency struct {
	OperatorID   int64
	DeviceTypeID int64
}

type planningOperatorDevice struct {
	OperatorID int64
	DeviceID   int64
}

type planningDeviceScore struct {
	DeviceID         int64
	CharacteristicID int64
	Score            float64
}

type planningBusyInterval struct {
	DeviceID   int64
	OperatorID int64
	Start      time.Time
	End        time.Time
}

type planningOptions struct {
	DayStartHour int
	DayEndHour   int
}

type planAssignment struct {
	TaskID      int64
	DeviceID    int64
	OperatorID  int64
	Start       time.Time
	End         time.Time
	Score       float64
	WarningCode string
	Explanation string
	Moved       bool
	Selected    bool
}

type planWarning struct {
	TaskID  int64
	Code    string
	Message string
}

type resourceCalendar struct {
	deviceBusy   map[int64][]interval
	operatorBusy map[int64][]interval
}

type candidateSlot struct {
	deviceID   int64
	operatorID int64
	start      time.Time
	end        time.Time
	score      float64
}

func buildScoreBasedPlan(input planningInput) planningOutput {
	input.Options = normalizePlanningOptions(input.Options)

	fixed, movable := splitPlanningTasks(input.Tasks, input.Now)
	baseCalendar := newResourceCalendar(input.Busy)
	assignments := make([]planAssignment, 0, len(input.Tasks))

	for _, task := range fixed {
		if task.PlanStart == nil || task.PlanEnd == nil {
			continue
		}
		assignment := planAssignment{
			TaskID:     task.ID,
			DeviceID:   preferredTaskDevice(input, task),
			OperatorID: preferredTaskOperator(input, task),
			Start:      *task.PlanStart,
			End:        *task.PlanEnd,
			Selected:   true,
		}
		assignments = append(assignments, assignment)
		baseCalendar.add(assignment.DeviceID, assignment.OperatorID, assignment.Start, assignment.End)
	}

	strictCalendar := baseCalendar.clone()
	strictAssignments, ok := scheduleMovableTasks(input, movable, strictCalendar, true)
	if ok {
		assignments = append(assignments, strictAssignments...)
		return planningOutput{Assignments: sortAssignments(assignments)}
	}

	fallbackCalendar := baseCalendar.clone()
	for _, task := range movable {
		if task.ID == input.NewTaskID || task.PlanStart == nil || task.PlanEnd == nil {
			continue
		}
		assignment := planAssignment{
			TaskID:     task.ID,
			DeviceID:   preferredTaskDevice(input, task),
			OperatorID: preferredTaskOperator(input, task),
			Start:      *task.PlanStart,
			End:        *task.PlanEnd,
			Selected:   true,
		}
		assignments = append(assignments, assignment)
		fallbackCalendar.add(assignment.DeviceID, assignment.OperatorID, assignment.Start, assignment.End)
	}

	var warnings []planWarning
	for _, task := range movable {
		if task.ID != input.NewTaskID {
			continue
		}
		slot, found := bestSlotForTask(input, task, fallbackCalendar, false)
		if !found {
			return planningOutput{
				Assignments: sortAssignments(assignments),
				Unscheduled: []int64{task.ID},
			}
		}
		assignment := assignmentFromSlot(task, slot, false)
		assignment.WarningCode = deadlineMissedWarning
		assignment.Explanation = "Closest realistic slot selected; deadline cannot be met without violating other task deadlines."
		assignments = append(assignments, assignment)
		warnings = append(warnings, planWarning{
			TaskID:  task.ID,
			Code:    deadlineMissedWarning,
			Message: "Cannot complete before deadline without violating other task deadlines.",
		})
	}

	return planningOutput{Assignments: sortAssignments(assignments), Warnings: warnings}
}

func scheduleMovableTasks(input planningInput, tasks []planningTask, calendar resourceCalendar, strict bool) ([]planAssignment, bool) {
	queue := append([]planningTask{}, tasks...)
	sort.SliceStable(queue, func(i, j int) bool {
		if queue[i].ID == input.NewTaskID && queue[j].ID != input.NewTaskID {
			if sameDeadline(queue[i].Deadline, queue[j].Deadline) {
				return true
			}
		}
		if queue[j].ID == input.NewTaskID && queue[i].ID != input.NewTaskID {
			if sameDeadline(queue[i].Deadline, queue[j].Deadline) {
				return false
			}
		}
		di := deadlineOrMax(queue[i].Deadline)
		dj := deadlineOrMax(queue[j].Deadline)
		if !di.Equal(dj) {
			return di.Before(dj)
		}
		if queue[i].Priority != queue[j].Priority {
			return queue[i].Priority > queue[j].Priority
		}
		return queue[i].ID < queue[j].ID
	})

	assignments := make([]planAssignment, 0, len(queue))
	for _, task := range queue {
		slot, found := bestSlotForTask(input, task, calendar, strict)
		if !found {
			return nil, false
		}
		assignment := assignmentFromSlot(task, slot, taskMoved(task, slot.start, slot.end))
		assignments = append(assignments, assignment)
		calendar.add(assignment.DeviceID, assignment.OperatorID, assignment.Start, assignment.End)
	}
	return assignments, true
}

func bestSlotForTask(input planningInput, task planningTask, calendar resourceCalendar, strict bool) (candidateSlot, bool) {
	var best candidateSlot
	found := false
	for _, device := range input.Devices {
		if !device.AddInRecSystem || !deviceMatchesTask(device, task) {
			continue
		}
		operatorIDs := []int64{0}
		if task.NeedOperator {
			operatorIDs = candidateOperators(input, device)
		}
		if len(operatorIDs) == 0 {
			continue
		}
		for _, operatorID := range operatorIDs {
			start, end, ok := earliestSlotForResource(input.Now, totalTaskDuration(task), calendar.deviceBusy[device.ID], calendar.operatorBusy[operatorID], input.Options)
			if !ok {
				continue
			}
			if strict && task.Deadline != nil && end.After(*task.Deadline) {
				continue
			}
			slot := candidateSlot{
				deviceID:   device.ID,
				operatorID: operatorID,
				start:      start,
				end:        end,
				score:      scoreCandidate(input, task, device, start, end),
			}
			if !found || betterSlot(slot, best, task, strict) {
				best = slot
				found = true
			}
		}
	}
	return best, found
}

func betterSlot(candidate, current candidateSlot, task planningTask, strict bool) bool {
	candidateMeetsDeadline := task.Deadline == nil || !candidate.end.After(*task.Deadline)
	currentMeetsDeadline := task.Deadline == nil || !current.end.After(*task.Deadline)
	if candidateMeetsDeadline != currentMeetsDeadline {
		return candidateMeetsDeadline
	}
	if strict && candidateMeetsDeadline && currentMeetsDeadline {
		if candidate.score != current.score {
			return candidate.score > current.score
		}
		return candidate.start.Before(current.start)
	}
	if !candidate.start.Equal(current.start) {
		return candidate.start.Before(current.start)
	}
	return candidate.score > current.score
}

func earliestSlotForResource(start time.Time, duration time.Duration, deviceBusy []interval, operatorBusy []interval, options planningOptions) (time.Time, time.Time, bool) {
	if duration <= 0 {
		duration = time.Minute
	}
	cur := alignToPlanningWorkday(start, options)
	busy := append([]interval{}, deviceBusy...)
	busy = append(busy, operatorBusy...)
	sort.Slice(busy, func(i, j int) bool {
		return busy[i].start.Before(busy[j].start)
	})

	limit := start.Add(maxScheduleAhead)
	for !cur.After(limit) {
		cur = alignToPlanningWorkday(cur, options)
		dayEnd := time.Date(cur.Year(), cur.Month(), cur.Day(), options.DayEndHour, 0, 0, 0, cur.Location())
		end := cur.Add(duration)
		if end.After(dayEnd) {
			cur = nextPlanningWorkdayStart(cur, options)
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
		return cur, end, true
	}
	return time.Time{}, time.Time{}, false
}

func splitPlanningTasks(tasks []planningTask, now time.Time) ([]planningTask, []planningTask) {
	var fixed []planningTask
	var movable []planningTask
	for _, task := range tasks {
		if task.CompletionMark == "true" || (task.PlanStart != nil && task.PlanEnd != nil && !task.PlanStart.After(now) && task.PlanEnd.After(now)) {
			fixed = append(fixed, task)
			continue
		}
		movable = append(movable, task)
	}
	return fixed, movable
}

func newResourceCalendar(busy []planningBusyInterval) resourceCalendar {
	calendar := resourceCalendar{
		deviceBusy:   map[int64][]interval{},
		operatorBusy: map[int64][]interval{},
	}
	for _, iv := range busy {
		calendar.add(iv.DeviceID, iv.OperatorID, iv.Start, iv.End)
	}
	return calendar
}

func (c resourceCalendar) clone() resourceCalendar {
	clone := resourceCalendar{
		deviceBusy:   map[int64][]interval{},
		operatorBusy: map[int64][]interval{},
	}
	for id, intervals := range c.deviceBusy {
		clone.deviceBusy[id] = append([]interval{}, intervals...)
	}
	for id, intervals := range c.operatorBusy {
		clone.operatorBusy[id] = append([]interval{}, intervals...)
	}
	return clone
}

func (c resourceCalendar) add(deviceID, operatorID int64, start, end time.Time) {
	iv := interval{start: start, end: end}
	if deviceID > 0 {
		c.deviceBusy[deviceID] = append(c.deviceBusy[deviceID], iv)
	}
	if operatorID > 0 {
		c.operatorBusy[operatorID] = append(c.operatorBusy[operatorID], iv)
	}
}

func assignmentFromSlot(task planningTask, slot candidateSlot, moved bool) planAssignment {
	return planAssignment{
		TaskID:      task.ID,
		DeviceID:    slot.deviceID,
		OperatorID:  slot.operatorID,
		Start:       slot.start,
		End:         slot.end,
		Score:       slot.score,
		Moved:       moved,
		Selected:    true,
		Explanation: "Selected by weighted resource score and feasible interval.",
	}
}

func candidateOperators(input planningInput, device planningDevice) []int64 {
	competent := map[int64]struct{}{}
	for _, competency := range input.Competencies {
		if competency.DeviceTypeID == device.DeviceTypeID {
			competent[competency.OperatorID] = struct{}{}
		}
	}
	var assigned []int64
	for _, link := range input.OperatorDevices {
		if link.DeviceID == device.ID {
			if _, ok := competent[link.OperatorID]; ok {
				assigned = append(assigned, link.OperatorID)
			}
		}
	}
	if len(assigned) > 0 {
		sort.Slice(assigned, func(i, j int) bool { return assigned[i] < assigned[j] })
		return assigned
	}
	var operators []int64
	for _, operator := range input.Operators {
		if _, ok := competent[operator.ID]; ok {
			operators = append(operators, operator.ID)
		}
	}
	sort.Slice(operators, func(i, j int) bool { return operators[i] < operators[j] })
	return operators
}

func scoreCandidate(input planningInput, task planningTask, device planningDevice, start, end time.Time) float64 {
	score := float64(task.Priority) * 10 * criterionWeight(input, "manual_priority", 1)
	score += deviceScore(input, device.ID, task.CharacteristicID) * criterionWeight(input, "device_efficiency", 1)
	if deviceMatchesTask(device, task) {
		score += 25 * criterionWeight(input, "characteristic_fit", 1)
	}
	if task.Deadline != nil {
		minutesUntilDeadline := task.Deadline.Sub(end).Minutes()
		if minutesUntilDeadline >= 0 {
			score += 1000 * criterionWeight(input, "deadline_urgency", 1)
			score += (minutesUntilDeadline / 60) * criterionWeight(input, "early_completion", 1)
		} else {
			score += minutesUntilDeadline * criterionWeight(input, "lateness_penalty", 1)
		}
	}
	score -= start.Sub(input.Now).Hours() * criterionWeight(input, "early_completion", 1)
	return score
}

func criterionWeight(input planningInput, code string, fallback float64) float64 {
	if input.Weights == nil {
		return fallback
	}
	weight, ok := input.Weights[code]
	if !ok {
		return fallback
	}
	return weight
}

func deviceScore(input planningInput, deviceID, characteristicID int64) float64 {
	for _, score := range input.DeviceScores {
		if score.DeviceID == deviceID && score.CharacteristicID == characteristicID {
			return score.Score
		}
	}
	return 0
}

func deviceMatchesTask(device planningDevice, task planningTask) bool {
	return task.CharacteristicID == 0 || device.CharacteristicID == 0 || device.CharacteristicID == task.CharacteristicID
}

func totalTaskDuration(task planningTask) time.Duration {
	return task.SetupTime + task.Duration + task.UnloadTime
}

func taskMoved(task planningTask, start, end time.Time) bool {
	if task.PlanStart == nil || task.PlanEnd == nil {
		return false
	}
	return !task.PlanStart.Equal(start) || !task.PlanEnd.Equal(end)
}

func sortAssignments(assignments []planAssignment) []planAssignment {
	sort.SliceStable(assignments, func(i, j int) bool {
		if !assignments[i].Start.Equal(assignments[j].Start) {
			return assignments[i].Start.Before(assignments[j].Start)
		}
		return assignments[i].TaskID < assignments[j].TaskID
	})
	return assignments
}

func normalizePlanningOptions(options planningOptions) planningOptions {
	if options.DayStartHour == 0 {
		options.DayStartHour = workDayStartHour
	}
	if options.DayEndHour == 0 {
		options.DayEndHour = workDayEndHour
	}
	return options
}

func alignToPlanningWorkday(t time.Time, options planningOptions) time.Time {
	dayStart := time.Date(t.Year(), t.Month(), t.Day(), options.DayStartHour, 0, 0, 0, t.Location())
	dayEnd := time.Date(t.Year(), t.Month(), t.Day(), options.DayEndHour, 0, 0, 0, t.Location())
	if t.Before(dayStart) {
		return dayStart
	}
	if !t.Before(dayEnd) {
		return nextPlanningWorkdayStart(t, options)
	}
	return t
}

func nextPlanningWorkdayStart(t time.Time, options planningOptions) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, options.DayStartHour, 0, 0, 0, t.Location())
}

func deadlineOrMax(deadline *time.Time) time.Time {
	if deadline == nil {
		return time.Unix(1<<62, 0)
	}
	return *deadline
}

func sameDeadline(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

func preferredTaskDevice(input planningInput, task planningTask) int64 {
	if task.DeviceID > 0 {
		return task.DeviceID
	}
	for _, device := range input.Devices {
		if device.AddInRecSystem && deviceMatchesTask(device, task) {
			return device.ID
		}
	}
	return 0
}

func preferredTaskOperator(input planningInput, task planningTask) int64 {
	if task.OperatorID > 0 {
		return task.OperatorID
	}
	for _, device := range input.Devices {
		if device.AddInRecSystem && deviceMatchesTask(device, task) {
			operators := candidateOperators(input, device)
			if len(operators) > 0 {
				return operators[0]
			}
		}
	}
	return 0
}
