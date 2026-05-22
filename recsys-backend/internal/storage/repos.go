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
	ID                        int64         `json:"id"`
	Name                      string        `json:"name"`
	Deadline                  *time.Time    `json:"deadline"`
	Duration                  time.Duration `json:"duration"`    // печать
	SetupTime                 time.Duration `json:"setup_time"`  // наладка
	UnloadTime                time.Duration `json:"unload_time"` // снятие изделия
	NeedOperator              bool          `json:"need_operator"`
	PlanStart                 *time.Time    `json:"plan_start"`
	PlanEnd                   *time.Time    `json:"plan_end"`
	DocNum                    string        `json:"doc_num"`
	PriorityID                int64         `json:"priority_id"`
	OperatorID                int64         `json:"operator_id"`
	DeviceID                  int64         `json:"device_id"`
	EquipmentCharacteristicID int64         `json:"equipment_characteristic_id"`
	DeviceTaskTypeID          int64         `json:"device_task_type_id"`
	WorkspaceID               int64         `json:"workspace_id"`
	CompletionMark            string        `json:"completion_mark"`
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
			COALESCE(operator, 0),
			COALESCE(device, 0),
			COALESCE(equipment_characteristic, 0),
			device_tasks_type,
			workspace,
			dvctsk_complitionmark
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
			&t.EquipmentCharacteristicID,
			&t.DeviceTaskTypeID,
			&t.WorkspaceID,
			&t.CompletionMark,
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
			COALESCE(operator, 0),
			COALESCE(device, 0),
			COALESCE(equipment_characteristic, 0),
			device_tasks_type,
			workspace,
			dvctsk_complitionmark
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
			&t.EquipmentCharacteristicID,
			&t.DeviceTaskTypeID,
			&t.WorkspaceID,
			&t.CompletionMark,
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

func (r *Repos) UpdateDeviceTaskPlan(ctx context.Context, id int64, deviceID int64, operatorID int64, start time.Time, end time.Time) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE device_task
		SET dvctsk_planestarttime = $2,
		    dvctsk_planecomptime  = $3,
		    device = $4,
		    operator = $5
		WHERE dvctsk_id = $1
	`, id, start, end, nullInt64(deviceID), nullInt64(operatorID))
	return err
}

func (r *Repos) ListPlanningWeights(ctx context.Context, workspaceID int64) ([]PlanningWeight, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT plnwgt_id, workspace, criterion_code, plnwgt_weight
		FROM planning_weights
		WHERE workspace = $1
		ORDER BY criterion_code
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []PlanningWeight{}
	for rows.Next() {
		var item PlanningWeight
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.CriterionCode, &item.Weight); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, rows.Err()
}

func (r *Repos) ListPlanningCriteria(ctx context.Context) ([]PlanningCriterion, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT plncrt_id, plncrt_code, plncrt_name
		FROM planning_criteria
		ORDER BY plncrt_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []PlanningCriterion{}
	for rows.Next() {
		var item PlanningCriterion
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, rows.Err()
}

func (r *Repos) UpsertPlanningWeight(ctx context.Context, item PlanningWeight) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO planning_weights (workspace, criterion_code, plnwgt_weight)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace, criterion_code)
		DO UPDATE SET plnwgt_weight = EXCLUDED.plnwgt_weight
		RETURNING plnwgt_id
	`, item.WorkspaceID, item.CriterionCode, item.Weight).Scan(&id)
	return id, err
}

func (r *Repos) ListDeviceCharacteristicScores(ctx context.Context, workspaceID int64) ([]DeviceCharacteristicScore, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT dvcchrsc_id, workspace, device, equipment_characteristic, dvcchrsc_score
		FROM device_characteristic_score
		WHERE workspace = $1
		ORDER BY device, equipment_characteristic
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []DeviceCharacteristicScore{}
	for rows.Next() {
		var item DeviceCharacteristicScore
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.DeviceID, &item.EquipmentCharacteristicID, &item.Score); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, rows.Err()
}

func (r *Repos) UpsertDeviceCharacteristicScore(ctx context.Context, item DeviceCharacteristicScore) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO device_characteristic_score (workspace, device, equipment_characteristic, dvcchrsc_score)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace, device, equipment_characteristic)
		DO UPDATE SET dvcchrsc_score = EXCLUDED.dvcchrsc_score
		RETURNING dvcchrsc_id
	`, item.WorkspaceID, item.DeviceID, item.EquipmentCharacteristicID, item.Score).Scan(&id)
	return id, err
}

func (r *Repos) CreatePlanningRun(ctx context.Context, workspaceID int64, status string) (int64, error) {
	if status == "" {
		status = "completed"
	}
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO planning_run (workspace, plnrun_status)
		VALUES ($1, $2)
		RETURNING plnrun_id
	`, workspaceID, status).Scan(&id)
	return id, err
}

func (r *Repos) CreatePlanningRecommendation(ctx context.Context, item PlanningRecommendation) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO planning_recommendation (
			planning_run,
			device_task,
			device,
			operator,
			plnrec_start,
			plnrec_end,
			plnrec_score,
			plnrec_selected,
			plnrec_warning_code,
			plnrec_warning_text,
			plnrec_explanation
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING plnrec_id
	`,
		item.RunID,
		item.TaskID,
		nullInt64(item.DeviceID),
		nullInt64(item.OperatorID),
		item.Start,
		item.End,
		item.Score,
		item.Selected,
		item.WarningCode,
		item.WarningText,
		item.Explanation,
	).Scan(&id)
	return id, err
}
