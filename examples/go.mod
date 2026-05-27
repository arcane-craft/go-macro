module github.com/arcane-craft/go-macro/examples

go 1.22.0

require (
	github.com/arcane-craft/go-macro v0.1.0
	github.com/arcane-craft/go-macro-contrib v0.1.1
)

require (
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
)

replace github.com/arcane-craft/go-macro => ../

replace github.com/arcane-craft/go-macro-contrib => ../../go-macro-contrib
