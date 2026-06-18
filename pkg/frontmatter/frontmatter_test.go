package frontmatter

import "testing"

// Description reads the flat-YAML `description:` from a markdown file's leading
// `---`…`---` frontmatter fence — the one field the skill menu / prototype
// reader needs. Quotes on the value are stripped; an absent fence or absent
// description yields "". Ported from the behavior weave's frontmatterDescription
// exercised via skills_test.go (the `---\nname: …\ndescription: …\n---` form).

func TestDescriptionPlain(t *testing.T) {
	content := "---\nname: sdlc\ndescription: SDLC checkpoint gates\n---\n\nBODY\n"
	if got := Description(content); got != "SDLC checkpoint gates" {
		t.Fatalf("Description = %q, want %q", got, "SDLC checkpoint gates")
	}
}

func TestDescriptionDoubleQuoted(t *testing.T) {
	content := "---\ndescription: \"quoted value\"\n---\n"
	if got := Description(content); got != "quoted value" {
		t.Fatalf("Description = %q, want %q", got, "quoted value")
	}
}

func TestDescriptionSingleQuoted(t *testing.T) {
	content := "---\ndescription: 'quoted value'\n---\n"
	if got := Description(content); got != "quoted value" {
		t.Fatalf("Description = %q, want %q", got, "quoted value")
	}
}

func TestDescriptionNoFence(t *testing.T) {
	// A body with no leading `---` fence has no frontmatter — "".
	content := "description: not in a fence\n\nbody\n"
	if got := Description(content); got != "" {
		t.Fatalf("Description = %q, want \"\" (no fence)", got)
	}
}

func TestDescriptionAbsentField(t *testing.T) {
	content := "---\nname: sdlc\n---\n\nbody\n"
	if got := Description(content); got != "" {
		t.Fatalf("Description = %q, want \"\" (no description field)", got)
	}
}

func TestDescriptionStopsAtClosingFence(t *testing.T) {
	// A `description:` AFTER the closing fence (in the body) is NOT read.
	content := "---\nname: sdlc\n---\n\ndescription: in the body, ignored\n"
	if got := Description(content); got != "" {
		t.Fatalf("Description = %q, want \"\" (description in body, not frontmatter)", got)
	}
}

func TestDescriptionEmptyContent(t *testing.T) {
	if got := Description(""); got != "" {
		t.Fatalf("Description(\"\") = %q, want \"\"", got)
	}
}

func TestUnquoteUnbalanced(t *testing.T) {
	// A single leading quote is not a symmetric pair — left intact.
	if got := unquote("\"unbalanced"); got != "\"unbalanced" {
		t.Fatalf("unquote = %q, want unchanged", got)
	}
}
