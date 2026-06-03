package macro

// LinkedExpanders holds Call and Decl expanders keyed by syntax-id for expand runs.
type LinkedExpanders struct {
	Call map[string]CallExpander
	Decl map[string]DeclExpander
}
