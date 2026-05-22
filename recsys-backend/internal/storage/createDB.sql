CREATE TABLE "device_state" (
  "dvcst_id" SERIAL PRIMARY KEY,
  "dvcst_name" TEXT NOT NULL
);

CREATE TABLE "priorities" (
  "prts_id" SERIAL PRIMARY KEY,
  "prts_name" TEXT NOT NULL
);

CREATE TABLE "user" (
  "user_login" TEXT PRIMARY KEY,
  "user_id" INTEGER NOT NULL,
  "user_password" TEXT NOT NULL,
  "user_email" TEXT NOT NULL,
  "is_admin" BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE "workspace" (
  "wrkspc_id" SERIAL PRIMARY KEY,
  "wrkspc_name" TEXT NOT NULL,
  "user" TEXT NOT NULL
);

CREATE INDEX "idx_workspace__user" ON "workspace" ("user");

ALTER TABLE "workspace" ADD CONSTRAINT "fk_workspace__user" FOREIGN KEY ("user") REFERENCES "user" ("user_login") ON DELETE CASCADE;

CREATE TABLE "device_tasks_type" (
  "dvctsktp_id" SERIAL PRIMARY KEY,
  "dvctsktp_name" TEXT NOT NULL,
  "workspace" INTEGER NOT NULL
);

CREATE INDEX "idx_device_tasks_type__workspace" ON "device_tasks_type" ("workspace");

ALTER TABLE "device_tasks_type" ADD CONSTRAINT "fk_device_tasks_type__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "eqpmnt_characteristics" (
  "eqpchrscs_id" SERIAL PRIMARY KEY,
  "eqpchrscs_name" TEXT NOT NULL,
  "workspace" INTEGER NOT NULL
);

CREATE INDEX "idx_eqpmnt_characteristics__workspace" ON "eqpmnt_characteristics" ("workspace");

ALTER TABLE "eqpmnt_characteristics" ADD CONSTRAINT "fk_eqpmnt_characteristics__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "devices_type" (
  "dvctp_id" SERIAL PRIMARY KEY,
  "dvctp_name" TEXT NOT NULL,
  "eqpmnt_characteristics" INTEGER NOT NULL,
  "workspace" INTEGER NOT NULL
);

CREATE INDEX "idx_devices_type__eqpmnt_characteristics" ON "devices_type" ("eqpmnt_characteristics");

CREATE INDEX "idx_devices_type__workspace" ON "devices_type" ("workspace");

ALTER TABLE "devices_type" ADD CONSTRAINT "fk_devices_type__eqpmnt_characteristics" FOREIGN KEY ("eqpmnt_characteristics") REFERENCES "eqpmnt_characteristics" ("eqpchrscs_id") ON DELETE CASCADE;

ALTER TABLE "devices_type" ADD CONSTRAINT "fk_devices_type__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "device" (
  "dvc_id" SERIAL PRIMARY KEY,
  "dvc_name" TEXT NOT NULL,
  "dvc_photourl" TEXT NOT NULL,
  "dvc_addinrecsystem" BOOLEAN,
  "devices__type" INTEGER NOT NULL,
  "device_state" INTEGER NOT NULL,
  "workspace" INTEGER NOT NULL
);

CREATE INDEX "idx_device__device_state" ON "device" ("device_state");

CREATE INDEX "idx_device__devices__type" ON "device" ("devices__type");

CREATE INDEX "idx_device__workspace" ON "device" ("workspace");

ALTER TABLE "device" ADD CONSTRAINT "fk_device__device_state" FOREIGN KEY ("device_state") REFERENCES "device_state" ("dvcst_id") ON DELETE CASCADE;

ALTER TABLE "device" ADD CONSTRAINT "fk_device__devices__type" FOREIGN KEY ("devices__type") REFERENCES "devices_type" ("dvctp_id") ON DELETE CASCADE;

ALTER TABLE "device" ADD CONSTRAINT "fk_device__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "operator" (
  "oprt_id" SERIAL PRIMARY KEY,
  "oprt_fio" TEXT NOT NULL,
  "oprt_phnnm" TEXT NOT NULL,
  "workspace" INTEGER NOT NULL,
  "user" TEXT NOT NULL
);

CREATE INDEX "idx_operator__user" ON "operator" ("user");

CREATE INDEX "idx_operator__workspace" ON "operator" ("workspace");

ALTER TABLE "operator" ADD CONSTRAINT "fk_operator__user" FOREIGN KEY ("user") REFERENCES "user" ("user_login") ON DELETE CASCADE;

ALTER TABLE "operator" ADD CONSTRAINT "fk_operator__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "competencies_operator" (
  "compt_oprt_id" SERIAL PRIMARY KEY,
  "workspace" INTEGER NOT NULL,
  "devices__type" INTEGER NOT NULL,
  "operator" INTEGER NOT NULL
);

CREATE INDEX "idx_competencies_operator__devices__type" ON "competencies_operator" ("devices__type");

CREATE INDEX "idx_competencies_operator__operator" ON "competencies_operator" ("operator");

CREATE INDEX "idx_competencies_operator__workspace" ON "competencies_operator" ("workspace");

ALTER TABLE "competencies_operator" ADD CONSTRAINT "fk_competencies_operator__devices__type" FOREIGN KEY ("devices__type") REFERENCES "devices_type" ("dvctp_id") ON DELETE CASCADE;

ALTER TABLE "competencies_operator" ADD CONSTRAINT "fk_competencies_operator__operator" FOREIGN KEY ("operator") REFERENCES "operator" ("oprt_id") ON DELETE CASCADE;

ALTER TABLE "competencies_operator" ADD CONSTRAINT "fk_competencies_operator__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "device_task" (
  "dvctsk_id" SERIAL PRIMARY KEY,
  "dvctsk_name" TEXT NOT NULL,
  "dvctsk_deadline" TIMESTAMP,
  "dvctsk_duration" TIME,
  "dvctsk_needoperator" BOOLEAN,
  "dvctsk_photourl" TEXT NOT NULL,
  "dvctsk_planestarttime" TIMESTAMP,
  "dvctsk_planecomptime" TIMESTAMP,
  "dvctsk_docnum" TEXT NOT NULL,
  "dvctsk_setuptime" TIME NOT NULL,
  "dvctsk_timetocomplite" TIME NOT NULL,
  "dvctsk_complitionmark" TEXT NOT NULL,
  "dvctsk_addinrecsystem" BOOLEAN,
  "equipment_characteristic" INTEGER,
  "device_tasks_type" INTEGER NOT NULL,
  "workspace" INTEGER NOT NULL,
  "operator" INTEGER,
  "device" INTEGER,
  "priorities" INTEGER NOT NULL
);

CREATE INDEX "idx_device_task__device" ON "device_task" ("device");

CREATE INDEX "idx_device_task__device_tasks_type" ON "device_task" ("device_tasks_type");

CREATE INDEX "idx_device_task__equipment_characteristic" ON "device_task" ("equipment_characteristic");

CREATE INDEX "idx_device_task__operator" ON "device_task" ("operator");

CREATE INDEX "idx_device_task__priorities" ON "device_task" ("priorities");

CREATE INDEX "idx_device_task__workspace" ON "device_task" ("workspace");

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__device" FOREIGN KEY ("device") REFERENCES "device" ("dvc_id") ON DELETE CASCADE;

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__device_tasks_type" FOREIGN KEY ("device_tasks_type") REFERENCES "device_tasks_type" ("dvctsktp_id") ON DELETE CASCADE;

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__equipment_characteristic" FOREIGN KEY ("equipment_characteristic") REFERENCES "eqpmnt_characteristics" ("eqpchrscs_id") ON DELETE SET NULL;

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__operator" FOREIGN KEY ("operator") REFERENCES "operator" ("oprt_id") ON DELETE CASCADE;

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__priorities" FOREIGN KEY ("priorities") REFERENCES "priorities" ("prts_id") ON DELETE CASCADE;

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "operator_device" (
  "dvc_oprt_id" SERIAL PRIMARY KEY,
  "operator" INTEGER NOT NULL,
  "device" INTEGER NOT NULL
);

CREATE INDEX "idx_operator_device__device" ON "operator_device" ("device");

CREATE INDEX "idx_operator_device__operator" ON "operator_device" ("operator");

ALTER TABLE "operator_device" ADD CONSTRAINT "fk_operator_device__device" FOREIGN KEY ("device") REFERENCES "device" ("dvc_id") ON DELETE CASCADE;

ALTER TABLE "operator_device" ADD CONSTRAINT "fk_operator_device__operator" FOREIGN KEY ("operator") REFERENCES "operator" ("oprt_id") ON DELETE CASCADE;

CREATE TABLE "user_task" (
  "usertsk_id" SERIAL PRIMARY KEY,
  "usertsk_name" TEXT NOT NULL,
  "usertsk_starttime" TIMESTAMP,
  "usertsk_endtime" TIMESTAMP,
  "usertsk_priority" INTEGER,
  "usertsk_complitionmark" BOOLEAN,
  "workspace" INTEGER NOT NULL,
  "device_task" INTEGER,
  "operator" INTEGER NOT NULL
);

CREATE INDEX "idx_user_task__device_task" ON "user_task" ("device_task");

CREATE INDEX "idx_user_task__operator" ON "user_task" ("operator");

CREATE INDEX "idx_user_task__workspace" ON "user_task" ("workspace");

ALTER TABLE "user_task" ADD CONSTRAINT "fk_user_task__device_task" FOREIGN KEY ("device_task") REFERENCES "device_task" ("dvctsk_id") ON DELETE CASCADE;

ALTER TABLE "user_task" ADD CONSTRAINT "fk_user_task__operator" FOREIGN KEY ("operator") REFERENCES "operator" ("oprt_id") ON DELETE CASCADE;

ALTER TABLE "user_task" ADD CONSTRAINT "fk_user_task__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "planning_criteria" (
  "plncrt_id" SERIAL PRIMARY KEY,
  "plncrt_code" TEXT NOT NULL UNIQUE,
  "plncrt_name" TEXT NOT NULL
);

CREATE TABLE "planning_weights" (
  "plnwgt_id" SERIAL PRIMARY KEY,
  "workspace" INTEGER NOT NULL,
  "criterion_code" TEXT NOT NULL,
  "plnwgt_weight" DOUBLE PRECISION NOT NULL DEFAULT 1,
  UNIQUE ("workspace", "criterion_code")
);

CREATE INDEX "idx_planning_weights__workspace" ON "planning_weights" ("workspace");

ALTER TABLE "planning_weights" ADD CONSTRAINT "fk_planning_weights__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "device_characteristic_score" (
  "dvcchrsc_id" SERIAL PRIMARY KEY,
  "workspace" INTEGER NOT NULL,
  "device" INTEGER NOT NULL,
  "equipment_characteristic" INTEGER NOT NULL,
  "dvcchrsc_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
  UNIQUE ("workspace", "device", "equipment_characteristic")
);

CREATE INDEX "idx_device_characteristic_score__workspace" ON "device_characteristic_score" ("workspace");
CREATE INDEX "idx_device_characteristic_score__device" ON "device_characteristic_score" ("device");
CREATE INDEX "idx_device_characteristic_score__characteristic" ON "device_characteristic_score" ("equipment_characteristic");

ALTER TABLE "device_characteristic_score" ADD CONSTRAINT "fk_device_characteristic_score__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;
ALTER TABLE "device_characteristic_score" ADD CONSTRAINT "fk_device_characteristic_score__device" FOREIGN KEY ("device") REFERENCES "device" ("dvc_id") ON DELETE CASCADE;
ALTER TABLE "device_characteristic_score" ADD CONSTRAINT "fk_device_characteristic_score__characteristic" FOREIGN KEY ("equipment_characteristic") REFERENCES "eqpmnt_characteristics" ("eqpchrscs_id") ON DELETE CASCADE;

CREATE TABLE "planning_run" (
  "plnrun_id" SERIAL PRIMARY KEY,
  "workspace" INTEGER NOT NULL,
  "plnrun_started_at" TIMESTAMP NOT NULL DEFAULT now(),
  "plnrun_status" TEXT NOT NULL DEFAULT 'completed'
);

CREATE INDEX "idx_planning_run__workspace" ON "planning_run" ("workspace");

ALTER TABLE "planning_run" ADD CONSTRAINT "fk_planning_run__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE;

CREATE TABLE "planning_recommendation" (
  "plnrec_id" SERIAL PRIMARY KEY,
  "planning_run" INTEGER NOT NULL,
  "device_task" INTEGER NOT NULL,
  "device" INTEGER,
  "operator" INTEGER,
  "plnrec_start" TIMESTAMP,
  "plnrec_end" TIMESTAMP,
  "plnrec_score" DOUBLE PRECISION NOT NULL DEFAULT 0,
  "plnrec_selected" BOOLEAN NOT NULL DEFAULT FALSE,
  "plnrec_warning_code" TEXT NOT NULL DEFAULT '',
  "plnrec_warning_text" TEXT NOT NULL DEFAULT '',
  "plnrec_explanation" TEXT NOT NULL DEFAULT ''
);

CREATE INDEX "idx_planning_recommendation__run" ON "planning_recommendation" ("planning_run");
CREATE INDEX "idx_planning_recommendation__task" ON "planning_recommendation" ("device_task");

ALTER TABLE "planning_recommendation" ADD CONSTRAINT "fk_planning_recommendation__run" FOREIGN KEY ("planning_run") REFERENCES "planning_run" ("plnrun_id") ON DELETE CASCADE;
ALTER TABLE "planning_recommendation" ADD CONSTRAINT "fk_planning_recommendation__task" FOREIGN KEY ("device_task") REFERENCES "device_task" ("dvctsk_id") ON DELETE CASCADE;
ALTER TABLE "planning_recommendation" ADD CONSTRAINT "fk_planning_recommendation__device" FOREIGN KEY ("device") REFERENCES "device" ("dvc_id") ON DELETE SET NULL;
ALTER TABLE "planning_recommendation" ADD CONSTRAINT "fk_planning_recommendation__operator" FOREIGN KEY ("operator") REFERENCES "operator" ("oprt_id") ON DELETE SET NULL;
