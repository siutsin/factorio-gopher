# Agent Instructions

## Linting Protocol

Always run `make lint` after editing any Markdown, Lua, Go, or text file and fix any failures immediately. This runs:

- `markdownlint-cli2` against all `*.md`
- `luacheck` against all `mod/*.lua`
- `golangci-lint` against all Go packages

## Testing

Run `make test` after any code change. This runs:

- `make test-go` — `go test -race -cover ./...`
- `make test-lua` — `busted -c mod/spec` plus a luacov coverage summary

`internal/SetFrameSize` is a test-only seam used by both `TestMain` blocks to shrink the 1024-px frame to 64-px so integration
tests run in seconds rather than minutes. Production code never calls it.

## Development Loop

Run `make install` once to symlink `mod/` into Factorio's mods folder. Re-run whenever `version` in `mod/info.json` changes
— Factorio identifies mods by `name_version` folder.

After editing `mod/data-updates.lua` or replacing PNGs in `mod/graphics/`, exit to the main menu and reopen the save. Sprite changes do not hot-reload.

`make uninstall` removes the symlink. It only acts if the path is a symlink.

## Sprite Modification Policy

**CRITICAL**: Always re-read the base game source before changing sprite fields. The character schema shifts between Factorio versions.

- Entity definition: `~/Library/Application Support/Steam/steamapps/common/Factorio/factorio.app/Contents/data/base/prototypes/entity/entities.lua` (search `name = "character"`)
- Animation tables: same directory, `character-animations.lua`
- Confirmed against base mod 2.0.76.

### Specific Cases

**Mutate from `data-updates.lua`, not `data.lua`** — the base prototype must already exist when we mutate it.

**Mutation target**: `data.raw.character.character`. Its `animations` field is an array of armour-tier sets.
Each set has keys `idle`, `idle_with_gun`, `running`, `running_with_gun`, `flipped_shadow_running_with_gun`, `mining_with_tool`, each with a `layers` list.
Walk every armour set when reskinning.

**Sprite paths** use the `__gopher__/...` prefix to resolve to this mod's folder.

## Asset Policy

Source files (Blender, Aseprite, references) live in `art/` and are gitignored. Only files in `mod/` ship.

PNGs only — Factorio does not load WebP. Convert with `dwebp` if needed.

Sprite sources come from <https://github.com/golang-samples/gopher-vector> (CC BY 3.0, Takuya Ueda — credit required).
Download SVGs into `art/`, then rasterise to `mod/graphics/`:

- **Free-standing sprite** (single pose, no matched set): preserve aspect, pad horizontally.

  ```bash
  rsvg-convert -h 256 -a art/foo.svg \
    | magick - -background none -gravity center -extent 256x256 mod/graphics/foo.png
  ```

- **Matched directional set** (front + side + back share a canvas): force every frame into the same square so they line up in a sheet.

  ```bash
  rsvg-convert -w 256 -h 256 -a art/foo.svg \
    | magick - -background none -gravity center -extent 256x256 mod/graphics/foo.png
  ```

## Tool Selection

| Task                | Tool                   | Purpose                                |
|---------------------|------------------------|----------------------------------------|
| Mod definition      | `info.json`            | name, version, factorio_version, deps  |
| Prototype mutation  | Lua                    | `data-updates.lua`                     |
| Sprite assets       | PNG                    | `mod/graphics/`, RGBA                  |
| SVG → PNG           | `rsvg-convert`         | Rasterise vector sources               |
| PNG canvas / sheets | `magick` (ImageMagick) | Pad, stitch, mirror frames             |
| WebP → PNG          | `dwebp`                | Source images from the web             |
| Install             | `make install`         | Symlink mod/ into Factorio mods folder |
