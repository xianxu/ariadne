package fleet

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// RenderInventory writes a deterministic human snapshot of already-collected fleet data.
func RenderInventory(w io.Writer, inventory Inventory) error {
	if w == nil {
		return fmt.Errorf("render inventory: nil writer")
	}
	var buffer bytes.Buffer
	if err := renderInventory(&buffer, inventory); err != nil {
		return err
	}
	return writeRendered(w, buffer.Bytes())
}

func renderInventory(w io.Writer, inventory Inventory) error {
	rows := append([]TreeRow(nil), inventory.Rows...)
	for i := range rows {
		if rows[i].Issues == nil {
			rows[i].Issues = []IssueAssociation{}
		}
		if err := rows[i].validate(); err != nil {
			return fmt.Errorf("render inventory: %w", err)
		}
	}
	diagnostics := append([]RepoDiagnostic(nil), inventory.Diagnostics...)
	for _, diagnostic := range diagnostics {
		if err := diagnostic.validate(); err != nil {
			return fmt.Errorf("render inventory: %w", err)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].RepoIdentity != rows[j].RepoIdentity {
			return rows[i].RepoIdentity < rows[j].RepoIdentity
		}
		return rows[i].TreePath < rows[j].TreePath
	})
	for i := 1; i < len(rows); i++ {
		if rows[i-1].RepoIdentity == rows[i].RepoIdentity && rows[i-1].TreePath == rows[i].TreePath {
			return fmt.Errorf("render inventory: duplicate tree identity")
		}
	}
	sort.Slice(diagnostics, func(i, j int) bool { return lessDiagnostic(diagnostics[i], diagnostics[j]) })
	for _, row := range rows {
		state := "branch=" + quote(row.Branch)
		if row.Detached {
			state = "detached"
		}
		if row.Bare {
			state = "bare"
		}
		if _, err := fmt.Fprintf(w, "tree=%s\trepo_identity=%s\trepo_root=%s\t%s\n", quote(row.TreePath), quote(row.RepoIdentity), quote(row.RepoRoot), state); err != nil {
			return err
		}
		if err := renderFacts(w, row.Facts); err != nil {
			return err
		}
		if row.Locked != nil {
			if _, err := fmt.Fprintf(w, "  locked=%s\n", quote(*row.Locked)); err != nil {
				return err
			}
		}
		if row.Prunable != nil {
			if _, err := fmt.Fprintf(w, "  prunable=%s\n", quote(*row.Prunable)); err != nil {
				return err
			}
		}
		for _, issue := range row.Issues {
			if _, err := fmt.Fprintf(w, "  issue=%s status=%s provenance=%s\n", quote(issue.Ref), quote(issue.DeclaredStatus), quote(issue.Provenance)); err != nil {
				return err
			}
		}
		if err := renderCapability(w, row.Policy); err != nil {
			return err
		}
	}
	for _, diagnostic := range diagnostics {
		if _, err := fmt.Fprintf(w, "diagnostic repo_path=%s stage=%s message=%s", quote(diagnostic.RepoPath), quote(diagnostic.Stage), quote(diagnostic.Message)); err != nil {
			return err
		}
		if diagnostic.RepoIdentity != "" {
			fmt.Fprintf(w, " repo_identity=%s", quote(diagnostic.RepoIdentity))
		}
		if diagnostic.TreePath != "" {
			fmt.Fprintf(w, " tree_path=%s", quote(diagnostic.TreePath))
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}

func RenderPolicy(w io.Writer, result PolicyResult) error {
	if w == nil {
		return fmt.Errorf("render policy: nil writer")
	}
	var buffer bytes.Buffer
	if err := renderPolicy(&buffer, result); err != nil {
		return err
	}
	return writeRendered(w, buffer.Bytes())
}

func renderPolicy(w io.Writer, result PolicyResult) error {
	if err := validatePolicyResult(result); err != nil {
		return fmt.Errorf("render policy: %w", err)
	}
	if !result.OK {
		_, err := fmt.Fprintf(w, "policy diagnostic code=%s message=%s path=%s", quote(result.Diagnostic.Code), quote(result.Diagnostic.Message), quote(result.Diagnostic.Path))
		if result.Diagnostic.PolicyVersion != nil {
			_, err = fmt.Fprintf(w, " policy_version=%d", *result.Diagnostic.PolicyVersion)
		}
		if err == nil {
			_, err = fmt.Fprintln(w)
		}
		return err
	}
	v := result.Value
	_, err := fmt.Fprintf(w, "policy version=%d digest=%s repo_identity=%s admission_key=%s capacity=%s", v.PolicyVersion, quote(v.PolicyDigest), quote(v.RepoIdentity), quote(v.AdmissionKey), quote(v.Capacity.Kind))
	if err != nil {
		return err
	}
	if v.Capacity.Limit != nil {
		fmt.Fprintf(w, " limit=%d", *v.Capacity.Limit)
	}
	if v.OnCapacity != "" {
		if _, err = fmt.Fprintf(w, " on_capacity=%s", quote(v.OnCapacity)); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w)
	return err
}

func writeRendered(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func renderFacts(w io.Writer, facts MeasuredFacts) error {
	if !facts.Available {
		_, err := fmt.Fprint(w, "  facts unavailable")
		if facts.Head != "" {
			_, err = fmt.Fprintf(w, " head=%s", quote(facts.Head))
		}
		if facts.CommitTimestamp != "" {
			_, err = fmt.Fprintf(w, " commit_timestamp=%s", quote(facts.CommitTimestamp))
		}
		if facts.DirtyCount != nil {
			_, err = fmt.Fprintf(w, " dirty_count=%d", *facts.DirtyCount)
		}
		if err == nil {
			_, err = fmt.Fprintf(w, " error=%s\n", quote(facts.Error))
		}
		return err
	}
	if _, err := fmt.Fprintf(w, "  facts head=%s commit_timestamp=%s dirty_count=%d", quote(facts.Head), quote(facts.CommitTimestamp), *facts.DirtyCount); err != nil {
		return err
	}
	if facts.BaseAvailable {
		_, err := fmt.Fprintf(w, " base_ref=%s ahead=%d behind=%d\n", quote(facts.BaseRef), *facts.Ahead, *facts.Behind)
		return err
	}
	_, err := fmt.Fprintf(w, " base_unavailable error=%s", quote(facts.BaseError))
	if facts.BaseRef != "" {
		_, err = fmt.Fprintf(w, " base_ref=%s", quote(facts.BaseRef))
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}
func renderCapability(w io.Writer, policy PolicyCapability) error {
	if err := validatePolicyCapability(policy); err != nil {
		return err
	}
	if !policy.OK {
		_, err := fmt.Fprintf(w, "  policy diagnostic code=%s message=%s", quote(policy.Diagnostic.Code), quote(policy.Diagnostic.Message))
		if policy.Diagnostic.Path != "" {
			_, err = fmt.Fprintf(w, " path=%s", quote(policy.Diagnostic.Path))
		}
		if policy.Diagnostic.PolicyVersion != nil {
			_, err = fmt.Fprintf(w, " policy_version=%d", *policy.Diagnostic.PolicyVersion)
		}
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w)
		return err
	}
	v := policy.Value
	roots := make([]string, len(v.Roots))
	for i := range v.Roots {
		roots[i] = quote(v.Roots[i])
	}
	_, err := fmt.Fprintf(w, "  policy=capability version=%d digest=%s key_kind=%s roots=%s capacity=%s", v.PolicyVersion, quote(v.PolicyDigest), quote(v.KeyKind), strings.Join(roots, ","), quote(v.Capacity.Kind))
	if err != nil {
		return err
	}
	if v.Capacity.Limit != nil {
		fmt.Fprintf(w, " limit=%d", *v.Capacity.Limit)
	}
	if v.OnCapacity != "" {
		if _, err = fmt.Fprintf(w, " on_capacity=%s", quote(v.OnCapacity)); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w)
	return err
}
func quote(value string) string { return strconv.Quote(value) }
func lessDiagnostic(a, b RepoDiagnostic) bool {
	av, bv := []string{a.RepoIdentity, a.RepoPath, a.Stage, a.TreePath, a.Message}, []string{b.RepoIdentity, b.RepoPath, b.Stage, b.TreePath, b.Message}
	for i := range av {
		if av[i] != bv[i] {
			return av[i] < bv[i]
		}
	}
	return false
}
