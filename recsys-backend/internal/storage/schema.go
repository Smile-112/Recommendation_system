package storage

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed createDB.sql
var createDBSQL string

func EnsureSchema(ctx context.Context, db *pgxpool.Pool) error {
	var exists bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('public."user"') IS NOT NULL`).Scan(&exists); err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if exists {
		return ensurePlanningSchema(ctx, db)
	}
	if _, err := db.Exec(ctx, createDBSQL); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return ensurePlanningSchema(ctx, db)
}

func ensurePlanningSchema(ctx context.Context, db *pgxpool.Pool) error {
	statements := []string{
		`ALTER TABLE device_task ADD COLUMN IF NOT EXISTS equipment_characteristic INTEGER`,
		`ALTER TABLE device_task ALTER COLUMN device DROP NOT NULL`,
		`ALTER TABLE device_task ALTER COLUMN operator DROP NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_device_task__equipment_characteristic ON device_task (equipment_characteristic)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_device_task__equipment_characteristic') THEN
				ALTER TABLE device_task ADD CONSTRAINT fk_device_task__equipment_characteristic
				FOREIGN KEY (equipment_characteristic) REFERENCES eqpmnt_characteristics (eqpchrscs_id) ON DELETE SET NULL;
			END IF;
		END $$`,
		`CREATE TABLE IF NOT EXISTS planning_criteria (
			plncrt_id SERIAL PRIMARY KEY,
			plncrt_code TEXT NOT NULL UNIQUE,
			plncrt_name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS planning_weights (
			plnwgt_id SERIAL PRIMARY KEY,
			workspace INTEGER NOT NULL REFERENCES workspace (wrkspc_id) ON DELETE CASCADE,
			criterion_code TEXT NOT NULL,
			plnwgt_weight DOUBLE PRECISION NOT NULL DEFAULT 1,
			UNIQUE (workspace, criterion_code)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_planning_weights__workspace ON planning_weights (workspace)`,
		`CREATE TABLE IF NOT EXISTS device_characteristic_score (
			dvcchrsc_id SERIAL PRIMARY KEY,
			workspace INTEGER NOT NULL REFERENCES workspace (wrkspc_id) ON DELETE CASCADE,
			device INTEGER NOT NULL REFERENCES device (dvc_id) ON DELETE CASCADE,
			equipment_characteristic INTEGER NOT NULL REFERENCES eqpmnt_characteristics (eqpchrscs_id) ON DELETE CASCADE,
			dvcchrsc_score DOUBLE PRECISION NOT NULL DEFAULT 0,
			UNIQUE (workspace, device, equipment_characteristic)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_device_characteristic_score__workspace ON device_characteristic_score (workspace)`,
		`CREATE INDEX IF NOT EXISTS idx_device_characteristic_score__device ON device_characteristic_score (device)`,
		`CREATE INDEX IF NOT EXISTS idx_device_characteristic_score__characteristic ON device_characteristic_score (equipment_characteristic)`,
		`CREATE TABLE IF NOT EXISTS planning_run (
			plnrun_id SERIAL PRIMARY KEY,
			workspace INTEGER NOT NULL REFERENCES workspace (wrkspc_id) ON DELETE CASCADE,
			plnrun_started_at TIMESTAMP NOT NULL DEFAULT now(),
			plnrun_status TEXT NOT NULL DEFAULT 'completed'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_planning_run__workspace ON planning_run (workspace)`,
		`CREATE TABLE IF NOT EXISTS planning_recommendation (
			plnrec_id SERIAL PRIMARY KEY,
			planning_run INTEGER NOT NULL REFERENCES planning_run (plnrun_id) ON DELETE CASCADE,
			device_task INTEGER NOT NULL REFERENCES device_task (dvctsk_id) ON DELETE CASCADE,
			device INTEGER REFERENCES device (dvc_id) ON DELETE SET NULL,
			operator INTEGER REFERENCES operator (oprt_id) ON DELETE SET NULL,
			plnrec_start TIMESTAMP,
			plnrec_end TIMESTAMP,
			plnrec_score DOUBLE PRECISION NOT NULL DEFAULT 0,
			plnrec_selected BOOLEAN NOT NULL DEFAULT FALSE,
			plnrec_warning_code TEXT NOT NULL DEFAULT '',
			plnrec_warning_text TEXT NOT NULL DEFAULT '',
			plnrec_explanation TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_planning_recommendation__run ON planning_recommendation (planning_run)`,
		`CREATE INDEX IF NOT EXISTS idx_planning_recommendation__task ON planning_recommendation (device_task)`,
		`INSERT INTO planning_criteria (plncrt_code, plncrt_name) VALUES
			('deadline_urgency', 'Deadline urgency'),
			('manual_priority', 'Manual priority'),
			('characteristic_fit', 'Characteristic fit'),
			('device_efficiency', 'Device efficiency'),
			('early_completion', 'Early completion'),
			('lateness_penalty', 'Lateness penalty'),
			('move_penalty', 'Move penalty')
		ON CONFLICT (plncrt_code) DO NOTHING`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("ensure planning schema: %w", err)
		}
	}
	return nil
}
