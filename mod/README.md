# Go Gopher

A Factorio 2.0 mod that replaces the player character with the Go gopher.

## Support

| Available option                  | API ID                            | No armour | Base armour | Mech armour    |
|-----------------------------------|-----------------------------------|-----------|-------------|----------------|
| Idle                              | `idle`                            | Supported | Supported   | Supported      |
| Idle with gun                     | `idle_with_gun`                   | Supported | Supported   | Supported      |
| Running                           | `running`                         | Supported | Supported   | Supported      |
| Running with gun                  | `running_with_gun`                | Supported | Supported   | Supported      |
| Running with gun (flipped shadow) | `flipped_shadow_running_with_gun` | Supported | Supported   | Supported      |
| Mining                            | `mining_with_tool`                | Supported | Supported   | Supported      |
| Airborne idle                     | `idle_in_air`                     | N/A       | N/A         | Conditional[1] |
| Airborne idle with gun            | `idle_with_gun_in_air`            | N/A       | N/A         | Supported      |
| Flying                            | `flying`                          | N/A       | N/A         | Conditional[1] |
| Flying with gun                   | `flying_with_gun`                 | N/A       | N/A         | Conditional[1] |
| Taking off                        | `take_off`                        | N/A       | N/A         | Supported      |
| Landing                           | `landing`                         | N/A       | N/A         | Supported      |
| Corpse graphics                   | `pictures`                        | Supported | Supported   | Supported      |
| Corpse graphics (armoured)        | `armor_picture_mapping`           | N/A       | Supported   | Supported      |

[1] Not defined by vanilla Space Age; supported if present.

## Roadmap

- Player-colour accents for multiplayer differentiation
- Animated ground idle and armed-idle cycles

There is no set timeline. Contributions are welcome.

## Code

The code in this mod is copyright 2026 Simon Li and is distributed under the
MIT License. See `LICENSE` in the packaged mod archive.

Source repository: <https://github.com/siutsin/factorio-gopher>.

## Artwork

The Go gopher was designed by Renee French.

The source vector artwork was created by Takuya Ueda and published in
[golang-samples/gopher-vector](https://github.com/golang-samples/gopher-vector)
under the Creative Commons Attribution 3.0 License:
<https://creativecommons.org/licenses/by/3.0/>.

This mod rasterises, resizes, pads, and stitches those source graphics into
Factorio sprite sheets.

The mech-armour knight sprites are based on artwork by Egon Elbre from
[egonelbre/gophers](https://github.com/egonelbre/gophers), released under
[CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) (public domain).
They appear only with mech armour from the Space Age expansion.
