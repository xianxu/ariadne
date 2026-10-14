package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// exitError carries a process exit code: 3 for the operator's lock and 6 for
// a section still stale after the retries, the codes couch uses for the same
// refusals.
type exitError struct {
	code int
	err  error
}

func (e exitError) Error() string { return e.err.Error() }

// writeTries bounds the read-recompute-write loop against a section that keeps
// changing under the TL.
const writeTries = 4

// writeSection is the TL's one write path: read the section and its hash,
// compute the new body from what was read, write it naming that hash, and on
// stale read and recompute again. A held lock refuses at once (no retry: the
// update waits for the operator's submit). With foldOwed, the write also folds
// the owed events the TL has already seen, and drops them from the seen-set.
func writeSection(ss SessionSync, out io.Writer, file, heading string, compute func(current string) (string, error), foldOwed bool) error {
	for range writeTries {
		current, hash, err := ss.ReadSection(file, heading)
		if err != nil {
			return refusal(file, heading, err)
		}
		body, err := compute(current)
		if err != nil {
			return err
		}
		var fold, seen []string
		if foldOwed {
			st, err := ss.Status(file)
			if err != nil {
				return refusal(file, heading, err)
			}
			if seen, err = loadSeen(file); err != nil {
				return err
			}
			fold = foldable(st.Owed, seen)
		}
		err = writeBody(body, func(path string) error { return ss.WriteSection(file, heading, path, hash, fold) })
		if errors.Is(err, ErrStale) {
			continue
		}
		if err != nil {
			return refusal(file, heading, err)
		}
		msg := "wrote ## " + heading
		if len(fold) > 0 {
			if err := saveSeen(file, without(seen, fold)); err != nil {
				return fmt.Errorf("wrote ## %s and folded %s, but the seen-set did not update: %v", heading, strings.Join(fold, ","), err)
			}
			msg += fmt.Sprintf(", folded %d", len(fold))
		}
		fmt.Fprintln(out, msg)
		return nil
	}
	return exitError{6, fmt.Errorf("refused: ## %s in %s is still stale after %d reads; someone is editing it", heading, file, writeTries)}
}

// writeBody hands couch the body through a private temporary file.
func writeBody(body string, write func(path string) error) error {
	f, err := os.CreateTemp("", "tl-lead-body-*.md")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(strings.Trim(body, "\n") + "\n"); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return write(f.Name())
}

func refusal(file, heading string, err error) error {
	if errors.Is(err, ErrLocked) {
		return exitError{3, fmt.Errorf("refused: the operator holds the lock on %s; queue the ## %s update and fold it in after their submit", file, heading)}
	}
	return err
}

// readBodyFile reads a body file the TL wrote; an empty one is refused, since
// writing it would blank the section.
func readBodyFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("refused: body file: %v", err)
	}
	if strings.TrimSpace(string(raw)) == "" {
		return "", fmt.Errorf("refused: body file %s is empty", path)
	}
	return string(raw), nil
}

// listOwed reads the owed events and records them as seen: the TL has now
// looked at them, so a later --fold-owed may fold them.
func listOwed(ss SessionSync, file string) ([]string, error) {
	st, err := ss.Status(file)
	if err != nil {
		return nil, err
	}
	return st.Owed, saveSeen(file, st.Owed)
}

// handleClobbered deals with parley's last-resort copy of a write that landed
// under the operator's lock: identical to the file, it is deleted; different,
// its diff is shown for the TL to merge (then `clobbered FILE --done`).
func handleClobbered(out io.Writer, file string, done bool) error {
	c := file + ".clobbered"
	cl, err := os.ReadFile(c)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	cur, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if done || string(cl) == string(cur) {
		if err := os.Remove(c); err != nil {
			return err
		}
		if done {
			fmt.Fprintln(out, "removed .clobbered")
		} else {
			fmt.Fprintln(out, "removed .clobbered (identical to the file)")
		}
		return nil
	}
	fmt.Fprintf(out, ".clobbered differs; merge what it has (+) that the file lacks, then `tl-lead status clobbered %s --done`:\n", file)
	fmt.Fprint(out, lineDiff(string(cur), string(cl)))
	return nil
}

func newStatusCmd(d deps) *cobra.Command {
	status := &cobra.Command{
		Use:   "status",
		Short: "Guarded writes to the TL's session-sync status file (over couch --session-sync)",
		Long: `Every write is compare-and-replace through couch --session-sync: read the
section's hash, compute the body from that read, write naming the hash, and on
stale read and recompute again (at most 4 times, then exit 6). The operator's
lock refuses every write (exit 3): queue the update and fold it in after their
submit. --fold-owed folds only the owed events an earlier submit or owed
listed (the seen-set, .<file>.ss-seen), so a submit that lands mid-turn stays
owed until it is read.`,
	}
	var foldOwed, done bool
	var drop []string

	submit := &cobra.Command{
		Use:   "submit FILE",
		Short: "Handle a submit: the operator's edits, their markers, .clobbered, the owed events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, file := cmd.OutOrStdout(), args[0]
			diff, err := d.ss.Diff(file)
			if err != nil {
				return err
			}
			fmt.Fprint(out, diff)
			text, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			ms := markerLines(string(text))
			if len(ms) == 0 {
				fmt.Fprintln(out, "markers: none")
			} else {
				fmt.Fprintln(out, "markers:")
				for _, m := range ms {
					fmt.Fprintf(out, "  %d: %s\n", m.Line, m.Text)
				}
			}
			if err := handleClobbered(out, file, false); err != nil {
				return err
			}
			owed, err := listOwed(d.ss, file)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, "owed:", orNone(strings.Join(owed, ", ")))
			return nil
		},
	}
	section := &cobra.Command{
		Use:   "section FILE HEADING BODYFILE",
		Short: "Replace one ## HEADING section's body",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readBodyFile(args[2])
			if err != nil {
				return err
			}
			return writeSection(d.ss, cmd.OutOrStdout(), args[0], args[1], func(string) (string, error) { return body, nil }, foldOwed)
		},
	}
	thread := &cobra.Command{
		Use:   "thread FILE TEXTFILE",
		Short: "Put a ### subsection into ## Threads (replace by heading, else before ### Miscellaneous); --drop removes subsections by heading prefix",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, err := readBodyFile(args[1])
			if err != nil {
				return err
			}
			if _, err := upsertThread("", entry, nil); err != nil {
				return err
			}
			return writeSection(d.ss, cmd.OutOrStdout(), args[0], "Threads", func(cur string) (string, error) {
				return upsertThread(cur, entry, drop)
			}, foldOwed)
		},
	}
	seen := &cobra.Command{
		Use:   "seen FILE",
		Short: "Mark the whole file seen (read --all) once a submit's markers are resolved",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := d.ss.MarkAllSeen(args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "marked the whole file seen")
			return nil
		},
	}
	owed := &cobra.Command{
		Use:   "owed FILE",
		Short: "List the owed event IDs (and record them as seen)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := listOwed(d.ss, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), orNone(strings.Join(ids, "\n")))
			return nil
		},
	}
	clobbered := &cobra.Command{
		Use:   "clobbered FILE",
		Short: "Show or (with --done) remove <file>.clobbered",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleClobbered(cmd.OutOrStdout(), args[0], done)
		},
	}
	rendered := renderedSectionCmds(d, &foldOwed)
	for _, c := range append([]*cobra.Command{section, thread}, rendered...) {
		c.Flags().BoolVar(&foldOwed, "fold-owed", false, "also fold the owed events an earlier submit/owed listed")
	}
	thread.Flags().StringArrayVar(&drop, "drop", nil, "drop subsections whose heading starts with this prefix, e.g. '### Train: old' (repeatable)")
	clobbered.Flags().BoolVar(&done, "done", false, "the TL has merged it; remove the copy")
	status.AddCommand(append([]*cobra.Command{submit, section, thread, seen, owed, clobbered}, rendered...)...)
	return status
}

// renderedSection is a status-file section rendered from sources rather than
// written by hand. Each one becomes `tl-lead status <verb> FILE`, which renders
// it and writes it through the guarded path. A later `status quota` (from
// `sdlc usage`, ariadne#365) is one more row here.
type renderedSection struct {
	verb, heading, short string
	render               func(d deps, file string) (string, error)
}

var renderedSections = []renderedSection{
	{"slots", "Slots", "Render ## Slots (every numbered slot: issue, phase, bar; then the free ones) and write it",
		func(d deps, file string) (string, error) { return gatherSlots(d.sdlc, filepath.Dir(file)) }},
}

func renderedSectionCmds(d deps, foldOwed *bool) []*cobra.Command {
	var cmds []*cobra.Command
	for _, rs := range renderedSections {
		cmds = append(cmds, &cobra.Command{
			Use:   rs.verb + " FILE",
			Short: rs.short,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				body, err := rs.render(d, args[0])
				if err != nil {
					return err
				}
				return writeSection(d.ss, cmd.OutOrStdout(), args[0], rs.heading, func(string) (string, error) { return body, nil }, *foldOwed)
			},
		})
	}
	return cmds
}

func newTrainsCmd(d deps) *cobra.Command {
	var defaultRepo string
	c := &cobra.Command{
		Use:   "trains PROJECT",
		Short: "Print a project's trains block, rendered from sources, for its ### Project thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			block, err := gatherTrains(d.sdlc, args[0], defaultRepo)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), block)
			return nil
		},
	}
	c.Flags().StringVar(&defaultRepo, "default-repo", "", "repository of a bare #N car with no qualified car before it")
	return c
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
