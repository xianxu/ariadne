package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// NewProjectCmd returns the model-backed project workflow command group.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "project",
		Short:         "Create and manage workshop projects",
		Long:          "Placeholder — replaced by renderLong(\"project\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newProjectNewCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectSetStatusCmd())
	cmd.AddCommand(newProjectValidateCmd())
	cmd.AddCommand(newProjectStatusCmd())
	cmd.AddCommand(newProjectRetroCmd())
	cmd.AddCommand(newProjectCloseCmd())
	cmd.AddCommand(newProjectFindCmd())
	return cmd
}

type projectNewFlags struct {
	Slug, Goal, DoneWhen, ProjectsDir string
}
type projectListFlags struct{ ProjectsDir string }
type projectShowFlags struct{ Slug, ProjectsDir string }
type projectValidateFlags struct {
	Slug, ProjectsDir string
	All               bool
}

var projectTodayFn = func() string { return time.Now().Format("2006-01-02") }

func defaultProjectsDir() string { return envOr("WF_PROJECTS_DIR", vocab.Project().Discovery().Home) }

func newProjectNewCmd() *cobra.Command {
	f := projectNewFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "new", Short: "Create a project from the model-derived scaffold", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectNew(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}})
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project filename/name slug")
	cmd.Flags().StringVar(&f.Goal, "goal", "", "one-sentence project goal")
	cmd.Flags().StringVar(&f.DoneWhen, "done-when", "", "falsifiable MVP completion boundary")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	_ = cmd.MarkFlagRequired("slug")
	_ = cmd.MarkFlagRequired("goal")
	_ = cmd.MarkFlagRequired("done-when")
	return cmd
}

func newProjectListCmd() *cobra.Command {
	f := projectListFlags{}
	cmd := &cobra.Command{Use: "list", Short: "List live projects", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectList(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}}
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	return cmd
}

func newProjectShowCmd() *cobra.Command {
	f := projectShowFlags{}
	cmd := &cobra.Command{Use: "show", Short: "Show one project and its task summary", Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProjectShow(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		}}
	cmd.Flags().StringVar(&f.Slug, "slug", "", "project slug")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	_ = cmd.MarkFlagRequired("slug")
	return cmd
}

func newProjectValidateCmd() *cobra.Command {
	f := projectValidateFlags{}
	cmd := &cobra.Command{Use: "validate [<file>...]", Short: "Validate project records against #Project", Args: cobra.ArbitraryArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectValidate(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args)
		}}
	cmd.Flags().StringVar(&f.Slug, "slug", "", "validate one project slug")
	cmd.Flags().BoolVar(&f.All, "all", false, "validate all live projects")
	cmd.Flags().StringVar(&f.ProjectsDir, "projects-dir", defaultProjectsDir(), "directory holding project files")
	return cmd
}

func runProjectNew(stdout, _ io.Writer, f *projectNewFlags) error {
	for name, value := range map[string]string{"slug": f.Slug, "goal": f.Goal, "done-when": f.DoneWhen} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("--%s is required and must be non-empty", name)
		}
	}
	dest, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("project already exists: %s", dest)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(f.ProjectsDir, 0o755); err != nil {
		return err
	}
	body := projectdoc.RenderScaffold(projectdoc.ScaffoldSpec{Name: f.Slug, Goal: f.Goal, DoneWhen: f.DoneWhen, Today: projectTodayFn()})
	if err := os.WriteFile(dest, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(stdout, dest)
	return nil
}

func projectFiles(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, vocab.Project().Discovery().Glob))
	sort.Strings(files)
	return files, err
}

func readProject(path string) (*projectdoc.Doc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return projectdoc.ParseDoc(string(b))
}

func runProjectList(stdout, _ io.Writer, f *projectListFlags) error {
	files, err := projectFiles(f.ProjectsDir)
	if err != nil {
		return err
	}
	for _, path := range files {
		d, err := readProject(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		fmt.Fprint(stdout, projectdoc.RenderListRow(projectdoc.Summarize(path, d)))
	}
	return nil
}

func runProjectShow(stdout, _ io.Writer, f *projectShowFlags) error {
	path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
	if err != nil {
		return err
	}
	d, err := readProject(path)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, projectdoc.RenderShow(projectdoc.Summarize(path, d)))
	return nil
}

func runProjectValidate(stdout, stderr io.Writer, f *projectValidateFlags, args []string) error {
	if f.All && (f.Slug != "" || len(args) > 0) || f.Slug != "" && len(args) > 0 {
		return fmt.Errorf("choose one of <file>, --slug, or --all")
	}
	files := args
	if f.Slug != "" {
		path, err := projectdoc.ResolvePath(f.ProjectsDir, f.Slug)
		if err != nil {
			return err
		}
		files = []string{path}
	}
	if f.All {
		var err error
		files, err = projectFiles(f.ProjectsDir)
		if err != nil {
			return err
		}
	}
	if len(files) == 0 {
		return fmt.Errorf("specify <file>, --slug, or --all")
	}
	bad := 0
	for _, path := range files {
		out, ok, err := validateFrontmatterFn("project", path)
		if err != nil {
			return err
		}
		if !ok {
			bad++
			fmt.Fprintf(stdout, "%s:\n%s\n", path, out)
		} else {
			cok(stderr, path+": conforms")
		}
	}
	if bad > 0 {
		return fmt.Errorf("%d of %d project file(s) nonconforming", bad, len(files))
	}
	return nil
}
