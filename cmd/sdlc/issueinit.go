// issueinit.go is the explicit, empty-repository setup action. Composition
// never calls it: publication requires the user's deliberate setup command.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func newIssueInitCmd() *cobra.Command {
	return markMutatingCommand(&cobra.Command{
		Use: "init", Short: "Initialize an empty new repository's issue tracker",
		Long: "Finish new-repository setup after committing and pushing main. Initializes only an empty issue inventory; existing issues require issue migrate. Brains and missing/declined issue scaffolds are skipped. A retry retains confirmed publication and completes a clean main checkout without resetting work.",
		Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueInit(commandContext(cmd.Context()), cmd.ErrOrStderr())
		},
	})
}

// initializationEffects is the controllable effect boundary. The dispatcher
// executes only effects returned by the opaque tracker initialization model.
type initializationEffects interface {
	Inspect(context.Context) (tracker.InitInspected, error)
	Bootstrap(context.Context) (string, error)
	Publish(context.Context, string) (string, error)
	FastForward(context.Context, string) error
}

func runIssueInit(ctx context.Context, stderr io.Writer) error {
	state, err := runInitialization(ctx, &gitInitializationEffects{})
	if err != nil {
		return fmt.Errorf("issue init incomplete; retain published state and rerun sdlc issue init after fixing the cause: %w", err)
	}
	if state.Skipped() {
		cinfo(stderr, "issue tracker initialization skipped: brain or missing/declined issue scaffold")
		return nil
	}
	cok(stderr, "issue tracker initialized; main checkout contains the verified cutover")
	return nil
}

func runInitialization(ctx context.Context, effects initializationEffects) (tracker.InitState, error) {
	state := tracker.NewInitialization()
	var event tracker.InitEvent = tracker.BeginInit{}
	var failure error
	for {
		next, effect, err := tracker.InitTransition(state, event)
		if err != nil {
			return state, err
		}
		state = next
		if effect == nil {
			if failure != nil {
				return state, failure
			}
			if state.Failure() != "" {
				return state, errors.New(state.Failure())
			}
			return state, nil
		}
		failure = ctx.Err()
		if failure == nil {
			switch effect := effect.(type) {
			case tracker.InspectInitialization:
				event, failure = effects.Inspect(ctx)
			case tracker.BootstrapTracker:
				var root string
				root, failure = effects.Bootstrap(ctx)
				event = tracker.BootstrapConfirmed{Root: root}
			case tracker.PublishCutover:
				var commit string
				commit, failure = effects.Publish(ctx, effect.Root)
				event = tracker.CutoverConfirmed{Commit: commit}
			case tracker.FastForwardCheckout:
				failure = effects.FastForward(ctx, effect.Commit)
				event = tracker.CheckoutConfirmed{}
			default:
				return state, fmt.Errorf("unsupported initialization effect %T", effect)
			}
		}
		if failure != nil {
			event = tracker.InitFailed{Reason: failure.Error()}
		}
	}
}

type gitInitializationEffects struct {
	env  *migrateEnv
	plan tracker.MigrationManifest
}

func (g *gitInitializationEffects) Inspect(ctx context.Context) (tracker.InitInspected, error) {
	refuse := func(err error) (tracker.InitInspected, error) { return tracker.InitInspected{}, err }
	raw, stderr, err := gitx.RunGitInContext(ctx, ".", nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return refuse(fmt.Errorf("run setup from a committed Git repository: %w: %s", err, strings.TrimSpace(string(stderr))))
	}
	root := strings.TrimSpace(string(raw))
	if gitx.IsBrainRepo(root) {
		return tracker.InitInspected{Disposition: tracker.InitSkip, Reason: "brain repository"}, nil
	}
	// A failed marker probe is not proof that this is a non-brain repository.
	if _, err := os.Stat(filepath.Join(root, ".brain", "config.md")); err != nil && !os.IsNotExist(err) {
		return refuse(err)
	}
	info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(vocab.Issue().Discovery().Home)))
	if os.IsNotExist(err) {
		return tracker.InitInspected{Disposition: tracker.InitSkip, Reason: "issue scaffold absent or declined"}, nil
	}
	if err != nil {
		return refuse(err)
	}
	if !info.IsDir() {
		return refuse(errors.New("issue init needs the repository's own issue directory, not a symlink or file"))
	}
	env, err := openMigrate(ctx)
	if err != nil {
		return refuse(fmt.Errorf("commit initial setup and configure/push main's publication upstream first: %w", err))
	}
	g.env = env
	main, err := env.main.Snapshot()
	if err != nil {
		return refuse(fmt.Errorf("push initial main to its configured publication remote first: %w", err))
	}
	if ancestor, err := env.ancestorOf(env.head, main.Ref()); err != nil || !ancestor {
		if err == nil {
			err = errors.New("local main is ahead of or diverged from published main; push/reconcile it before issue init")
		}
		return refuse(err)
	}
	trackerRoot, complete, err := migratedAlready(env, main)
	if err != nil {
		return refuse(err)
	}
	if complete {
		if err = g.cleanMain(); err != nil {
			return refuse(err)
		}
		commit, err := initializationCutoverCommit(env, main.Ref(), trackerRoot)
		if err != nil {
			return refuse(err)
		}
		localRoot, present, err := tracker.ReadCutoverMarkerAt(env.root, env.head)
		if err != nil {
			return refuse(err)
		}
		if present {
			if localRoot != trackerRoot {
				return refuse(errors.New("local cutover names a different tracker root; reconcile it without recreating the tracker"))
			}
			return tracker.InitInspected{Disposition: tracker.InitCurrent, Root: trackerRoot, Commit: commit}, nil
		}
		return tracker.InitInspected{Disposition: tracker.InitRemoteReady, Root: trackerRoot, Commit: commit}, nil
	}
	if err := verifyInitializationEmpty(env, main.Ref()); err != nil {
		return refuse(err)
	}
	if err = g.cleanMain(); err != nil {
		return refuse(err)
	}
	// Presence was verified across every root/scope. Construct only an empty
	// bootstrap manifest; init never invokes migration's importer or diff rules.
	in := tracker.MigrationInput{Repository: env.target.Repository, ObjectFormat: env.format, Main: main.Ref()}
	g.plan = tracker.PlanTrackerMigration(in)
	if !tracker.EmptyInitEligible(in, g.plan) {
		return refuse(errors.New("existing or uncertain legacy issue inventory; use sdlc issue migrate for reviewed migration"))
	}
	exists, err := env.tracker.Initialized()
	if err != nil {
		return refuse(err)
	}
	if exists {
		root, err := trackerBootstrapRoot(env, g.plan)
		if err != nil {
			return refuse(fmt.Errorf("%w: existing branch is not a verified empty init bootstrap: %w", errInitTrackerAlreadyInitialized, err))
		}
		view, err := env.trunk.Snapshot()
		if err != nil {
			return refuse(err)
		}
		if err = sameTrackerFiles(env, view.Ref(), g.plan); err != nil {
			return refuse(fmt.Errorf("%w: existing tracker is not a verified empty init bootstrap: %w", errInitTrackerAlreadyInitialized, err))
		}
		return tracker.InitInspected{Disposition: tracker.InitTrackerOnly, Root: root}, nil
	}
	return tracker.InitInspected{Disposition: tracker.InitEmpty}, nil
}

func (g *gitInitializationEffects) cleanMain() error {
	env := g.env
	branch, err := env.git("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || branch != "main" || env.resting != "main" {
		return errors.New("issue init requires the clean main checkout; return to main before retrying")
	}
	entries, err := env.statusEntries(env.root, "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return errors.New("commit the initial setup or preserve uncommitted changes before issue init; main must be clean")
	}
	return nil
}

func (g *gitInitializationEffects) Bootstrap(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return applyTrackerBootstrap(g.env, g.plan)
}

func (g *gitInitializationEffects) Publish(ctx context.Context, root string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := applyMainCutover(g.env, g.plan, root); err != nil {
		return "", err
	}
	main, err := g.env.main.Snapshot()
	if err != nil {
		return "", err
	}
	actual, complete, err := migratedAlready(g.env, main)
	if err != nil {
		return "", err
	}
	if !complete || actual != root {
		return "", errors.New("published cutover identity could not be confirmed")
	}
	return initializationCutoverCommit(g.env, main.Ref(), root)
}

func initializationCutoverCommit(env *migrateEnv, main, root string) (string, error) {
	commit, err := migrationCommit(env, main)
	if err != nil {
		return "", err
	}
	actual, present, err := tracker.ReadCutoverMarkerAt(env.root, commit)
	if err != nil {
		return "", err
	}
	if !present || actual != root {
		return "", errors.New("cutover introduction does not name the verified tracker root")
	}
	return commit, nil
}

func (g *gitInitializationEffects) FastForward(ctx context.Context, commit string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := g.cleanMain(); err != nil {
		return err
	}
	if ancestor, err := g.env.ancestorOf("HEAD", commit); err != nil || !ancestor {
		if err != nil {
			return err
		}
		return errors.New("main advanced beyond or diverged from the pinned cutover; retain it and rerun issue init")
	}
	_, err := g.env.git("-c", "submodule.recurse=false", "merge", "--ff-only", "--no-autostash", "--no-overwrite-ignore", commit)
	return err
}
