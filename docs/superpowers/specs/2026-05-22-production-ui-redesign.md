# Production UI Redesign

## Goal

Rebuild the current interface into a production control cockpit for task planning, equipment loading, operator availability, and recommendation review.

The UI must make the score-based planner understandable and usable in daily production work. Users should see what is planned, why it was planned that way, which resources are busy, and where deadlines or resource constraints create risk.

## Design Direction

Use a full application redesign rather than small cosmetic fixes.

The first screen should be the actual working cockpit, not a landing page. The interface should feel like a quiet operational tool for repeated use:

- Dense but readable production information.
- Clear resource timelines.
- Fast actions for recomputing the plan and adding production tasks.
- Warnings and recommendation explanations visible near the plan.
- Reduced decorative copy and fewer marketing-style blocks.

## Main Navigation

Replace the current top-heavy header with a persistent application shell:

- Left sidebar:
  - Brand.
  - Current workspace.
  - Navigation items.
  - API/user status.
- Main content area:
  - Active work screen.
- Right-side contextual panel on planning-heavy screens:
  - Current recommendation.
  - Warnings.
  - Resource notes.

Navigation sections:

- Production plan.
- Tasks.
- Equipment.
- Operators.
- Schedule.
- References.
- Profile/admin when available.

The navigation should be stable and compact so the user does not need to scroll past header controls before working.

## Production Plan Screen

The production plan screen becomes the default working screen.

It should contain:

- Operational summary:
  - tasks in plan;
  - deadline risks;
  - unscheduled tasks;
  - equipment/operator utilization;
  - last planning run.
- Primary timeline:
  - rows grouped by equipment by default;
  - optional operator view where feasible;
  - current-time marker;
  - visible workday hours;
  - status colors for completed, in-progress, pending, warning, and blocked items.
- Recommendation panel:
  - selected task;
  - selected equipment and operator;
  - score;
  - warning code/text;
  - explanation.
- Action bar:
  - recompute plan;
  - create task;
  - refresh;
  - date selector.

The timeline is the center of the product. Cards and summaries support it instead of competing with it.

## Task UX

Tasks should be created as production intents.

The main task form should emphasize:

- name;
- deadline;
- duration;
- setup time;
- unload time;
- production characteristic;
- manual priority;
- add-to-recommendation flag.

Equipment and operator assignment should be optional and visually secondary because the planner is expected to recommend them.

Task rows/cards should show:

- deadline status;
- planned start/end;
- assigned/recommended equipment;
- assigned/recommended operator;
- characteristic;
- priority;
- warning badge when the planner cannot meet the deadline.

Editing an already planned task should make it clear which fields are source inputs and which fields are planner outputs.

## Planner UX

After recomputation, the UI should not only show a toast.

It should expose:

- number of updated tasks;
- warnings;
- unscheduled IDs;
- selected recommendations;
- moved tasks when backend data is available;
- explanation text for selected recommendations.

The UI should preserve the most recent recompute result in client state so the user can inspect it after the toast disappears.

Warning messages should be direct and operational, for example:

- "Невозможно выполнить до дедлайна без нарушения сроков других задач."
- "Нет совместимого оборудования для этой характеристики."
- "Нет доступного оператора с нужной компетенцией."

## Operator UX

Operators are production resources, not only contact records.

Operator cards should show:

- name and phone;
- competencies;
- assigned devices;
- current planned load;
- active breaks or unavailable intervals;
- next assigned task.

The operator screen should help answer:

- who can run this equipment type;
- who is already busy;
- who is preferred for a concrete device;
- which operators have no competencies configured.

## Schedule UX

The schedule screen should merge operator work intervals with production workload.

It should show:

- operator rows;
- planned task bars;
- break/unavailable bars;
- conflict hints when a task overlaps an unavailable interval;
- date selector;
- add/edit break action.

The schedule should use the same visual language as the production plan timeline so users do not have to learn two different chart systems.

## Equipment UX

Equipment cards should show:

- device name;
- type;
- production characteristic;
- recommendation participation flag;
- utilization summary;
- currently assigned task;
- next planned task;
- configured characteristic score where available.

Equipment should make it obvious whether a device can participate in planning and why it is preferred for a characteristic.

## References UX

References remain available but should be organized as planning settings:

- characteristics;
- equipment types;
- task types;
- planning criteria weights;
- device-characteristic scores.

The screen should use compact editable tables rather than large explanatory cards.

## Visual System

Use a restrained operational style:

- dark sidebar;
- light work area;
- compact cards with 8px radius or less;
- clear borders;
- neutral background;
- status colors only where they carry meaning;
- no oversized hero blocks;
- no decorative gradients or blobs;
- dense tables and timelines with readable spacing.

Primary status colors:

- pending: blue;
- in progress: green;
- completed: neutral gray;
- warning: orange/red;
- break/unavailable: muted slate.

The color system should not be one-note. Avoid making the interface dominated by one hue family.

## Data Flow

On workspace load:

1. Load workspace reference data.
2. Load tasks, devices, operators, competencies, operator-device assignments, user tasks, planning weights, and device scores.
3. Build derived UI state:
   - task schedule rows;
   - equipment utilization;
   - operator workload;
   - deadline risk counts;
   - missing configuration warnings.

On recompute:

1. Call `POST /api/plans/recompute`.
2. Store the result in client state.
3. Refresh workspace data.
4. Render updated plan, recommendation panel, and warning list.

## Error Handling

The UI should distinguish:

- API unavailable;
- validation errors;
- planner warnings;
- resource configuration gaps;
- empty workspace state.

Empty states should guide the user to the next action without long explanatory paragraphs.

## Testing And Verification

Manual UI verification:

- open app at `http://localhost:8080`;
- check desktop layout;
- check narrow viewport layout;
- create a task without manual device/operator;
- recompute plan;
- inspect recommendation panel;
- add operator unavailable interval;
- verify schedule timeline updates.

Automated or lightweight checks:

- `go test ./...`;
- browser smoke test for main pages;
- visual screenshot check for major overlap/blank layout issues.

## Future Refinement

The design must remain easy to adjust after review.

Expected follow-up refinements:

- sidebar density;
- color palette;
- task card details;
- timeline row height;
- recommendation panel wording;
- separate equipment/operator timeline modes;
- additional filters for large production queues.

The implementation should keep these concerns separated in CSS and rendering functions so visual refinements do not require rewriting planner logic.
