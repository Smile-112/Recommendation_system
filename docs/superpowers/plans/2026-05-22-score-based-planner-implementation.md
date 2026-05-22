# Score-Based Planner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first working backend implementation of the score-based planner with deadline-aware full-stack rescheduling, warnings, selected resources, and explainable recommendations.

**Architecture:** Keep the algorithm testable as pure service logic, then adapt storage/API around it. The first implementation persists the applied selected plan to `device_task` and stores recommendation history in new planning tables. UI polish and full settings screens remain later work.

**Tech Stack:** Go 1.24, pgx/PostgreSQL, chi HTTP API, Docker-based test execution because local `go.exe` is unavailable.

---

### Task 1: Pure Planner Model And Tests

**Files:**
- Create: `recsys-backend/internal/service/planner_engine.go`
- Create: `recsys-backend/internal/service/planner_engine_test.go`
- Modify: `recsys-backend/internal/service/planner.go`

- [ ] **Step 1: Write failing service tests**

Create tests for:

```go
func TestBuildPlanMovesFutureTaskToFitNewUrgentTask(t *testing.T)
func TestBuildPlanFallsBackWithDeadlineWarningWhenNoStrictPlanExists(t *testing.T)
func TestBuildPlanPrefersHigherDeviceCharacteristicScoreWhenDeadlineAllows(t *testing.T)
func TestBuildPlanUsesEarlierDeviceWhenPreferredDeviceMissesDeadline(t *testing.T)
func TestBuildPlanLeavesInProgressTaskFixed(t *testing.T)
```

Run:

```powershell
docker run --rm -v ${PWD}/recsys-backend:/src -w /src golang:1.24-alpine go test ./internal/service -run TestBuildPlan -count=1
```

Expected: fail because planner engine types/functions do not exist.

- [ ] **Step 2: Implement minimal pure planner engine**

Add internal service types:

```go
type planningTask struct { ... }
type planningDevice struct { ... }
type planningOperator struct { ... }
type planningOptions struct { Now time.Time; DayStartHour int; DayEndHour int }
type planAssignment struct { TaskID, DeviceID, OperatorID int64; Start, End time.Time; Score float64; WarningCode, Explanation string; Moved bool }
```

Add `buildScoreBasedPlan(input planningInput) planningOutput`.

- [ ] **Step 3: Verify service tests pass**

Run the same Docker test command. Expected: PASS.

### Task 2: Storage Schema For Criteria And Recommendations

**Files:**
- Modify: `recsys-backend/internal/storage/createDB.sql`
- Modify: `recsys-backend/internal/storage/schema.go`
- Modify: `recsys-backend/internal/storage/entities.go`
- Modify: `recsys-backend/internal/storage/repos.go`

- [ ] **Step 1: Add failing compile checks via storage tests**

Add storage-facing structs/method expectations in tests or compile references:

```go
type PlanningCriterion struct { ID int64; Code string; Name string }
type PlanningWeight struct { ID int64; WorkspaceID int64; CriterionCode string; Weight float64 }
type DeviceCharacteristicScore struct { ID int64; WorkspaceID int64; DeviceID int64; EquipmentCharacteristicID int64; Score float64 }
type PlanningRecommendation struct { ... }
```

Run:

```powershell
docker run --rm -v ${PWD}/recsys-backend:/src -w /src golang:1.24-alpine go test ./internal/storage -count=1
```

Expected: fail until methods/types exist.

- [ ] **Step 2: Add schema and repository methods**

Add `CREATE TABLE IF NOT EXISTS` support for:

- `planning_criteria`
- `planning_weights`
- `device_characteristic_score`
- `planning_run`
- `planning_recommendation`

Add idempotent schema ensure statements in `schema.go` for existing databases.

- [ ] **Step 3: Verify storage package compiles**

Run Docker storage test command. Expected: PASS.

### Task 3: Wire Planner To Storage And API

**Files:**
- Modify: `recsys-backend/internal/service/planner.go`
- Modify: `recsys-backend/internal/httpapi/handlers.go`
- Modify: `recsys-backend/internal/storage/repos.go`

- [ ] **Step 1: Write failing recompute behavior tests**

Add tests for:

```go
func TestRecomputeResultIncludesRecommendationsAndWarnings(t *testing.T)
func TestRecomputeResultIncludesSelectedDeviceAndOperator(t *testing.T)
```

Run:

```powershell
docker run --rm -v ${PWD}/recsys-backend:/src -w /src golang:1.24-alpine go test ./internal/service -run TestRecompute -count=1
```

Expected: fail until `RecomputeResult` includes explainable fields.

- [ ] **Step 2: Extend recompute result**

Add:

```go
type RecomputeWarning struct { TaskID int64; Code string; Message string }
type RecomputeRecommendation struct { TaskID, DeviceID, OperatorID int64; PlanStart, PlanEnd time.Time; Score float64; Selected bool; Explanation string }
```

Use pure engine output, persist selected plan with device/operator/start/end, save recommendations, and return explainable result.

- [ ] **Step 3: Verify service and API compile**

Run:

```powershell
docker run --rm -v ${PWD}/recsys-backend:/src -w /src golang:1.24-alpine go test ./...
```

Expected: PASS.

### Task 4: Documentation And UI Minimum

**Files:**
- Modify: `README.md`
- Modify: `recsys-backend/web/app.js`

- [ ] **Step 1: Add minimal UI handling**

Show recompute warnings in toasts and keep current Gantt rendering compatible.

- [ ] **Step 2: Update README algorithm section**

Document score-based recompute response and the deadline fallback rule.

- [ ] **Step 3: Run final verification**

Run:

```powershell
docker run --rm -v ${PWD}/recsys-backend:/src -w /src golang:1.24-alpine go test ./...
```

Expected: PASS.
