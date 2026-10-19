package issue

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// mergeUnknownCardNodes carries only members this binary does not own into a
// replacement of the same known record. A removed known record stays removed.
// The source has already passed validateCardYAML, including recursive duplicate
// and alias checks. The rendered replacement is produced from validated structs.
func mergeUnknownCardNodes(source, replacement *yaml.Node, shape reflect.Type) {
	for shape.Kind() == reflect.Pointer {
		shape = shape.Elem()
	}
	if shape.Kind() != reflect.Struct || source.Kind != yaml.MappingNode || replacement.Kind != yaml.MappingNode {
		return
	}
	known := map[string]reflect.Type{}
	for i := 0; i < shape.NumField(); i++ {
		f := shape.Field(i)
		if !f.IsExported() {
			continue
		}
		parts := strings.Split(f.Tag.Get("yaml"), ",")
		if len(parts) == 0 || parts[0] == "" || parts[0] == "-" || slices.Contains(parts[1:], "inline") {
			continue
		}
		known[parts[0]] = f.Type
	}
	mergeUnknownCardFields(source, replacement, known)
}

func mergeUnknownCardFields(source, replacement *yaml.Node, known map[string]reflect.Type) {
	for i := 0; i < len(source.Content); i += 2 {
		key, value := source.Content[i], source.Content[i+1]
		var next *yaml.Node
		for j := 0; j < len(replacement.Content); j += 2 {
			if replacement.Content[j].Value == key.Value {
				next = replacement.Content[j+1]
				break
			}
		}
		if childShape, owned := known[key.Value]; owned {
			if next != nil {
				mergeUnknownCardNodes(value, next, childShape)
			}
		} else if next == nil {
			replacement.Content = append(replacement.Content, key, value)
		}
	}
}

func preserveUnknownBlock(source *yaml.Node, rendered []byte, shape reflect.Type) (string, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(rendered, &doc); err != nil {
		return "", err
	}
	if len(doc.Content) != 1 || len(doc.Content[0].Content) != 2 {
		return "", fmt.Errorf("invalid rendered card block")
	}
	mergeUnknownCardNodes(source, doc.Content[0].Content[1], shape)
	out, err := yaml.Marshal(doc.Content[0])
	return strings.TrimSuffix(string(out), "\n"), err
}

func preserveClaimantUnknown(source *yaml.Node, rendered string) (string, error) {
	// Claimant uses a manual key list rather than YAML tags. Its known keys are
	// single-line scalars, and unknown children can be appended unchanged.
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(rendered), &doc); err != nil {
		return "", err
	}
	if len(doc.Content) != 1 || len(doc.Content[0].Content) != 2 {
		return "", fmt.Errorf("invalid rendered claimant")
	}
	replacement := doc.Content[0].Content[1]
	known := map[string]bool{}
	for _, field := range claimantKeys {
		known[field.key] = true
	}
	fields := map[string]reflect.Type{}
	for key := range known {
		fields[key] = reflect.TypeOf("")
	}
	mergeUnknownCardFields(source, replacement, fields)
	out, err := yaml.Marshal(doc.Content[0])
	return strings.TrimSuffix(string(out), "\n"), err
}

func preserveTrackerUnknown(source *yaml.Node, rendered []byte) ([]byte, error) {
	text, err := preserveUnknownBlock(source, rendered, reflect.TypeOf(trackerEnvelope{}))
	return []byte(text), err
}
