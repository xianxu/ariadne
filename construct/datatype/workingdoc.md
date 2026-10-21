---
type: type
collaboration-mode: discussion
name: workingdoc
description: Use when maintaining a living, co-edited page on one topic whose agreed content can be settled into a tracked copy and updated as discussion continues. Distinct from a pensive's moment of thought and a project's execution record.
---

# workingdoc

A workingdoc is the ongoing editing surface for one topic. The operator and its owner develop the page together; settling preserves the agreed content in a tracked copy without ending the live discussion. Git keeps the settled copy's history, and a source pointer connects the live page to its last published edition.

## Frontmatter shape

| Field | Required | Notes |
|---|---|---|
| `type` | yes | `workingdoc` |
| `collaboration-mode` | yes | `discussion`: the default review collaboration mode. |
| `owner` | yes | Person or team responsible for the topic. Business ownership does not route Couch submissions. |
| `topic` | yes | Short phrase naming the page's single topic. |
| `parent` | for a side doc | Relative path to the parent document, following the shared side-doc contract. |
| `settled-into` | no | Tool-managed plain scalar: `<repo-relative-path> @ <full-commit>`. Written in the live source only after the exact copy is confirmed published. |

## Body skeleton

Start with `# <Topic>` and coherent prose that stands on its own. Use the topic's natural headings. Include Decisions and Open questions only when they have content; a workingdoc has no mandatory ledger or empty sections. A side doc retains its parent backlink.

## Authoring instructions

1. Infer the topic and existing substance from the conversation. Ask for a missing owner rather than inventing one. Use an existing workingdoc when the topic already has one.
2. A session side document lives in its parent's `.side/` folder and follows the shared side-doc naming, backlink and discussions-index rules. Other workingdocs follow the repository's established document layout.
3. Preserve the operator's unresolved review items. Answer questions through the review protocol; never silently discard commentary or accept proposals while making a copy. The operator resolves proposals with the editor's acceptance gesture.
4. When explicitly asked to settle, use `tl-lead doc settle SIDE-DOC [--to PATH]`. It refuses unresolved commentary, unaccepted proposals and malformed markers. It publishes the selected copy, then stamps the source through Couch's guarded write. The default copy is `workshop/workingdoc/YYYY-MM-DD-<slug>.md`; a later settle updates that same target. This explicit command owns its commit/publication; ordinary capture and edits retain the dispatcher's normal review discipline.
5. Keep editing the live source after settling. Use `tl-lead doc status` to see pages that changed since the last settled edition. A project links the tracked copy as background rather than relying on an ignored side-doc path. Reconcile an independently edited copy before re-settling; never overwrite it by hand to evade a refusal.
6. Preserve the settlement pointer during ordinary edits. Do not type or update the commit from memory, add a duplicate pointer, or copy that pointer into its own target. A publication/stamp interruption retains the known copy and reports the retry action.

## Search recipes

```sh
rg -l '^type: workingdoc' session.side/ workshop/workingdoc/
rg '^topic:|^owner:|^settled-into:' session.side/ workshop/workingdoc/
rg -l '^type: workingdoc' workshop/workingdoc/ | xargs rg -n '^## Decisions|^## Open questions'
```

## Rules

- One topic per page. Existing side-doc pages can be settled without bulk conversion.
- `settled-into` records a confirmed Git edition, not an intention or an automatic live mirror.
- Unknown frontmatter belongs to the document and survives settlement edits.
- Removing or archiving the tracked copy is an explicit repository edit; status reports its absence rather than silently recreating it.
