-- Replaces the character with an 8-direction gopher (N/NE/E/SE/S/SW/W/NW)
-- with shadow layer. Source frames are 1024x1024; scale = 0.0375 keeps
-- on-screen size matched to the 0.0.2 256-source/0.15-scale baseline.

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

-- Knight gopher for mech-armour flight (Space Age). Half-size source frames
-- use double the scale so the knight matches the 1024px ground sprites while
-- 16 animation columns still fit Factorio's 8192px texture limit.
local KNIGHT_FRAME = FRAME / 2
local KNIGHT_SCALE = SCALE * 2
-- The authored knight feet sit three rendered pixels above the normal gopher
-- baseline. Offset every flight frame together so takeoff starts in place.
local KNIGHT_SHIFT = {0, -0.40625}
local LANDING_SEQUENCE = {
  16, 15, 14, 13, 12, 11, 10, 9,
  8, 7, 6, 5, 4, 3, 2, 1,
}

local knight_idle = {
  filename = "__gopher__/graphics/knight-idle.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_idle_shadow = {
  filename = "__gopher__/graphics/knight-idle-shadow.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local knight_mining = {
  filename = "__gopher__/graphics/knight-idle.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  repeat_count = 27,
  direction_count = 8,
  animation_speed = 0.45,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_mining_shadow = {
  filename = "__gopher__/graphics/knight-idle-shadow.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  repeat_count = 27,
  direction_count = 8,
  animation_speed = 0.45,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local knight_running = {
  filename = "__gopher__/graphics/knight-running.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.5,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_running_shadow = {
  filename = "__gopher__/graphics/knight-running-shadow.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.5,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local knight_running_with_gun = {
  filename = "__gopher__/graphics/knight-running-with-gun.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  direction_count = 18,
  line_length = 6,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_running_with_gun_shadow = {
  filename = "__gopher__/graphics/knight-running-with-gun-shadow.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  direction_count = 18,
  line_length = 6,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local function knight_flight_animation(name, frame_count, animation_speed, frame_sequence)
  local body = {
    filename = "__gopher__/graphics/knight-" .. name .. ".png",
    width = KNIGHT_FRAME,
    height = KNIGHT_FRAME,
    frame_count = frame_count,
    direction_count = 8,
    line_length = frame_count,
    animation_speed = animation_speed,
    scale = KNIGHT_SCALE,
    shift = KNIGHT_SHIFT,
    usage = "player",
  }
  local shadow = {
    filename = "__gopher__/graphics/knight-" .. name .. "-shadow.png",
    width = KNIGHT_FRAME,
    height = KNIGHT_FRAME,
    frame_count = frame_count,
    direction_count = 8,
    line_length = frame_count,
    animation_speed = animation_speed,
    scale = KNIGHT_SCALE,
    shift = KNIGHT_SHIFT,
    usage = "player",
    draw_as_shadow = true,
  }
  if frame_sequence then
    body.frame_sequence = frame_sequence
    shadow.frame_sequence = frame_sequence
  end
  return { layers = { body, shadow } }
end

local knight_take_off = knight_flight_animation("take-off", 16, 0.6)
local knight_landing = knight_flight_animation("take-off", 16, 0.6, LANDING_SEQUENCE)
local knight_hover = knight_flight_animation("hover", 5, 0.2)

local function has_armour(armour_set, armour_name)
  for _, name in ipairs(armour_set.armors or {}) do
    if name == armour_name then
      return true
    end
  end
  return false
end

local function use_gopher(armour_set)
  armour_set.mining_with_tool = { layers = { gopher_s } }
  armour_set.idle = { layers = { gopher_8dir, gopher_8dir_shadow } }
  armour_set.idle_with_gun = { layers = { gopher_8dir, gopher_8dir_shadow } }
  armour_set.running = { layers = { gopher_running, gopher_running_shadow } }
  armour_set.running_with_gun = {
    layers = { gopher_running_with_gun, gopher_running_with_gun_shadow }
  }
  -- shadow layer would paint a duplicate gopher offset; drop it
  armour_set.flipped_shadow_running_with_gun = nil
end

local function use_knight(armour_set)
  armour_set.idle = { layers = { knight_idle, knight_idle_shadow } }
  armour_set.idle_with_gun = { layers = { knight_idle, knight_idle_shadow } }
  armour_set.mining_with_tool = { layers = { knight_mining, knight_mining_shadow } }
  armour_set.running = { layers = { knight_running, knight_running_shadow } }
  armour_set.running_with_gun = {
    layers = { knight_running_with_gun, knight_running_with_gun_shadow }
  }
  armour_set.flipped_shadow_running_with_gun = nil

  if armour_set.take_off then
    armour_set.take_off = knight_take_off
  end
  if armour_set.landing then
    armour_set.landing = knight_landing
  end
  if armour_set.idle_with_gun_in_air then
    armour_set.idle_with_gun_in_air = knight_hover
  end
  -- The knight has no thruster vents.
  armour_set.smoke_in_air = nil
end

for _, armour_set in ipairs(data.raw.character.character.animations) do
  if has_armour(armour_set, "mech-armor") then
    use_knight(armour_set)
  else
    use_gopher(armour_set)
  end
end
