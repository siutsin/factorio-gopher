# factorio-gopher

A Factorio 2.0 mod that replaces the player character with the Go gopher.

## Support

| Available Option                  | API ID                            | No Armour | Base Armour | Mech Armour |
|-----------------------------------|-----------------------------------|-----------|-------------|-------------|
| Idle                              | `idle`                            | ✅        | ✅          | ✅          |
| Idle with gun                     | `idle_with_gun`                   | ✅        | ✅          | ✅          |
| Running                           | `running`                         | ✅        | ✅          | ✅          |
| Running with gun                  | `running_with_gun`                | ✅        | ✅          | ✅          |
| Running with gun (flipped shadow) | `flipped_shadow_running_with_gun` | ✅        | ✅          | ✅          |
| Mining                            | `mining_with_tool`                | ✅        | ✅          | ✅          |
| Airborne idle                     | `idle_in_air`                     | N/A       | N/A         | ⚠️[^1]      |
| Airborne idle with gun            | `idle_with_gun_in_air`            | N/A       | N/A         | ✅          |
| Flying                            | `flying`                          | N/A       | N/A         | ⚠️[^1]      |
| Flying with gun                   | `flying_with_gun`                 | N/A       | N/A         | ⚠️[^1]      |
| Taking off                        | `take_off`                        | N/A       | N/A         | ✅          |
| Landing                           | `landing`                         | N/A       | N/A         | ✅          |
| Corpse graphics                   | `pictures`                        | ✅        | ✅          | ✅          |
| Corpse graphics (armoured)        | `armor_picture_mapping`           | N/A       | ✅          | ✅          |

[^1]: Not defined by vanilla Space Age; supported if present.

## Roadmap

- Player-colour accents for multiplayer differentiation
- Animated ground idle and armed-idle cycles

There is no set timeline. Contributions are welcome.

## Install

```bash
make install
```

This symlinks `mod/` into Factorio's per-OS mods folder:

- macOS: `~/Library/Application Support/factorio/mods`
- Linux: `~/.factorio/mods`
- Windows: `%APPDATA%\Factorio\mods`

Then launch Factorio, open Mods, enable "Go Gopher", and load or start a game.

`make uninstall` removes the symlink. Override the location with
`go run ./cmd -mods <path> install`.

## Development

### Tooling

Install via your package manager of choice. For example, on macOS with Homebrew:

```bash
brew install go jq golangci-lint busted luarocks luacheck markdownlint-cli2
luarocks install --local luacov
```

### Sprite pipeline

Per-direction build-input PNGs live in
`mod/graphics/gopher-{n,ne,e,se,s,sw,w,nw}.png` and
`mod/graphics/knight-{n,ne,e,se,s,sw,w,nw}.png` at 1024 px by 1024 px. Edit
those inputs, then regenerate the derived sheets:

```bash
make build
```

Derived runtime frames are 256 px by 256 px and render at `scale = 0.15`.
Keeping the tracked inputs larger preserves generator quality without making
Factorio decode the high-resolution working canvases at runtime.

This runs `go run ./cmd all` and writes `gopher-running.png`,
`gopher-shadow-*.png`, `gopher-8dir.png`,
`gopher-running-with-gun-*.png`, `gopher-idle-with-gun*.png`,
`knight-idle*.png`,
`knight-running.png`, `knight-running-shadow.png`,
`knight-running-with-gun*.png`, `knight-flying-with-gun*.png`,
`knight-take-off*.png`, and
`knight-hover*.png`, plus the `gopher-corpse*.png` and
`knight-corpse*.png` death sheets and both `*-mining*.png` cycles.
Sprite changes do not hot-reload; exit to
Factorio's main menu and reopen the save to pick them up.

The Go pipeline lives in `cmd/main.go` (entry point) and `internal/*.go` (run
cycle, shadow projection, sheet stitching, PNG helpers). The Factorio prototype
mutation is in `mod/data-updates.lua`.

### Make targets

```bash
make help    # list every target
make build   # regenerate sprite sheets
make package # create build/gopher_<version>.zip for Factorio
make test    # Go and Lua tests; race detection + 100% coverage gates
make lint    # markdown + lua + go
```

### Releases

Pushing a numeric version tag runs the `Publish Factorio Mod` workflow. The tag
must match `mod/info.json` and point to a commit on `master`. The workflow
rebuilds and verifies the sprites, runs the full validation suite, packages the
mod, then uses `go run ./cmd/publish` to upload it through the configured
`Automation` environment.

Factorio release versions cannot be uploaded twice. Always bump
`mod/info.json` before creating the next tag.

## Licences

Code: MIT (see `LICENSE`).

Gopher artwork: the Go gopher was designed by Renee French.
Vector artwork by Takuya Ueda, from
[golang-samples/gopher-vector](https://github.com/golang-samples/gopher-vector),
licensed under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/).
The vector source was rasterised to PNG and padded to a square canvas for use in
this mod.

Knight gopher sprites are based on artwork by Egon Elbre, from
[egonelbre/gophers](https://github.com/egonelbre/gophers),
released under [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/)
(public domain). They appear only with mech armour from the Space Age
expansion.
