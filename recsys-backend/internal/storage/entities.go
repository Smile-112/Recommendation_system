package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type Workspace struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	UserLogin string `json:"user_login"`
}

type DeviceState struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Priority struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type EquipmentCharacteristic struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	WorkspaceID int64  `json:"workspace_id"`
}

type DeviceType struct {
	ID                        int64  `json:"id"`
	Name                      string `json:"name"`
	EquipmentCharacteristicID int64  `json:"equipment_characteristic_id"`
	WorkspaceID               int64  `json:"workspace_id"`
}

type Device struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PhotoURL       string `json:"photo_url"`
	AddInRecSystem *bool  `json:"add_in_rec_system"`
	DeviceTypeID   int64  `json:"device_type_id"`
	DeviceStateID  int64  `json:"device_state_id"`
	WorkspaceID    int64  `json:"workspace_id"`
}

type Operator struct {
	ID          int64  `json:"id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	WorkspaceID int64  `json:"workspace_id"`
	UserLogin   string `json:"user_login"`
}

type OperatorCompetency struct {
	ID           int64 `json:"id"`
	WorkspaceID  int64 `json:"workspace_id"`
	DeviceTypeID int64 `json:"device_type_id"`
	OperatorID   int64 `json:"operator_id"`
}

type OperatorDevice struct {
	ID         int64 `json:"id"`
	OperatorID int64 `json:"operator_id"`
	DeviceID   int64 `json:"device_id"`
}

type DeviceTaskType struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	WorkspaceID int64  `json:"workspace_id"`
}

type DeviceTask struct {
	ID               int64         `json:"id"`
	Name             string        `json:"name"`
	Deadline         *time.Time    `json:"deadline"`
	Duration         time.Duration `json:"duration"`
	SetupTime        time.Duration `json:"setup_time"`
	UnloadTime       time.Duration `json:"unload_time"`
	NeedOperator     bool          `json:"need_operator"`
	PhotoURL         string        `json:"photo_url"`
	PlanStart        *time.Time    `json:"plan_start"`
	PlanEnd          *time.Time    `json:"plan_end"`
	DocNum           string        `json:"doc_num"`
	CompletionMark   string        `json:"completion_mark"`
	AddInRecSystem   *bool         `json:"add_in_rec_system"`
	DeviceTaskTypeID int64         `json:"device_task_type_id"`
	WorkspaceID      int64         `json:"workspace_id"`
	OperatorID       int64         `json:"operator_id"`
	DeviceID         int64         `json:"device_id"`
	PriorityID       int64         `json:"priority_id"`
}

type UserTask struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Priority       *int       `json:"priority"`
	CompletionMark *bool      `json:"completion_mark"`
	WorkspaceID    int64      `json:"workspace_id"`
	DeviceTaskID   int64      `json:"device_task_id"`
	OperatorID     int64      `json:"operator_id"`
}

func (r *Repos) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := r.DB.Query(ctx, `SELECT user_login, user_id, user_email FROM "user" ORDER BY user_login`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Login, &u.ID, &u.Email); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, rows.Err()
}

func (r *Repos) GetUser(ctx context.Context, login string) (User, error) {
	var u User
	err := r.DB.QueryRow(ctx, `SELECT user_login, user_id, user_email FROM "user" WHERE user_login = $1`, login).
		Scan(&u.Login, &u.ID, &u.Email)
	return u, err
}

func (r *Repos) CreateUser(ctx context.Context, u User, passwordHash string) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO "user" (user_login, user_id, user_password, user_email)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id
	`, u.Login, u.ID, passwordHash, u.Email).Scan(&id)
	return id, err
}

func (r *Repos) UpdateUser(ctx context.Context, u User, passwordHash *string) error {
	if passwordHash != nil {
		_, err := r.DB.Exec(ctx, `
			UPDATE "user"
			SET user_id = $2,
				user_password = $3,
				user_email = $4
			WHERE user_login = $1
		`, u.Login, u.ID, *passwordHash, u.Email)
		return err
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE "user"
		SET user_id = $2,
			user_email = $3
		WHERE user_login = $1
	`, u.Login, u.ID, u.Email)
	return err
}

func (r *Repos) DeleteUser(ctx context.Context, login string) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM "user" WHERE user_login = $1`, login)
	return err
}

func (r *Repos) ListWorkspaces(ctx context.Context, userLogin *string) ([]Workspace, error) {
	query := `SELECT wrkspc_id, wrkspc_name, "user" FROM workspace`
	args := []any{}
	if userLogin != nil {
		query += ` WHERE "user" = $1`
		args = append(args, *userLogin)
	}
	query += ` ORDER BY wrkspc_id`

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.UserLogin); err != nil {
			return nil, err
		}
		res = append(res, w)
	}
	return res, rows.Err()
}

func (r *Repos) GetWorkspace(ctx context.Context, id int64) (Workspace, error) {
	var w Workspace
	err := r.DB.QueryRow(ctx, `SELECT wrkspc_id, wrkspc_name, "user" FROM workspace WHERE wrkspc_id = $1`, id).
		Scan(&w.ID, &w.Name, &w.UserLogin)
	return w, err
}

func (r *Repos) CreateWorkspace(ctx context.Context, w Workspace) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO workspace (wrkspc_name, "user")
		VALUES ($1, $2)
		RETURNING wrkspc_id
	`, w.Name, w.UserLogin).Scan(&id)
	return id, err
}

func (r *Repos) UpdateWorkspace(ctx context.Context, w Workspace) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE workspace
		SET wrkspc_name = $2, "user" = $3
		WHERE wrkspc_id = $1
	`, w.ID, w.Name, w.UserLogin)
	return err
}

func (r *Repos) DeleteWorkspace(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM workspace WHERE wrkspc_id = $1`, id)
	return err
}

func (r *Repos) ListDeviceStates(ctx context.Context) ([]DeviceState, error) {
	rows, err := r.DB.Query(ctx, `SELECT dvcst_id, dvcst_name FROM device_state ORDER BY dvcst_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DeviceState
	for rows.Next() {
		var s DeviceState
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *Repos) GetDeviceState(ctx context.Context, id int64) (DeviceState, error) {
	var s DeviceState
	err := r.DB.QueryRow(ctx, `SELECT dvcst_id, dvcst_name FROM device_state WHERE dvcst_id = $1`, id).
		Scan(&s.ID, &s.Name)
	return s, err
}

func (r *Repos) CreateDeviceState(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `INSERT INTO device_state (dvcst_name) VALUES ($1) RETURNING dvcst_id`, name).Scan(&id)
	return id, err
}

func (r *Repos) UpdateDeviceState(ctx context.Context, id int64, name string) error {
	_, err := r.DB.Exec(ctx, `UPDATE device_state SET dvcst_name = $2 WHERE dvcst_id = $1`, id, name)
	return err
}

func (r *Repos) DeleteDeviceState(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM device_state WHERE dvcst_id = $1`, id)
	return err
}

func (r *Repos) ListPriorities(ctx context.Context) ([]Priority, error) {
	rows, err := r.DB.Query(ctx, `SELECT prts_id, prts_name FROM priorities ORDER BY prts_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Priority
	for rows.Next() {
		var p Priority
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, rows.Err()
}

func (r *Repos) GetPriority(ctx context.Context, id int64) (Priority, error) {
	var p Priority
	err := r.DB.QueryRow(ctx, `SELECT prts_id, prts_name FROM priorities WHERE prts_id = $1`, id).
		Scan(&p.ID, &p.Name)
	return p, err
}

func (r *Repos) CreatePriority(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `INSERT INTO priorities (prts_name) VALUES ($1) RETURNING prts_id`, name).Scan(&id)
	return id, err
}

func (r *Repos) UpdatePriority(ctx context.Context, id int64, name string) error {
	_, err := r.DB.Exec(ctx, `UPDATE priorities SET prts_name = $2 WHERE prts_id = $1`, id, name)
	return err
}

func (r *Repos) DeletePriority(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM priorities WHERE prts_id = $1`, id)
	return err
}

func (r *Repos) ListEquipmentCharacteristics(ctx context.Context, workspaceID int64) ([]EquipmentCharacteristic, error) {
	rows, err := r.DB.Query(ctx, `SELECT eqpchrscs_id, eqpchrscs_name, workspace FROM eqpmnt_characteristics WHERE workspace = $1 ORDER BY eqpchrscs_id`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []EquipmentCharacteristic
	for rows.Next() {
		var c EquipmentCharacteristic
		if err := rows.Scan(&c.ID, &c.Name, &c.WorkspaceID); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func (r *Repos) GetEquipmentCharacteristic(ctx context.Context, id int64) (EquipmentCharacteristic, error) {
	var c EquipmentCharacteristic
	err := r.DB.QueryRow(ctx, `SELECT eqpchrscs_id, eqpchrscs_name, workspace FROM eqpmnt_characteristics WHERE eqpchrscs_id = $1`, id).
		Scan(&c.ID, &c.Name, &c.WorkspaceID)
	return c, err
}

func (r *Repos) CreateEquipmentCharacteristic(ctx context.Context, c EquipmentCharacteristic) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO eqpmnt_characteristics (eqpchrscs_name, workspace)
		VALUES ($1, $2)
		RETURNING eqpchrscs_id
	`, c.Name, c.WorkspaceID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateEquipmentCharacteristic(ctx context.Context, c EquipmentCharacteristic) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE eqpmnt_characteristics
		SET eqpchrscs_name = $2, workspace = $3
		WHERE eqpchrscs_id = $1
	`, c.ID, c.Name, c.WorkspaceID)
	return err
}

func (r *Repos) DeleteEquipmentCharacteristic(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM eqpmnt_characteristics WHERE eqpchrscs_id = $1`, id)
	return err
}

func (r *Repos) ListDeviceTypes(ctx context.Context, workspaceID int64) ([]DeviceType, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT dvctp_id, dvctp_name, eqpmnt_characteristics, workspace
		FROM devices_type
		WHERE workspace = $1
		ORDER BY dvctp_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DeviceType
	for rows.Next() {
		var t DeviceType
		if err := rows.Scan(&t.ID, &t.Name, &t.EquipmentCharacteristicID, &t.WorkspaceID); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, rows.Err()
}

func (r *Repos) GetDeviceType(ctx context.Context, id int64) (DeviceType, error) {
	var t DeviceType
	err := r.DB.QueryRow(ctx, `
		SELECT dvctp_id, dvctp_name, eqpmnt_characteristics, workspace
		FROM devices_type
		WHERE dvctp_id = $1
	`, id).Scan(&t.ID, &t.Name, &t.EquipmentCharacteristicID, &t.WorkspaceID)
	return t, err
}

func (r *Repos) CreateDeviceType(ctx context.Context, t DeviceType) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO devices_type (dvctp_name, eqpmnt_characteristics, workspace)
		VALUES ($1, $2, $3)
		RETURNING dvctp_id
	`, t.Name, t.EquipmentCharacteristicID, t.WorkspaceID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateDeviceType(ctx context.Context, t DeviceType) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE devices_type
		SET dvctp_name = $2, eqpmnt_characteristics = $3, workspace = $4
		WHERE dvctp_id = $1
	`, t.ID, t.Name, t.EquipmentCharacteristicID, t.WorkspaceID)
	return err
}

func (r *Repos) DeleteDeviceType(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM devices_type WHERE dvctp_id = $1`, id)
	return err
}

func (r *Repos) ListDevices(ctx context.Context, workspaceID int64) ([]Device, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT dvc_id, dvc_name, dvc_photourl, dvc_addinrecsystem, devices__type, device_state, workspace
		FROM device
		WHERE workspace = $1
		ORDER BY dvc_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.Name, &d.PhotoURL, &d.AddInRecSystem, &d.DeviceTypeID, &d.DeviceStateID, &d.WorkspaceID); err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, rows.Err()
}

func (r *Repos) GetDevice(ctx context.Context, id int64) (Device, error) {
	var d Device
	err := r.DB.QueryRow(ctx, `
		SELECT dvc_id, dvc_name, dvc_photourl, dvc_addinrecsystem, devices__type, device_state, workspace
		FROM device
		WHERE dvc_id = $1
	`, id).Scan(&d.ID, &d.Name, &d.PhotoURL, &d.AddInRecSystem, &d.DeviceTypeID, &d.DeviceStateID, &d.WorkspaceID)
	return d, err
}

func (r *Repos) CreateDevice(ctx context.Context, d Device) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO device (dvc_name, dvc_photourl, dvc_addinrecsystem, devices__type, device_state, workspace)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING dvc_id
	`, d.Name, d.PhotoURL, d.AddInRecSystem, d.DeviceTypeID, d.DeviceStateID, d.WorkspaceID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateDevice(ctx context.Context, d Device) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE device
		SET dvc_name = $2,
			dvc_photourl = $3,
			dvc_addinrecsystem = $4,
			devices__type = $5,
			device_state = $6,
			workspace = $7
		WHERE dvc_id = $1
	`, d.ID, d.Name, d.PhotoURL, d.AddInRecSystem, d.DeviceTypeID, d.DeviceStateID, d.WorkspaceID)
	return err
}

func (r *Repos) DeleteDevice(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM device WHERE dvc_id = $1`, id)
	return err
}

func (r *Repos) ListOperators(ctx context.Context, workspaceID int64) ([]Operator, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT oprt_id, oprt_fio, oprt_phnnm, workspace, "user"
		FROM operator
		WHERE workspace = $1
		ORDER BY oprt_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Operator
	for rows.Next() {
		var o Operator
		if err := rows.Scan(&o.ID, &o.FullName, &o.PhoneNumber, &o.WorkspaceID, &o.UserLogin); err != nil {
			return nil, err
		}
		res = append(res, o)
	}
	return res, rows.Err()
}

func (r *Repos) GetOperator(ctx context.Context, id int64) (Operator, error) {
	var o Operator
	err := r.DB.QueryRow(ctx, `
		SELECT oprt_id, oprt_fio, oprt_phnnm, workspace, "user"
		FROM operator
		WHERE oprt_id = $1
	`, id).Scan(&o.ID, &o.FullName, &o.PhoneNumber, &o.WorkspaceID, &o.UserLogin)
	return o, err
}

func (r *Repos) CreateOperator(ctx context.Context, o Operator) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO operator (oprt_fio, oprt_phnnm, workspace, "user")
		VALUES ($1, $2, $3, $4)
		RETURNING oprt_id
	`, o.FullName, o.PhoneNumber, o.WorkspaceID, o.UserLogin).Scan(&id)
	return id, err
}

func (r *Repos) UpdateOperator(ctx context.Context, o Operator) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE operator
		SET oprt_fio = $2,
			oprt_phnnm = $3,
			workspace = $4,
			"user" = $5
		WHERE oprt_id = $1
	`, o.ID, o.FullName, o.PhoneNumber, o.WorkspaceID, o.UserLogin)
	return err
}

func (r *Repos) DeleteOperator(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM operator WHERE oprt_id = $1`, id)
	return err
}

func (r *Repos) ListOperatorCompetencies(ctx context.Context, workspaceID int64) ([]OperatorCompetency, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT compt_oprt_id, workspace, devices__type, operator
		FROM competencies_operator
		WHERE workspace = $1
		ORDER BY compt_oprt_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []OperatorCompetency
	for rows.Next() {
		var c OperatorCompetency
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.DeviceTypeID, &c.OperatorID); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func (r *Repos) GetOperatorCompetency(ctx context.Context, id int64) (OperatorCompetency, error) {
	var c OperatorCompetency
	err := r.DB.QueryRow(ctx, `
		SELECT compt_oprt_id, workspace, devices__type, operator
		FROM competencies_operator
		WHERE compt_oprt_id = $1
	`, id).Scan(&c.ID, &c.WorkspaceID, &c.DeviceTypeID, &c.OperatorID)
	return c, err
}

func (r *Repos) CreateOperatorCompetency(ctx context.Context, c OperatorCompetency) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO competencies_operator (workspace, devices__type, operator)
		VALUES ($1, $2, $3)
		RETURNING compt_oprt_id
	`, c.WorkspaceID, c.DeviceTypeID, c.OperatorID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateOperatorCompetency(ctx context.Context, c OperatorCompetency) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE competencies_operator
		SET workspace = $2, devices__type = $3, operator = $4
		WHERE compt_oprt_id = $1
	`, c.ID, c.WorkspaceID, c.DeviceTypeID, c.OperatorID)
	return err
}

func (r *Repos) DeleteOperatorCompetency(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM competencies_operator WHERE compt_oprt_id = $1`, id)
	return err
}

func (r *Repos) ListOperatorDevices(ctx context.Context, workspaceID int64) ([]OperatorDevice, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT dvc_oprt_id, operator, device
		FROM operator_device
		WHERE operator IN (SELECT oprt_id FROM operator WHERE workspace = $1)
		ORDER BY dvc_oprt_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []OperatorDevice
	for rows.Next() {
		var d OperatorDevice
		if err := rows.Scan(&d.ID, &d.OperatorID, &d.DeviceID); err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, rows.Err()
}

func (r *Repos) GetOperatorDevice(ctx context.Context, id int64) (OperatorDevice, error) {
	var d OperatorDevice
	err := r.DB.QueryRow(ctx, `
		SELECT dvc_oprt_id, operator, device
		FROM operator_device
		WHERE dvc_oprt_id = $1
	`, id).Scan(&d.ID, &d.OperatorID, &d.DeviceID)
	return d, err
}

func (r *Repos) CreateOperatorDevice(ctx context.Context, d OperatorDevice) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO operator_device (operator, device)
		VALUES ($1, $2)
		RETURNING dvc_oprt_id
	`, d.OperatorID, d.DeviceID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateOperatorDevice(ctx context.Context, d OperatorDevice) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE operator_device
		SET operator = $2, device = $3
		WHERE dvc_oprt_id = $1
	`, d.ID, d.OperatorID, d.DeviceID)
	return err
}

func (r *Repos) DeleteOperatorDevice(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM operator_device WHERE dvc_oprt_id = $1`, id)
	return err
}

func (r *Repos) ListDeviceTaskTypes(ctx context.Context, workspaceID int64) ([]DeviceTaskType, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT dvctsktp_id, dvctsktp_name, workspace
		FROM device_tasks_type
		WHERE workspace = $1
		ORDER BY dvctsktp_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []DeviceTaskType
	for rows.Next() {
		var t DeviceTaskType
		if err := rows.Scan(&t.ID, &t.Name, &t.WorkspaceID); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, rows.Err()
}

func (r *Repos) GetDeviceTaskType(ctx context.Context, id int64) (DeviceTaskType, error) {
	var t DeviceTaskType
	err := r.DB.QueryRow(ctx, `
		SELECT dvctsktp_id, dvctsktp_name, workspace
		FROM device_tasks_type
		WHERE dvctsktp_id = $1
	`, id).Scan(&t.ID, &t.Name, &t.WorkspaceID)
	return t, err
}

func (r *Repos) CreateDeviceTaskType(ctx context.Context, t DeviceTaskType) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO device_tasks_type (dvctsktp_name, workspace)
		VALUES ($1, $2)
		RETURNING dvctsktp_id
	`, t.Name, t.WorkspaceID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateDeviceTaskType(ctx context.Context, t DeviceTaskType) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE device_tasks_type
		SET dvctsktp_name = $2, workspace = $3
		WHERE dvctsktp_id = $1
	`, t.ID, t.Name, t.WorkspaceID)
	return err
}

func (r *Repos) DeleteDeviceTaskType(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM device_tasks_type WHERE dvctsktp_id = $1`, id)
	return err
}

func (r *Repos) GetDeviceTask(ctx context.Context, id int64) (DeviceTask, error) {
	var t DeviceTask
	var duration pgtype.Time
	var setup pgtype.Time
	var unload pgtype.Time
	err := r.DB.QueryRow(ctx, `
		SELECT dvctsk_id, dvctsk_name, dvctsk_deadline, dvctsk_duration, dvctsk_setuptime,
			dvctsk_timetocomplite, COALESCE(dvctsk_needoperator,false), dvctsk_photourl,
			dvctsk_planestarttime, dvctsk_planecomptime, dvctsk_docnum, dvctsk_complitionmark,
			dvctsk_addinrecsystem, device_tasks_type, workspace, operator, device, priorities
		FROM device_task
		WHERE dvctsk_id = $1
	`, id).Scan(
		&t.ID,
		&t.Name,
		&t.Deadline,
		&duration,
		&setup,
		&unload,
		&t.NeedOperator,
		&t.PhotoURL,
		&t.PlanStart,
		&t.PlanEnd,
		&t.DocNum,
		&t.CompletionMark,
		&t.AddInRecSystem,
		&t.DeviceTaskTypeID,
		&t.WorkspaceID,
		&t.OperatorID,
		&t.DeviceID,
		&t.PriorityID,
	)
	if err != nil {
		return t, err
	}
	t.Duration = timeToDuration(duration)
	t.SetupTime = timeToDuration(setup)
	t.UnloadTime = timeToDuration(unload)
	return t, nil
}

func (r *Repos) CreateDeviceTask(ctx context.Context, t DeviceTask) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO device_task (
			dvctsk_name,
			dvctsk_deadline,
			dvctsk_duration,
			dvctsk_needoperator,
			dvctsk_photourl,
			dvctsk_planestarttime,
			dvctsk_planecomptime,
			dvctsk_docnum,
			dvctsk_setuptime,
			dvctsk_timetocomplite,
			dvctsk_complitionmark,
			dvctsk_addinrecsystem,
			device_tasks_type,
			workspace,
			operator,
			device,
			priorities
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING dvctsk_id
	`,
		t.Name,
		t.Deadline,
		formatDuration(t.Duration),
		t.NeedOperator,
		t.PhotoURL,
		t.PlanStart,
		t.PlanEnd,
		t.DocNum,
		formatDuration(t.SetupTime),
		formatDuration(t.UnloadTime),
		t.CompletionMark,
		t.AddInRecSystem,
		t.DeviceTaskTypeID,
		t.WorkspaceID,
		t.OperatorID,
		t.DeviceID,
		t.PriorityID,
	).Scan(&id)
	return id, err
}

func (r *Repos) UpdateDeviceTask(ctx context.Context, t DeviceTask) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE device_task
		SET dvctsk_name = $2,
			dvctsk_deadline = $3,
			dvctsk_duration = $4,
			dvctsk_needoperator = $5,
			dvctsk_photourl = $6,
			dvctsk_planestarttime = $7,
			dvctsk_planecomptime = $8,
			dvctsk_docnum = $9,
			dvctsk_setuptime = $10,
			dvctsk_timetocomplite = $11,
			dvctsk_complitionmark = $12,
			dvctsk_addinrecsystem = $13,
			device_tasks_type = $14,
			workspace = $15,
			operator = $16,
			device = $17,
			priorities = $18
		WHERE dvctsk_id = $1
	`,
		t.ID,
		t.Name,
		t.Deadline,
		formatDuration(t.Duration),
		t.NeedOperator,
		t.PhotoURL,
		t.PlanStart,
		t.PlanEnd,
		t.DocNum,
		formatDuration(t.SetupTime),
		formatDuration(t.UnloadTime),
		t.CompletionMark,
		t.AddInRecSystem,
		t.DeviceTaskTypeID,
		t.WorkspaceID,
		t.OperatorID,
		t.DeviceID,
		t.PriorityID,
	)
	return err
}

func (r *Repos) DeleteDeviceTask(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM device_task WHERE dvctsk_id = $1`, id)
	return err
}

func (r *Repos) ListUserTasks(ctx context.Context, workspaceID int64) ([]UserTask, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT usertsk_id, usertsk_name, usertsk_starttime, usertsk_endtime,
			usertsk_priority, usertsk_complitionmark, workspace, device_task, operator
		FROM user_task
		WHERE workspace = $1
		ORDER BY usertsk_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []UserTask
	for rows.Next() {
		var t UserTask
		if err := rows.Scan(&t.ID, &t.Name, &t.StartTime, &t.EndTime, &t.Priority, &t.CompletionMark, &t.WorkspaceID, &t.DeviceTaskID, &t.OperatorID); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, rows.Err()
}

func (r *Repos) GetUserTask(ctx context.Context, id int64) (UserTask, error) {
	var t UserTask
	err := r.DB.QueryRow(ctx, `
		SELECT usertsk_id, usertsk_name, usertsk_starttime, usertsk_endtime,
			usertsk_priority, usertsk_complitionmark, workspace, device_task, operator
		FROM user_task
		WHERE usertsk_id = $1
	`, id).Scan(&t.ID, &t.Name, &t.StartTime, &t.EndTime, &t.Priority, &t.CompletionMark, &t.WorkspaceID, &t.DeviceTaskID, &t.OperatorID)
	return t, err
}

func (r *Repos) CreateUserTask(ctx context.Context, t UserTask) (int64, error) {
	var id int64
	err := r.DB.QueryRow(ctx, `
		INSERT INTO user_task (
			usertsk_name,
			usertsk_starttime,
			usertsk_endtime,
			usertsk_priority,
			usertsk_complitionmark,
			workspace,
			device_task,
			operator
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING usertsk_id
	`, t.Name, t.StartTime, t.EndTime, t.Priority, t.CompletionMark, t.WorkspaceID, t.DeviceTaskID, t.OperatorID).Scan(&id)
	return id, err
}

func (r *Repos) UpdateUserTask(ctx context.Context, t UserTask) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE user_task
		SET usertsk_name = $2,
			usertsk_starttime = $3,
			usertsk_endtime = $4,
			usertsk_priority = $5,
			usertsk_complitionmark = $6,
			workspace = $7,
			device_task = $8,
			operator = $9
		WHERE usertsk_id = $1
	`, t.ID, t.Name, t.StartTime, t.EndTime, t.Priority, t.CompletionMark, t.WorkspaceID, t.DeviceTaskID, t.OperatorID)
	return err
}

func (r *Repos) DeleteUserTask(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM user_task WHERE usertsk_id = $1`, id)
	return err
}
