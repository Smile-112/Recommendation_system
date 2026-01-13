package storage

import "context"

// ClearAllData удаляет все тестовые данные, сохраняя пользователей (dev-only helper).
func (r *Repos) ClearAllData(ctx context.Context) error {
	_, err := r.DB.Exec(ctx, `
		TRUNCATE TABLE
			user_task,
			device_task,
			operator_device,
			competencies_operator,
			operator,
			device,
			devices_type,
			eqpmnt_characteristics,
			device_tasks_type,
			workspace,
			device_state,
			priorities
		RESTART IDENTITY CASCADE
	`)
	return err
}
