package macro

// LinkedExpanders holds unified expanders keyed by syntax-id for expand runs.
type LinkedExpanders struct {
	Expand map[string]Expander
}
