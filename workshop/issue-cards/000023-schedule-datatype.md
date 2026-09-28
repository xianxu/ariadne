---
id: 000023
status: open
created: 2026-05-06
updated: 2026-05-06
estimate_hours: 1.5
---

# schedule datatype (singleton; tagged scheduled reminders)

## Problem

There seems to be several levels we can implement this, from just have a schedule that human running the harness can use, to schedule a service running certain things without agentic component during the run, to service calling agent to one-shot the task, to maybe using `pair` style wrapper on top of an agentic harness, so that we can do some rounds or at least interpret result to see if human attention is needed, all the way to building a full harness environment aware of scheduling as a core component. We need to figure out what's the best for investment/learning. 

New realization on 06/17/26, that we are essentially heading into the whole scheduling task execution world. the most simple form would be that:

1. there's some logic to be executed, in this case, quite similar to a manually invoked skill. 
2. we need to consider input and output. since we follow a repo based approach, then it stands to reason that input and output is just a function over the state of the repo itself, well, maybe also plus some external sources, such as time. 
3. and quickly you will have a DAG of tasks that would follow the completion, or failure of a scheduled job. the programming semantic of that is still hazy to me in the agentic era. but it seems likely such DAG should be managed by deterministic binary. 

the 2. (executable are just skills) + 3. (scheduling is a binary) + repo centric view, means the state of execution (e.g. the instantiation of the DAG at schedule) may be just in some file/directory structure inside the repo. at "personal" level, this is fine. if this became a bottleneck we will deal with that later by moving that portion of state into some database. or maybe we should use git's history to manage history, e.g. in repo we only keep track of basically state of latest run of a DAG. still hazy. 

the above view is tied to some periodical scheduling. another pattern would be trigger based, and periodical tick of time is just one trigger. maybe we should do that instead. 

at least on one level, those scheduled tasks are about maintaining some side effect. I feel some desire to separate such side effect from scheduling itself. meaning the consumer of such side effect would generally be oblivious of the schedule or trigger those artifact being constructed. otherwise, that consumer really is part of the scheduling system. 

This needs some brainstorming. to find what we actually want to implement at this point in time. The key use case of this, is really periodically sync gmail, for example, and build insights/memory, so that user can just ask about them without triggering the external fetches. essentially, the centralization of personal data. 
