---
id: 000023
status: open
deps: []
created: 2026-05-06
updated: 2026-05-06
estimate_hours: 1.5
---

# schedule datatype (singleton; tagged scheduled reminders)

There seems to be several levels we can implement this, from just have a schedule that human running the harness can use, to schedule a service running certain things without agentic component during the run, to service calling agent to one-shot the task, to maybe using `pair` style wrapper on top of an agentic harness, so that we can do some rounds or at least interpret result to see if human attention is needed, all the way to building a full harness environment aware of scheduling as a core component. We need to figure out what's the best for investment/learning. 

New realization on 06/17/26, that we are essentially heading into the whole scheduling task execution world. the most simple form would be that:

1. there's some logic to be executed, in this case, quite similar to a manually invoked skill. 
2. we need to consider input and output. since we follow a repo based approach, then it stands to reason that input and output is just a function over the state of the repo itself, well, maybe also plus some external sources, such as time. 
3. and quickly you will have a DAG of tasks that would follow the completion, or failure of a scheduled job. the programming semantic of that is still hazy to me in the agentic era. but it seems likely such DAG should be managed by deterministic binary. 

the 2. (executable are just skills) + 3. (scheduling is a binary) + repo centric view, means the state of execution (e.g. the instantiation of the DAG at schedule) may be just in some file/directory structure inside the repo. at "personal" level, this is fine. if this became a bottleneck we will deal with that later by moving that portion of state into some database. or maybe we should use git's history to manage history, e.g. in repo we only keep track of basically state of latest run of a DAG. still hazy. 

the above view is tied to some periodical scheduling. another pattern would be trigger based, and periodical tick of time is just one trigger. maybe we should do that instead. 

at least on one level, those scheduled tasks are about maintaining some side effect. I feel some desire to separate such side effect from scheduling itself. meaning the consumer of such side effect would generally be oblivious of the schedule or trigger those artifact being constructed. otherwise, that consumer really is part of the scheduling system. 

This needs some brainstorming. to find what we actually want to implement at this point in time. The key use case of this, is really periodically sync gmail, for example, and build insights/memory, so that user can just ask about them without triggering the external fetches. essentially, the centralization of personal data. 

## Done when

- A new `schedule` typed-data prototype lives in `construct/datatype/schedule.md`, defining frontmatter + body skeleton + authoring instructions.
- The `xx-datatype` skill recognizes `schedule` as a valid type, can author a new `schedule.md` (singleton per brain), and can append entries to an existing one.
- `data/schedule.md` (or similar canonical path) exists in the user's brain holding scheduled items with tags.
- An accompanying tool or skill checks `data/schedule.md` periodically (or on demand) and surfaces upcoming items, filtered by tag if specified.
- Atlas entry under `data-artifacts.md` (or similar) documents the convention.

## Spec

The motivating use case: while working on the shared-brain project, we identified end-of-project tasks that need a "remind me about this in N weeks" mechanism — e.g., "delete `brain-backup` GitHub repo 1 month after the gcrypt cutover ships." Today there's no datatype for that. We could shoehorn it into project tasks, but project files are execution containers for active work, not for "do this thing later" reminders.

The user proposed: **schedule is a singleton datatype** — one `data/schedule.md` per brain, with multiple entries tagged for filtering (e.g., `work`, `personal`, `urgent`). Mirrors the shape of `roadmap` (also singleton, multi-entry).

### Sketched shape (to iterate on during M1)

Frontmatter:
```
type: schedule
name: schedule          # singleton — name is just a label
created: <ISO>
updated: <ISO>
```

Body:
- Section per time horizon (today / this week / this month / future) OR sorted single list with `due:` per item — pick during design.
- Each entry: title, due-date or recurrence rule, tags (`[work, urgent]`), optional action description, optional `done:` marker.

Open design questions for M1:
- Sorted single list vs sectioned-by-horizon — both have UX trade-offs; pick based on how the entries get authored and consumed.
- Recurrence: support cron-style? Just "every N days/weeks/months"? Or skip recurrence entirely (one-shot only) and let the user re-author after triggering?
- Notification mechanism: should this issue ship a notifier, or is reading the file on demand enough? Probably the latter for v1; notifier as a follow-on once the file format is stable.
- Cross-brain schedule references: a shared brain could have its own schedule (e.g., `brain-shared-family/data/schedule.md` for joint reminders). The datatype should support both private and shared placement.

### Out of scope for this issue

- A daemon that proactively notifies (push notifications, email, etc.). v1 is "reading the file surfaces what's upcoming." Daemon can come later.
- Calendar integration (iCal export, Google Calendar sync). Possible follow-on.
- Recurrence semantics beyond simple "every N days" (cron, RRULE). Skip if we don't have a use case for v1.

## Plan

### M1 — design + author the prototype

- [ ] Decide single-list vs sectioned body shape based on use-case sketches (start with: shared-brain end-of-project reminders, brain#10 follow-on tasks, anything else the user surfaces).
- [ ] Decide recurrence model (probably: optional `recurrence: every Nd | every Nw | every Nm`; default one-shot).
- [ ] Author `construct/datatype/schedule.md` with the chosen shape.
- [ ] Update `xx-datatype` skill recognition table.
- [ ] Atlas entry pointer.

### M2 — first user

- [ ] Author `data/schedule.md` in the user's brain, seeded with the immediate use cases (e.g., the 1-month brain-backup cleanup deadline from `nous#3` cutover).
- [ ] Verify the datatype is consumable: can the user (or an agent) `rg` for upcoming items, filter by tag, mark done?

### M3 — surfacing tool

- [ ] A small skill or script (e.g., `/schedule check` or `make schedule`) that reads the file and prints due-now / upcoming items, filtered by tag.
- [ ] Decide whether to wire into a startup-of-session hint (parley, claude-code session start) or leave as on-demand. Probably on-demand for v1.

## Log

### 2026-05-06

- Issue spec'd. Carved out from the shared-brain project's `nous#3` M3 cutover work — we identified a "remind me to delete the backup in 1 month" need with no good home in the existing datatype set. Filing for later, not as part of shared-brain MVP.
