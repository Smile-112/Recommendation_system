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

// ВНИМАНИЕ: это “минимальная” версия планировщика.
// Она корректна как старт: ставит задания подряд (по дедлайну),
// учитывая занятость оператора через user_task, и “needoperator” для печати.
// Дальше ты расширишь: подбор устройств, тарифы энергии, совместимости и т.д.
func (p *Planner) Recompute(ctx context.Context, workspaceID int64) (RecomputeResult, error) {
	tasks, err := p.repos.ListTasksForPlanning(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}
	busy, err := p.repos.ListOperatorBusy(ctx, workspaceID)
	if err != nil {
		return RecomputeResult{}, err
	}

	// сгруппируем занятость по оператору
	busyMap := map[int64][][2]time.Time{}
	for _, b := range busy {
		busyMap[b.OperatorID] = append(busyMap[b.OperatorID], [2]time.Time{b.Start, b.End})
	}

	// сортировка по дедлайну, затем по приоритету
	sort.Slice(tasks, func(i, j int) bool {
		di := farFutureIfNil(tasks[i].Deadline)
		dj := farFutureIfNil(tasks[j].Deadline)
		if !di.Equal(dj) {
			return di.Before(dj)
		}
		return tasks[i].PriorityID < tasks[j].PriorityID
	})

	now := time.Now()
	updated := 0
	var unscheduled []int64

	for _, t := range tasks {
		// общий блок = setup + print + unload
		total := t.SetupTime + t.Duration + t.UnloadTime

		// В этой простой версии:
		// - мы используем назначенного оператора и устройство из device_task
		// - ищем ближайший интервал, где оператор не занят:
		start := findNextFreeStart(now, total, busyMap[t.OperatorID])
		end := start.Add(total)

		// дедлайн: если хотите запрещать просрочку — тут можно "continue"
		// (пока просто пометим как unscheduled, если дедлайн есть и нарушен)
		if t.Deadline != nil && end.After(*t.Deadline) {
			unscheduled = append(unscheduled, t.ID)
			continue
		}

		if err := p.repos.UpdateDeviceTaskPlan(ctx, t.ID, start, end); err != nil {
			return RecomputeResult{}, err
		}
		updated++
		now = end // “заполняем” последовательно (MVP)
	}

	return RecomputeResult{Updated: updated, UnscheduledIDs: unscheduled}, nil
}

func farFutureIfNil(t *time.Time) time.Time {
	if t == nil {
		return time.Now().Add(365 * 24 * time.Hour)
	}
	return *t
}

// Находит ближайший старт, где [start, start+dur] не пересекается с занятостью оператора
func findNextFreeStart(start time.Time, dur time.Duration, busy [][2]time.Time) time.Time {
	cur := start

	for {
		conflict := false
		end := cur.Add(dur)

		for _, iv := range busy {
			a, b := iv[0], iv[1]
			if intersects(cur, end, a, b) {
				// сдвигаем старт на конец занятого интервала
				cur = b
				conflict = true
				break
			}
		}
		if !conflict {
			return cur
		}
	}
}

func intersects(a1, a2, b1, b2 time.Time) bool {
	return a1.Before(b2) && b1.Before(a2)
}
