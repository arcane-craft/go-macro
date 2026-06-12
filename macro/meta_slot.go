package macro

// MetaSlot is implemented by engine site Syntax for internal meta lifecycle.
type MetaSlot interface {
	Syntax
	ClearExpansionMeta()
}
