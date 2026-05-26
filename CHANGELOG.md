# Changelog

## Unreleased

### BREAKING

Official macro libraries moved from `github.com/arcane-craft/go-macro/contrib` to independent module `github.com/arcane-craft/go-macro-contrib`.

| Old import | New import |
|------------|------------|
| `.../go-macro/contrib/inline` | `.../go-macro-contrib/inline` |
| `.../go-macro/contrib/try` | `.../go-macro-contrib/try` |
| `.../go-macro/contrib/register` | `.../go-macro-contrib/register` |

```bash
go get github.com/arcane-craft/go-macro-contrib@v0.1.0
```

Local development: clone `go-macro-contrib` as a sibling of this repo (`../go-macro-contrib` from repo root). The `examples` module uses `replace github.com/arcane-craft/go-macro-contrib => ../../go-macro-contrib`.

Removed in-repo `contrib/` directory. Root `go.work` now only includes root and `./examples`.
