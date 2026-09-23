Resolve an existing workspace from the current repository context (read-only).

  sdlc workspace --json          current checkout, including ordinary worktrees
  sdlc workspace :0 --json       primary checkout
  sdlc workspace pair:1 --json   registered durable slot 1 in pair

A repo name aliases repo:0; :N is contextual. Qualified repository names match
exactly. Slots occupy <fleet>/worktree/<repo>-slotN and rest on main-slotN;
identity survives checking out an issue branch. Git membership is verified.
No directory, branch, claim or repository state is created or changed.

JSON v1 fields: schema_version, repo, repo_identity (Git common directory),
primary_root, fleet_root, worktree_root, kind, address, slot, branch, head,
resting_branch. Ordinary worktrees have null address, slot and resting_branch;
detached checkouts have null branch; an unborn primary has null head.
Errors exit nonzero without a partial success object. Results are observations,
not reservations: commands that later mutate a workspace must revalidate it.
