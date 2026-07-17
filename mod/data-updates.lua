-- Replaces the character with an 8-direction gopher (N/NE/E/SE/S/SW/W/NW)
-- with shadow layer. Canonical generator inputs are 1024x1024; runtime frames
-- are 256x256 at the original 0.15 scale to limit decoded texture memory.

local SCALE = 0.15
-- FRAME must match `runtimeFrameSize()` in the Go pipeline.
local FRAME = 256
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

local gopher_mining = {
  filename = "__gopher__/graphics/gopher-mining.png",
  width = FRAME,
  height = FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.45,
  scale = SCALE,
  shift = SHIFT,
}

local gopher_mining_shadow = {
  filename = "__gopher__/graphics/gopher-mining-shadow.png",
  width = FRAME,
  height = FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.45,
  scale = SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

-- running_with_gun requires 18 or 40 directions. Factorio's 18-row layout
-- sweeps N -> E -> S over the east half-circle; the engine mirrors those rows
-- for west-facing combinations. A separate flipped shadow keeps the west body
-- silhouette while preserving the eastward cast direction.
-- Two stripes hold 16 and 2 direction rows respectively.
local GUN_FRAME = FRAME
local GUN_SCALE = SCALE
local function animation_stripes(name, frame_count, shadow)
  local suffix = shadow and "-shadow" or ""
  return {
    {
      filename = "__gopher__/graphics/" .. name .. "-1" .. suffix .. ".png",
      width_in_frames = frame_count,
      height_in_frames = 16,
    },
    {
      filename = "__gopher__/graphics/" .. name .. "-2" .. suffix .. ".png",
      width_in_frames = frame_count,
      height_in_frames = 2,
    },
  }
end

local function gun_stripes(name, shadow)
  return animation_stripes(name .. "-running-with-gun", 8, shadow)
end

local function flipped_gun_shadow_stripes(name)
  return animation_stripes(name .. "-running-with-gun-flipped", 8, true)
end

local gopher_idle_with_gun = {
  filename = "__gopher__/graphics/gopher-idle-with-gun.png",
  width = GUN_FRAME,
  height = GUN_FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = GUN_SCALE,
  shift = SHIFT,
}

local gopher_idle_with_gun_shadow = {
  filename = "__gopher__/graphics/gopher-idle-with-gun-shadow.png",
  width = GUN_FRAME,
  height = GUN_FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = GUN_SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

local gopher_running_with_gun = {
  stripes = gun_stripes("gopher", false),
  width = GUN_FRAME,
  height = GUN_FRAME,
  frame_count = 8,
  direction_count = 18,
  animation_speed = 0.5,
  scale = GUN_SCALE,
  shift = SHIFT,
}

local gopher_running_with_gun_shadow = {
  stripes = gun_stripes("gopher", true),
  width = GUN_FRAME,
  height = GUN_FRAME,
  frame_count = 8,
  direction_count = 18,
  animation_speed = 0.5,
  scale = GUN_SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

local gopher_flipped_running_with_gun_shadow = {
  stripes = flipped_gun_shadow_stripes("gopher"),
  width = GUN_FRAME,
  height = GUN_FRAME,
  frame_count = 8,
  direction_count = 18,
  animation_speed = 0.5,
  scale = GUN_SCALE,
  shift = SHIFT,
  draw_as_shadow = true,
}

-- Knight gopher for mech-armour flight (Space Age).
local KNIGHT_FRAME = FRAME
local KNIGHT_SCALE = SCALE
local HOVER_FRAMES = 5
-- The authored knight feet sit three rendered pixels above the normal gopher
-- baseline. Offset every flight frame together so takeoff starts in place.
local KNIGHT_SHIFT = {0, -0.40625}
local LANDING_SEQUENCE = {
  16, 15, 14, 13, 12, 11, 10, 9,
  8, 7, 6, 5, 4, 3, 2, 1,
}
local LEFT_STEP_FRAME = 6
local RIGHT_STEP_FRAME = 2

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

local knight_idle_with_gun = {
  filename = "__gopher__/graphics/knight-idle-with-gun.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_idle_with_gun_shadow = {
  filename = "__gopher__/graphics/knight-idle-with-gun-shadow.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 1,
  direction_count = 8,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local knight_mining = {
  filename = "__gopher__/graphics/knight-mining.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
  animation_speed = 0.45,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_mining_shadow = {
  filename = "__gopher__/graphics/knight-mining-shadow.png",
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 8,
  line_length = 8,
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
  stripes = gun_stripes("knight", false),
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 18,
  animation_speed = 0.5,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_running_with_gun_shadow = {
  stripes = gun_stripes("knight", true),
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 18,
  animation_speed = 0.5,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local knight_flipped_running_with_gun_shadow = {
  stripes = flipped_gun_shadow_stripes("knight"),
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = 8,
  direction_count = 18,
  animation_speed = 0.5,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
  draw_as_shadow = true,
}

local knight_flying_with_gun = {
  stripes = animation_stripes("knight-flying-with-gun", HOVER_FRAMES, false),
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = HOVER_FRAMES,
  direction_count = 18,
  animation_speed = 0.2,
  scale = KNIGHT_SCALE,
  shift = KNIGHT_SHIFT,
}

local knight_flying_with_gun_shadow = {
  stripes = animation_stripes("knight-flying-with-gun", HOVER_FRAMES, true),
  width = KNIGHT_FRAME,
  height = KNIGHT_FRAME,
  frame_count = HOVER_FRAMES,
  direction_count = 18,
  animation_speed = 0.2,
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
local knight_hover = knight_flight_animation("hover", HOVER_FRAMES, 0.2)

local gopher_corpse = {
  layers = {
    {
      filename = "__gopher__/graphics/gopher-corpse.png",
      width = FRAME,
      height = FRAME,
      frame_count = 2,
      line_length = 2,
      scale = SCALE,
      shift = SHIFT,
      usage = "player",
    },
    {
      filename = "__gopher__/graphics/gopher-corpse-shadow.png",
      width = FRAME,
      height = FRAME,
      frame_count = 2,
      line_length = 2,
      scale = SCALE,
      shift = SHIFT,
      usage = "player",
      draw_as_shadow = true,
    },
  },
}

local knight_corpse = {
  layers = {
    {
      filename = "__gopher__/graphics/knight-corpse.png",
      width = KNIGHT_FRAME,
      height = KNIGHT_FRAME,
      frame_count = 2,
      line_length = 2,
      scale = KNIGHT_SCALE,
      shift = KNIGHT_SHIFT,
      usage = "player",
    },
    {
      filename = "__gopher__/graphics/knight-corpse-shadow.png",
      width = KNIGHT_FRAME,
      height = KNIGHT_FRAME,
      frame_count = 2,
      line_length = 2,
      scale = KNIGHT_SCALE,
      shift = KNIGHT_SHIFT,
      usage = "player",
      draw_as_shadow = true,
    },
  },
}

local function has_armour(armour_set, armour_name)
  for _, name in ipairs(armour_set.armors or {}) do
    if name == armour_name then
      return true
    end
  end
  return false
end

local function reskin_corpses()
  local corpse = data.raw["character-corpse"]["character-corpse"]
  local pictures = corpse.pictures
  -- AnimationVariations may be a direct Animation, a sheet/sheets wrapper, or
  -- an array. Only the array form can safely accept our extra mech variation.
  if not pictures or (next(pictures) and not pictures[1]) then
    pictures = {}
    rawset(corpse, "pictures", pictures)
  end
  rawset(corpse, "picture", nil)
  rawset(corpse, "water_reflection", nil)
  rawset(pictures, 1, gopher_corpse)

  local mapping = corpse.armor_picture_mapping
  if not mapping then
    return
  end
  for armour_name in pairs(mapping) do
    if armour_name ~= "mech-armor" then
      rawset(mapping, armour_name, 1)
    end
  end
  if mapping["mech-armor"] then
    table.insert(pictures, knight_corpse)
    rawset(mapping, "mech-armor", #pictures)
  end
end

local function use_gopher(armour_set)
  armour_set.mining_with_tool = { layers = { gopher_mining, gopher_mining_shadow } }
  armour_set.mining_with_tool_particles_animation_positions = {5}
  armour_set.idle = { layers = { gopher_8dir, gopher_8dir_shadow } }
  armour_set.idle_with_gun = {
    layers = { gopher_idle_with_gun, gopher_idle_with_gun_shadow }
  }
  armour_set.running = { layers = { gopher_running, gopher_running_shadow } }
  armour_set.running_with_gun = {
    layers = { gopher_running_with_gun, gopher_running_with_gun_shadow }
  }
  armour_set.flipped_shadow_running_with_gun = {
    layers = { gopher_flipped_running_with_gun_shadow }
  }
end

local function use_knight(armour_set)
  armour_set.idle = { layers = { knight_idle, knight_idle_shadow } }
  armour_set.idle_with_gun = {
    layers = { knight_idle_with_gun, knight_idle_with_gun_shadow }
  }
  armour_set.mining_with_tool = { layers = { knight_mining, knight_mining_shadow } }
  armour_set.mining_with_tool_particles_animation_positions = {5}
  armour_set.running = { layers = { knight_running, knight_running_shadow } }
  armour_set.running_with_gun = {
    layers = { knight_running_with_gun, knight_running_with_gun_shadow }
  }
  armour_set.flipped_shadow_running_with_gun = {
    layers = { knight_flipped_running_with_gun_shadow }
  }

  if armour_set.take_off then
    armour_set.take_off = knight_take_off
  end
  if armour_set.landing then
    armour_set.landing = knight_landing
  end
  if armour_set.idle_with_gun_in_air then
    armour_set.idle_with_gun_in_air = knight_hover
  end
  if armour_set.idle_in_air then
    armour_set.idle_in_air = knight_hover
  end
  if armour_set.flying then
    armour_set.flying = knight_hover
  end
  if armour_set.flying_with_gun then
    armour_set.flying_with_gun = {
      layers = { knight_flying_with_gun, knight_flying_with_gun_shadow }
    }
  end
  -- The knight has no thruster vents.
  armour_set.smoke_in_air = nil
end

local character = data.raw.character.character
rawset(character, "water_reflection", nil)
rawset(character, "running_sound_animation_positions", {LEFT_STEP_FRAME})
rawset(character, "moving_sound_animation_positions", {RIGHT_STEP_FRAME})
rawset(character, "left_footprint_frames", {LEFT_STEP_FRAME})
rawset(character, "right_footprint_frames", {RIGHT_STEP_FRAME})

for _, armour_set in ipairs(character.animations) do
  if has_armour(armour_set, "mech-armor") then
    use_knight(armour_set)
  else
    use_gopher(armour_set)
  end
end

reskin_corpses()
