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
  "device_tasks_type" INTEGER NOT NULL,
  "workspace" INTEGER NOT NULL,
  "operator" INTEGER NOT NULL,
  "device" INTEGER NOT NULL,
  "priorities" INTEGER NOT NULL
);

CREATE INDEX "idx_device_task__device" ON "device_task" ("device");

CREATE INDEX "idx_device_task__device_tasks_type" ON "device_task" ("device_tasks_type");

CREATE INDEX "idx_device_task__operator" ON "device_task" ("operator");

CREATE INDEX "idx_device_task__priorities" ON "device_task" ("priorities");

CREATE INDEX "idx_device_task__workspace" ON "device_task" ("workspace");

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__device" FOREIGN KEY ("device") REFERENCES "device" ("dvc_id") ON DELETE CASCADE;

ALTER TABLE "device_task" ADD CONSTRAINT "fk_device_task__device_tasks_type" FOREIGN KEY ("device_tasks_type") REFERENCES "device_tasks_type" ("dvctsktp_id") ON DELETE CASCADE;

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
  "device_task" INTEGER NOT NULL,
  "operator" INTEGER NOT NULL
);

CREATE INDEX "idx_user_task__device_task" ON "user_task" ("device_task");

CREATE INDEX "idx_user_task__operator" ON "user_task" ("operator");

CREATE INDEX "idx_user_task__workspace" ON "user_task" ("workspace");

ALTER TABLE "user_task" ADD CONSTRAINT "fk_user_task__device_task" FOREIGN KEY ("device_task") REFERENCES "device_task" ("dvctsk_id") ON DELETE CASCADE;

ALTER TABLE "user_task" ADD CONSTRAINT "fk_user_task__operator" FOREIGN KEY ("operator") REFERENCES "operator" ("oprt_id") ON DELETE CASCADE;

ALTER TABLE "user_task" ADD CONSTRAINT "fk_user_task__workspace" FOREIGN KEY ("workspace") REFERENCES "workspace" ("wrkspc_id") ON DELETE CASCADE
