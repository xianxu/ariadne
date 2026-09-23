Resolve an existing workspace from the current repository context (read-only).

  sdlc workspace --json          current checkout, including ordinary worktrees
  sdlc workspace :0 --json       primary checkout
  sdlc workspace pair:1 --json   registered durable slot 1 in pair

A repo name aliases repo:0; :N is contextual. Qualified repository names match
exactly. Slots occupy <fleet>/worktree/<repo>-slotN/<repo> and rest on main-slotN;
identity survives checking out an issue branch. Git membership is verified.
No directory, branch, claim or repository state is created or changed.

JSON v2 fields: schema_version, repo, repo_identity (Git common directory),
primary_root, fleet_root, environment_root, environment_host (when numbered),
worktree_root, kind, address, slot, branch, head,
resting_branch. Ordinary worktrees have null address, slot and resting_branch;
detached checkouts have null branch; an unborn primary has null head.
Errors exit nonzero without a partial success object. Results are observations,
not reservations: commands that later mutate a workspace must revalidate it.

Numbered environments contain one linked host worktree and independent ordinary
source clones. Dependencies have kind dependency and null address, slot and
resting_branch; their feature worktrees have kind worktree. environment_host
contains repo, slot, repo_identity, primary_root and worktree_root. Dependencies
retain their own Git identity while fleet_root names the canonical host fleet.
Qualified workspace addresses always select that fleet, even for a dependency's
own name. A dependency caller must name the repository instead of using :N.
Legacy flat slot-looking worktrees remain ordinary worktrees.

Consumers must reject unsupported schema versions or malformed/truncated JSON
and rerun this command. Archived v1 output is stale evidence, never mutation authority.
