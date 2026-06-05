package quote

import "fmt"

func errf(format string, args ...any) error {
	return fmt.Errorf("quote: "+format, args...)
}

func errKindMismatch(kind Kind, api string) error {
	return errf("template root is @%s but called %s", kind, api)
}

func errMissingHole(name string) error {
	return errf("missing binding for hole %q", name)
}

func errBadBinding(hole string, got string) error {
	return errf("hole %q: invalid binding type (%s)", hole, got)
}
