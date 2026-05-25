-- Replaces the character with an 8-direction gopher (N/NE/E/SE/S/SW/W/NW)
-- with shadow layer. Source frames are 1024x1024; scale = 0.0375 keeps
-- on-screen size matched to the v0.0.2 256-source/0.15-scale baseline.

local SCALE = 0.0375
-- FRAME must match `frameSize` in internal/sprite (Go pipeline). The
-- 8-frame run sheet hits 8 * FRAME = 8192 in both axes, which is exactly
-- Factorio's texture cap; raising FRAME or frame_count breaks load.
local FRAME = 1024
local SHIFT = {0, -0.5}

-- 8-direction sheet for idle. Stacked vertically; row order matches
-- Factorio's direction enum: N, NE, E, SE, S, SW, W, NW.
local gopher_8dir = {
  filename = "__gopher__/graphics/gopher-8dir.png",
  width = FRAME,
  height = FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = SCALE,
  shift = SHIFT,
}

local gopher_8dir_shadow = {
  filename = "__gopher__/graphics/gopher-shadow-8dir.png",
  width = FRAME,
  height = FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

-- Animated 8-frame run cycle, 8 directions. Body bobs; shadow doesn't.
local gopher_running = {
  filename = "__gopher__/graphics/gopher-running.png",
  width = FRAME,
  height = FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.5,
  scale = SCALE,
  shift = SHIFT,
}

local gopher_running_shadow = {
  filename = "__gopher__/graphics/gopher-shadow-running.png",
  width = FRAME,
  height = FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.5,
  scale = SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

-- Single south-facing pose for slots that only need one direction.
local gopher_s = {
  filename = "__gopher__/graphics/gopher-s.png",
  width = FRAME,
  height = FRAME,
  frame_count = 1,
  direction_count = 1,
  scale = SCALE,
  shift = SHIFT,
}

-- running_with_gun requires 18 or 40 directions (Factorio composes
-- move-direction × aim-direction from this set). We map each 20° slot
-- to its nearest of 8 directions (45° wide each):
--   idx 0-1   (0°,20°)         → N
--   idx 2-3   (40°,60°)        → NE
--   idx 4-5   (80°,100°)       → E
--   idx 6-7   (120°,140°)      → SE
--   idx 8-10  (160°,180°,200°) → S
--   idx 11-12 (220°,240°)      → SW
--   idx 13-14 (260°,280°)      → W
--   idx 15-16 (300°,320°)      → NW
--   idx 17    (340°)           → N
-- Stored as a 6x3 grid via line_length to fit under the 8192 texture limit.
local gopher_running_with_gun = {
  filename = "__gopher__/graphics/gopher-running-with-gun.png",
  width = FRAME,
  height = FRAME,
  frame_count = 1,
  direction_count = 18,
  line_length = 6,
  scale = SCALE,
  shift = SHIFT,
}

local gopher_running_with_gun_shadow = {
  filename = "__gopher__/graphics/gopher-shadow-running-with-gun.png",
  width = FRAME,
  height = FRAME,
  frame_count = 1,
  direction_count = 18,
  line_length = 6,
  scale = SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

local single_dir_keys = {
  "mining_with_tool",
}

for _, armour_set in ipairs(data.raw.character.character.animations) do
  for _, key in ipairs(single_dir_keys) do
    armour_set[key] = { layers = { gopher_s } }
  end
  armour_set.idle = { layers = { gopher_8dir, gopher_8dir_shadow } }
  armour_set.idle_with_gun = { layers = { gopher_8dir, gopher_8dir_shadow } }
  armour_set.running = { layers = { gopher_running, gopher_running_shadow } }
  armour_set.running_with_gun = {
    layers = { gopher_running_with_gun, gopher_running_with_gun_shadow }
  }
  -- shadow layer would paint a duplicate gopher offset; drop it
  armour_set.flipped_shadow_running_with_gun = nil
end
