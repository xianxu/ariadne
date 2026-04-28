---
type: travel-plan
name: travel-plan
description: Use when the user wants to capture or plan a trip — destination, dates, travelers, itinerary, bookings, references. Triggers on "capture this trip", "plan a trip to X", "save the Rome trip", etc.
---

# travel-plan

A self-contained record of one trip: where, when, with whom, how it's structured day-to-day, what's booked, what's open. Lives long enough to be reused as a reference next time.

## Frontmatter shape

| Field | Required | Notes |
|---|---|---|
| `type` | yes | `travel-plan` |
| `destination` | yes | Free-form, can be a city, country, region, or multi-stop ("Rome → Florence → Venice"). |
| `start` | yes | ISO date of departure. |
| `end` | yes | ISO date of return. |
| `travelers` | yes | List of names (or "self" for solo). |
| `purpose` | no | One word: `leisure`, `family`, `work`, `conference`, etc. Useful for filtering across many plans. |
| `status` | yes | `planning`, `booked`, `in-progress`, `done`. Reflects current state. |

## Body skeleton

1. `# <destination> — <start> to <end>` — title.
2. **One-line summary** — who, what, why, in one sentence.
3. **Itinerary** — day-by-day or phase-by-phase. Each day a sub-heading with the date; bullets for the day's activities. Light is fine; this isn't a tour script.
4. **Bookings** — table or bullet list: flights, lodging, transit, reservations. Include confirmation numbers and links. Mark unbooked items explicitly.
5. **Logistics** — visas, vaccinations, currency, packing, pet/house care, work coverage. Skip the ones that don't apply. Don't pre-fill empty rows.
6. **Open questions** — undecided pieces. This is what "planning → booked" status transitions reduce.
7. **Links & references** — articles, recommendations, prior trips to the same place, contacts.

## Authoring instructions

When the dispatcher applies this prototype:

1. Pull as much as possible from conversation context first: **destination**, **dates**, **travelers**, any restaurants/sites/hotels mentioned. Use those to seed the file before asking.
2. If only a fuzzy time was given ("next month", "second week of June"), ask for concrete dates.
3. **Status** defaults to `planning` unless the conversation mentions confirmed bookings.
4. Itinerary: don't fabricate. If the user only said "Rome for a week," write the date range with empty day stubs and surface "needs day-by-day plan" under Open questions. Don't invent activities.
5. **Default location:** under a `memory/` directory if one exists. Prefer `memory/life/travel/` for personal trips and `memory/work/travel/` for work trips; differentiate by `purpose`. If neither subtree exists, run `find memory -type d` (or `find . -type d` if no memory dir) and propose 1–2 candidates. Always confirm location with the user when more than one plausible home exists.
6. Filename: `<start-date>-<destination-slug>.md` (e.g., `2026-06-14-rome.md`). For multi-destination, use the first stop.
