package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repos struct {
	DB *pgxpool.Pool
}

func NewRepos(db *pgxpool.Pool) *Repos {
	return &Repos{DB: db}
}

type DeviceTaskRow struct {
	ID               int64         `json:"id"`
	Name             string        `json:"name"`
	Deadline         *time.Time    `json:"deadline"`
	Duration         time.Duration `json:"duration"`    // печать
	SetupTime        time.Duration `json:"setup_time"`  // наладка
	UnloadTime       time.Duration `json:"unload_time"` // снятие изделия
	NeedOperator     bool          `json:"need_operator"`
	PlanStart        *time.Time    `json:"plan_start"`
	PlanEnd          *time.Time    `json:"plan_end"`
	DocNum           string        `json:"doc_num"`
	PriorityID       int64         `json:"priority_id"`
	OperatorID       int64         `json:"operator_id"`
	DeviceID         int64         `json:"device_id"`
	DeviceTaskTypeID int64         `json:"device_task_type_id"`
	WorkspaceID      int64         `json:"workspace_id"`
}

type UserTaskBusy struct {
	OperatorID int64     `json:"operator_id"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
}

// Health-check
func (r *Repos) Ping(ctx context.Context) error {
	return r.DB.Ping(ctx)
}

func (r *Repos) ListDeviceTasksForWorkspace(ctx context.Context, workspaceID int64) ([]DeviceTaskRow, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT
			dvctsk_id,
			dvctsk_name,
			dvctsk_deadline,
			dvctsk_duration,
			dvctsk_setuptime,
			dvctsk_timetocomplite,
			COALESCE(dvctsk_needoperator,false),
			dvctsk_planestarttime,
			dvctsk_planecomptime,
			dvctsk_docnum,
			priorities,
			operator,
			device,
			device_tasks_type,
			workspace
		FROM device_task
		WHERE workspace = $1
		ORDER BY dvctsk_id DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DeviceTaskRow
	for rows.Next() {
		var t DeviceTaskRow
		var duration pgtype.Time
		var setup pgtype.Time
		var unload pgtype.Time
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Deadline,
			&duration,
			&setup,
			&unload,
			&t.NeedOperator,
			&t.PlanStart,
			&t.PlanEnd,
			&t.DocNum,
			&t.PriorityID,
			&t.OperatorID,
			&t.DeviceID,
			&t.DeviceTaskTypeID,
			&t.WorkspaceID,
		); err != nil {
			return nil, err
		}
		t.Duration = timeToDuration(duration)
		t.SetupTime = timeToDuration(setup)
		t.UnloadTime = timeToDuration(unload)
		res = append(res, t)
	}
	return res, rows.Err()
}

func (r *Repos) ListTasksForPlanning(ctx context.Context, workspaceID int64) ([]DeviceTaskRow, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT
			dvctsk_id,
			dvctsk_name,
			dvctsk_deadline,
			dvctsk_duration,
			dvctsk_setuptime,
			dvctsk_timetocomplite,
			COALESCE(dvctsk_needoperator,false),
			dvctsk_planestarttime,
			dvctsk_planecomptime,
			dvctsk_docnum,
			priorities,
			operator,
			device,
			device_tasks_type,
			workspace
		FROM device_task
		WHERE workspace = $1
		  AND COALESCE(dvctsk_addinrecsystem,false) = true
		  AND (dvctsk_complitionmark IS NULL OR dvctsk_complitionmark = '' OR dvctsk_complitionmark = 'false')
		ORDER BY COALESCE(dvctsk_deadline, now() + interval '365 days') ASC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DeviceTaskRow
	for rows.Next() {
		var t DeviceTaskRow
		var duration pgtype.Time
		var setup pgtype.Time
		var unload pgtype.Time
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Deadline,
			&duration,
			&setup,
			&unload,
			&t.NeedOperator,
			&t.PlanStart,
			&t.PlanEnd,
			&t.DocNum,
			&t.PriorityID,
			&t.OperatorID,
			&t.DeviceID,
			&t.DeviceTaskTypeID,
			&t.WorkspaceID,
		); err != nil {
			return nil, err
		}
		t.Duration = timeToDuration(duration)
		t.SetupTime = timeToDuration(setup)
		t.UnloadTime = timeToDuration(unload)
		res = append(res, t)
	}
	return res, rows.Err()
}

func (r *Repos) ListOperatorBusy(ctx context.Context, workspaceID int64) ([]UserTaskBusy, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT operator, usertsk_starttime, usertsk_endtime
		FROM user_task
		WHERE workspace = $1
		  AND usertsk_starttime IS NOT NULL
		  AND usertsk_endtime IS NOT NULL
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []UserTaskBusy
	for rows.Next() {
		var b UserTaskBusy
		if err := rows.Scan(&b.OperatorID, &b.Start, &b.End); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, rows.Err()
}

func (r *Repos) UpdateDeviceTaskPlan(ctx context.Context, id int64, start time.Time, end time.Time) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE device_task
		SET dvctsk_planestarttime = $2,
		    dvctsk_planecomptime  = $3
		WHERE dvctsk_id = $1
	`, id, start, end)
	return err
}
