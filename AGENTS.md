# Agent Instructions

## Project Documentation

Read `README.md` for the support matrix, installation steps, development
commands, release workflow, and licences. Keep user-facing project guidance in
the README; this file should contain only agent-specific workflow and technical
invariants.

After every code, asset, behaviour, or workflow update, review the project
documentation and update all affected files in the same change.

## Instruction Scope

When working in a subdirectory, check it and each directory between it and the
repository root for an `AGENTS.md`. Read every applicable file and follow the
most specific instructions for the files being changed.

## Linting Protocol

Always run `make lint` after editing any Markdown, Lua, Go, or text file and
fix any failures immediately. This runs:

- `markdownlint-cli2` against all `*.md`
- `luacheck` against all `mod/*.lua`
- `golangci-lint` against all Go packages

## Testing

Run `make test` after any code change. This runs:

- `make test-go` with `go test -race -cover ./...`
- `make test-lua` with `busted -c mod/spec` plus luacov coverage

`internal/SetFrameSize` is a test-only seam used by the internal test suite to
shrink 1024-px sources to 64 px and 256-px runtime frames to 16 px so
integration tests run in seconds rather than minutes. Production code never
calls it.

## Development Loop

Run `make install` once to symlink `mod/` into Factorio's mods folder. Re-run
whenever `version` in `mod/info.json` changes; Factorio identifies mods by
`name_version` folder.

After editing `mod/data-updates.lua` or replacing PNGs in `mod/graphics/`, exit
to the main menu and reopen the save. Sprite changes do not hot-reload.

`make uninstall` removes the symlink. It only acts if the path is a symlink.

## Sprite Modification Policy

**CRITICAL**: Always re-read the base game source before changing sprite
fields. The character schema shifts between Factorio versions.

Inspect the files from the Factorio version targeted by `mod/info.json`. Search
the game installation for:

- `data/base/prototypes/entity/entities.lua` (then search for
  `name = "character"`)
- `data/base/prototypes/entity/character-animations.lua`
- `data/space-age/prototypes/entity/mech-armor-animations.lua` when validating
  mech armour

On macOS these paths are inside `factorio.app/Contents/`. Steam users can
locate the installation with **Manage > Browse local files**.

### Specific Cases

**Mutate from `data-updates.lua`, not `data.lua`**; the base prototype must
already exist when we mutate it.

**Mutation target**: `data.raw.character.character`. Its `animations` field is
an array of armour-tier sets.
Each set has ground keys `idle`, `idle_with_gun`, `running`,
`running_with_gun`, `flipped_shadow_running_with_gun`, and
`mining_with_tool`, each with a `layers` list. The Space Age mech set also has
`take_off`, `landing`, `idle_with_gun_in_air`, and `smoke_in_air`.
Optional mech fields `idle_in_air`, `flying`, and `flying_with_gun` are
replaced only when another mod defines them.
Walk every armour set when reskinning. The default set and sets containing only
base-game armour names use the gopher. Only a set containing exactly
`mech-armor` uses knight sheets. Leave every set containing an unknown armour
name untouched.

**Mining invariant**: generated gopher and knight mining sheets contain body
motion plus fitted shadows only. Do not draw a synthetic tool into either
sheet; there are no per-direction hand anchors to keep it aligned.

**Corpse target**:
`data.raw["character-corpse"]["character-corpse"]`. Its `pictures` field
contains armour variations; the first is the default, and
`armor_picture_mapping` assigns armour names to variation indices. Factorio
uses the highest mapped armour item found anywhere in the corpse inventory,
not only the item worn at death. Both frames in every custom corpse animation
must already be horizontal fallen poses; they are not an upright-to-fall
transition. Remap only built-in armour names and `mech-armor`; preserve unknown
mapping entries and their picture variations. The base corpse never expires,
so remove old corpses before comparing variations in-game.

**Sprite paths** use the `__gopher__/...` prefix to resolve to this mod's
folder.

## Asset Policy

The canonical generator inputs are these tracked 1024-by-1024 RGBA PNGs:

- `mod/graphics/gopher-{n,ne,e,se,s,sw,w,nw}.png`
- `mod/graphics/knight-{n,ne,e,se,s,sw,w,nw}.png`

`make build` regenerates every derived sprite sheet from those files. A clean
clone must contain everything the build requires. `make package` removes
directional inputs that Factorio does not load.

Factorio assets must be PNGs. Keep artwork attribution and licence details in
`README.md` and `mod/README.md`.

## Tool Selection

- Mod definition: `info.json` for name, version, Factorio version, and deps.
- Prototype mutation: Lua in `data-updates.lua`.
- Corpse mutation: `data-updates.lua` targeting
  `data.raw["character-corpse"]["character-corpse"]`.
- Sprite inputs and runtime assets: tracked RGBA PNGs under `mod/graphics/`.
- Sprite generation: `make build` (or `go run ./cmd all`).
- Generated-asset verification: after `make build` in a clean clone,
  `git status --short --untracked-files=all -- mod/graphics` must print
  nothing.
- Install: `make install` to symlink `mod/` into Factorio's mods folder.
