# factorio-gopher

A Factorio 2.0 mod that replaces the player character with the Go gopher.

## Install

```bash
make install
```

This symlinks `mod/` into Factorio's per-OS mods folder:

- macOS: `~/Library/Application Support/factorio/mods`
- Linux: `~/.factorio/mods`
- Windows: `%APPDATA%\Factorio\mods`

Then launch Factorio → Mods → enable "Go Gopher" → New Game.

`make uninstall` removes the symlink. Override the location with `go run ./cmd -mods <path> install`.

## Development

### Tooling

Install via your package manager of choice. For example, on macOS with Homebrew:

```bash
brew install go jq golangci-lint busted luarocks luacheck markdownlint-cli2
luarocks install --local luacov
```

### Sprite pipeline

Per-direction source PNGs live in `mod/graphics/gopher-{n,ne,e,se,s,sw,w,nw}.png` at 1024×1024. Edit those, then
regenerate the derived sheets:

```bash
make build
```

This runs `go run ./cmd all` and writes `gopher-running.png`, `gopher-shadow-*.png`, `gopher-8dir.png`, and
`gopher-running-with-gun.png`. Sprite changes do not hot-reload — exit to Factorio's main menu and reopen the save to
pick them up.

The Go pipeline lives in `cmd/main.go` (entry point) and `internal/*.go` (run cycle, shadow projection, sheet
stitching, PNG helpers). The Factorio prototype mutation is in `mod/data-updates.lua`.

### Make targets

```bash
make help    # list every target
make build   # regenerate sprite sheets
make package # create build/gopher_<version>.zip for Factorio
make test    # Go (-race -cover) + Lua (busted + luacov), both at 100% coverage
make lint    # markdown + lua + go
```

See `AGENTS.md` for the detailed development workflow and conventions.

## Licences

Code: MIT (see `LICENSE`).

Gopher artwork: the Go gopher was designed by Renee French.
Vector artwork by Takuya Ueda, from
[golang-samples/gopher-vector](https://github.com/golang-samples/gopher-vector),
licensed under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/).
The vector source was rasterised to PNG and padded to a square canvas for use in this mod.
