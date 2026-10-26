// side-doc — typed frontmatter for a document in an issue's or plan's side
// tree. It has no lifecycle. Its parent is a reference, not an existence check.
package side_doc

#SideDoc: {
	type!:   "side-doc"
	parent!: string & !=""
	// Additive metadata (including marker) belongs to the document's author.
	...
}
