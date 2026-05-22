# Score-Based Planner Design

## Goal

Replace the current earliest-slot planner with an explainable score-based recommendation algorithm for production tasks.

The planner must rebuild the queue from the full unfinished task stack whenever the plan is recomputed, select suitable equipment and operators automatically, respect deadlines where possible, and explain why a recommendation was produced.

## Current Context

The project already has the main production entities:

- `device_task`: production task with duration, setup time, unload time, deadline, priority, selected device/operator, and planned start/end.
- `eqpmnt_characteristics`: workspace-specific equipment/task characteristics.
- `devices_type`: equipment types linked to a characteristic.
- `device`: concrete equipment linked to a device type.
- `competencies_operator`: operator competency by equipment type.
- `operator_device`: optional operator-to-device assignment.
- `user_task`: operator busy intervals, breaks, and shift-like tasks.

The existing planner sorts tasks by deadline and priority, then finds the earliest slot for the task's already selected device/operator. This is too narrow for the target recommendation workflow because it does not score alternatives, does not choose the best equipment, does not explain decisions, and does not rebuild the whole future queue.

## Core Behavior

When a new task is added or the plan is recomputed, the planner processes the full stack of unfinished tasks that participate in recommendations.

Tasks are split into two groups:

- Fixed tasks: already completed or currently in progress. Their intervals cannot move and become busy intervals.
- Movable tasks: not yet started. The planner may move them, but only if each moved task remains within its own deadline.

The planner first attempts to build a plan where the new task fits before its deadline and no movable task misses its deadline. If that is impossible, the new task is placed in the nearest realistic slot and receives a warning that the deadline cannot be met without violating constraints.

## Task Inputs

A task should be treated as an intent, not as a manually fixed device/operator assignment.

Primary task inputs:

- Name.
- Deadline.
- Duration, setup time, and unload time.
- Characteristic or work requirement from workspace dictionaries.
- Optional manual priority.
- Add-to-recommendation flag.

Examples of characteristics:

- Multicolor printing.
- Wood-compatible processing.
- Ventilation-required work.
- Material or processing mode.

The model should stay generic. "Multicolor printing" is one possible characteristic, not a hard-coded concept.

## Candidate Construction

For each task, the planner builds candidate resource combinations:

1. Determine required characteristic from the task.
2. Find device types that satisfy that characteristic.
3. Find active devices of those types that participate in recommendations.
4. Find operators that can work with the selected device type through competencies.
5. Optionally prefer operators assigned to the selected concrete device through `operator_device`.
6. Build feasible time intervals from equipment availability, operator availability, existing plan, and `user_task` busy intervals.

Operator requirements are derived from equipment and competencies. Users do not manually specify operator requirements on the task.

## Scoring Model

Each candidate is a tuple:

```text
task + device + operator + interval
```

The candidate receives an explanation-friendly score composed from weighted criteria.

Initial criteria:

- Deadline urgency.
- Optional manual priority.
- Equipment-characteristic suitability.
- Equipment efficiency or profitability for the characteristic.
- Earlier feasible completion time.
- Penalty for lateness.
- Penalty for moving existing tasks.

The formula should be configurable by workspace weights rather than hard-coded as one opaque expression.

Suggested storage:

- `planning_criteria`: dictionary of supported criteria.
- `planning_weights`: workspace-specific criterion weights.
- `device_characteristic_score`: score or preference for a concrete device and characteristic.

This allows one printer to be preferred for multicolor tasks, while still letting an earlier slot win when the deadline is at risk.

## Rescheduling Rules

The planner may reschedule movable tasks only when all of these are true:

- The task has not started yet.
- The task is not completed.
- The new interval does not exceed that task's deadline.
- The move does not conflict with fixed equipment or operator intervals.

The planner must not satisfy a new task by causing another task to miss its own deadline.

If no no-lateness plan exists for the new task, the planner switches to real-plan mode:

- Keep fixed constraints.
- Find the closest feasible execution slot.
- Preserve warning metadata that the task misses its deadline.
- Return or store the nearest realistic completion time.

## Recommendation Output

The planner should save and return more than `plan_start` and `plan_end`.

Recommended output:

- Planned start and end.
- Selected device.
- Selected operator when required.
- Total score.
- Warning code and warning text when applicable.
- Human-readable explanation.
- Alternative candidates where useful.
- List of tasks that were moved.

Suggested storage:

- `planning_run`: one recomputation run, workspace, timestamp, status.
- `planning_recommendation`: result per task and candidate, including selected flag, score, warning, explanation, device, operator, start, and end.

The existing `device_task.plan_start`, `device_task.plan_end`, `device`, and `operator` can still store the applied selected plan for compatibility with the current UI.

## Algorithm Outline

1. Load workspace tasks that are unfinished and enabled for recommendations.
2. Classify fixed and movable tasks using current time, completion mark, and plan boundaries.
3. Load devices, device types, characteristics, operators, competencies, operator-device assignments, user tasks, planning weights, and device-characteristic scores.
4. Create busy calendars from fixed tasks and user tasks.
5. Build candidate resource combinations for each movable task.
6. Try strict scheduling:
   - Place tasks into a queue ordered by deadline risk, manual priority, and weighted task score.
   - Search feasible slots before each task's deadline.
   - Allow moving future tasks only if their deadlines remain valid.
7. If strict scheduling fails for the new task, run fallback scheduling:
   - Find the closest realistic slot even if it finishes after the new task's deadline.
   - Mark the recommendation with a deadline warning.
8. Score selected and alternative candidates.
9. Persist the selected plan and recommendation explanation.
10. Return counts, warnings, selected plans, alternatives, and unscheduled tasks if any resource requirement cannot be satisfied at all.

## API Impact

Extend `POST /api/plans/recompute` response from a simple updated count to an explainable result.

Suggested response shape:

```json
{
  "run_id": 42,
  "updated": 5,
  "warnings": [
    {
      "task_id": 12,
      "code": "deadline_missed",
      "message": "Невозможно выполнить до дедлайна без нарушения сроков других задач."
    }
  ],
  "recommendations": [
    {
      "task_id": 12,
      "device_id": 3,
      "operator_id": 7,
      "plan_start": "2026-05-23T09:00:00Z",
      "plan_end": "2026-05-23T14:30:00Z",
      "score": 87.5,
      "selected": true,
      "explanation": "Выбран ближайший реальный слот; дедлайн недостижим без просрочки других задач."
    }
  ],
  "unscheduled_ids": []
}
```

## UI Impact

The UI should expose the algorithm rather than hiding it:

- Recommendation warning badge on tasks that miss deadline.
- Explanation panel for a selected recommendation.
- Criteria/weights editor in references or planning settings.
- Device-characteristic score controls near equipment or characteristics.
- Alternative recommendations where a better machine lost to an earlier slot.

## Testing Strategy

Add tests at the service layer before changing planner behavior:

- Fixed in-progress task is never moved.
- Movable task can move only when its own deadline remains satisfied.
- New task is scheduled before deadline when a valid reshuffle exists.
- New task receives a deadline warning and nearest realistic slot when no valid reshuffle exists.
- Operator is derived from selected equipment type competency.
- Device with higher characteristic score wins when deadlines are equal.
- Earlier slot wins over preferred device when otherwise the deadline would be missed.
- No compatible equipment yields an unscheduled recommendation.

## Non-Goals

- Full optimization across every possible permutation for very large production queues.
- Hard-coding 3D-printing-specific fields such as multicolor into the algorithm.
- Replacing the whole UI before the algorithm and data model are stable.
