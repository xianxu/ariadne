---
id: 000015
status: blocked
created: 2026-04-29
updated: 2026-05-03
---

# product and roadmap data types

## Problem

Add two typed-document prototypes to the data system: `product` (the durable charter for a thing being built, spanning 1..N peer repos of the brain) and `roadmap` (a structured monthly snapshot tracking progress against a product's components). Co-design the prototypes and the workflows around them inside the prototype files themselves — the prototype is the spec.

"Product" is the umbrella term, deliberately preferred over "project" because "project" carries too many overloaded meanings (engineering effort, IDE workspace, the ariadne `workshop/` notion, etc.). Externally-sold products, internal efforts, and infra all fit the same charter shape under this name.
