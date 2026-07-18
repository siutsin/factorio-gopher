-- Test spec for mod/data-updates.lua. Stubs the bare minimum of Factorio's
-- data table that data-updates.lua mutates, sources the file, then asserts
-- the resulting per-armour-set wiring matches the intended layer layout.

describe("data-updates", function()
  -- Build a fresh `data` table for each test so mutation can't leak between
  -- specs. The helper returns the lone armour set so assertions stay terse.
  local function load_data_updates(fields, corpse_fields)
    local armour_set = {
      idle = nil,
      idle_with_gun = nil,
      running = nil,
      running_with_gun = nil,
      flipped_shadow_running_with_gun = { layers = { "stub" } },
      mining_with_tool = nil,
    }
    for key, value in pairs(fields or {}) do
      armour_set[key] = value
    end
    local corpse = {
      pictures = { "base", "heavy", "power", "mech" },
      water_reflection = { pictures = { "engineer-corpse" } },
      armor_picture_mapping = {
        ["light-armor"] = 1,
        ["heavy-armor"] = 2,
        ["power-armor"] = 3,
        ["mech-armor"] = 4,
        ["third-party-armor"] = 2,
      },
    }
    for key, value in pairs(corpse_fields or {}) do
      if value == false then
        corpse[key] = nil
      else
        corpse[key] = value
      end
    end
    local character = {
      animations = { armour_set },
      water_reflection = { pictures = { "engineer" } },
    }
    _G.data = {
      raw = {
        character = {
          character = character,
        },
        ["character-corpse"] = {
          ["character-corpse"] = corpse,
        },
      },
    }
    -- Reset the package cache so repeated dofile/require land cleanly.
    dofile("mod/data-updates.lua")
    return armour_set, corpse, character
  end

  it("wires idle with body and shadow layers", function()
    local set = load_data_updates()
    assert.are.equal(2, #set.idle.layers)
    assert.are.equal("__gopher__/graphics/gopher-8dir.png", set.idle.layers[1].filename)
    assert.is_true(set.idle.layers[2].draw_as_shadow)
    assert.are.equal("__gopher__/graphics/gopher-shadow-8dir.png", set.idle.layers[2].filename)
  end)

  it("wires idle_with_gun to dedicated armed sheets", function()
    local set = load_data_updates()
    assert.are.equal(
      "__gopher__/graphics/gopher-idle-with-gun.png",
      set.idle_with_gun.layers[1].filename
    )
    assert.are.equal(8, set.idle_with_gun.layers[1].direction_count)
    assert.is_true(set.idle_with_gun.layers[2].draw_as_shadow)
  end)

  it("running uses the 8-frame run-cycle sheet", function()
    local set = load_data_updates()
    local body = set.running.layers[1]
    assert.are.equal("__gopher__/graphics/gopher-running.png", body.filename)
    assert.are.equal(8, body.frame_count)
    assert.are.equal(8, body.direction_count)
    assert.are.equal(8, body.line_length)
    assert.is_true(set.running.layers[2].draw_as_shadow)
  end)

  it("running_with_gun uses an eight-frame cycle across two stripes", function()
    local set = load_data_updates()
    local body = set.running_with_gun.layers[1]
    assert.are.equal(8, body.frame_count)
    assert.are.equal(18, body.direction_count)
    assert.are.equal(2, #body.stripes)
    assert.are.equal(8, body.stripes[1].width_in_frames)
    assert.are.equal(16, body.stripes[1].height_in_frames)
    assert.are.equal(2, body.stripes[2].height_in_frames)
    assert.is_true(set.running_with_gun.layers[2].draw_as_shadow)
  end)

  it("wires an east-cast shadow for engine-mirrored armed rows", function()
    local set = load_data_updates()
    local shadow = set.flipped_shadow_running_with_gun.layers[1]
    assert.are.equal(18, shadow.direction_count)
    assert.are.equal(
      "__gopher__/graphics/gopher-running-with-gun-flipped-1-shadow.png",
      shadow.stripes[1].filename
    )
    assert.is_true(shadow.draw_as_shadow)
  end)

  it("aligns sounds and footprints with the eight-frame run cycle", function()
    local _, _, character = load_data_updates()
    assert.are.same({ 6 }, character.running_sound_animation_positions)
    assert.are.same({ 2 }, character.moving_sound_animation_positions)
    assert.are.same({ 6 }, character.left_footprint_frames)
    assert.are.same({ 2 }, character.right_footprint_frames)
  end)

  it("removes vanilla engineer water reflections", function()
    local _, corpse, character = load_data_updates()
    assert.is_nil(character.water_reflection)
    assert.is_nil(corpse.water_reflection)
  end)

  it("uses knight sheets for mech-armour ground and flight", function()
    local stub = { layers = { "vanilla" } }
    local set = load_data_updates({
      armors = { "mech-armor" },
      take_off = stub,
      landing = stub,
      idle_with_gun_in_air = stub,
      smoke_in_air = { "thruster-smoke" },
    })

    assert.are.equal("__gopher__/graphics/knight-idle.png", set.idle.layers[1].filename)
    assert.are.equal(
      "__gopher__/graphics/knight-idle-shadow.png",
      set.idle.layers[2].filename
    )
    assert.are.equal(
      "__gopher__/graphics/knight-idle-with-gun.png",
      set.idle_with_gun.layers[1].filename
    )
    assert.is_true(set.idle_with_gun.layers[2].draw_as_shadow)
    assert.are.equal(
      "__gopher__/graphics/knight-running.png",
      set.running.layers[1].filename
    )
    assert.are.equal(8, set.running.layers[1].frame_count)
    assert.are.equal(
      "__gopher__/graphics/knight-running-with-gun-1.png",
      set.running_with_gun.layers[1].stripes[1].filename
    )
    assert.are.equal(18, set.running_with_gun.layers[1].direction_count)
    assert.are.equal(8, set.running_with_gun.layers[1].frame_count)
    assert.are.equal(
      "__gopher__/graphics/knight-running-with-gun-flipped-1-shadow.png",
      set.flipped_shadow_running_with_gun.layers[1].stripes[1].filename
    )

    local take_off = set.take_off
    assert.are.equal(2, #take_off.layers)
    assert.are.equal("__gopher__/graphics/knight-take-off.png", take_off.layers[1].filename)
    assert.are.equal(256, take_off.layers[1].width)
    assert.are.equal(16, take_off.layers[1].frame_count)
    assert.are.equal(16, take_off.layers[1].line_length)
    assert.are.equal(8, take_off.layers[1].direction_count)
    assert.are.equal(0.6, take_off.layers[1].animation_speed)
    assert.are.equal(0.15, take_off.layers[1].scale)
    assert.are.same({ 0, -0.40625 }, take_off.layers[1].shift)
    assert.are.equal("player", take_off.layers[1].usage)
    assert.are.equal(
      "__gopher__/graphics/knight-take-off-shadow.png",
      take_off.layers[2].filename
    )
    assert.is_true(take_off.layers[2].draw_as_shadow)

    local landing = set.landing
    assert.are.equal("__gopher__/graphics/knight-take-off.png", landing.layers[1].filename)
    assert.are.same(
      { 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1 },
      landing.layers[1].frame_sequence
    )
    assert.are.same(landing.layers[1].frame_sequence, landing.layers[2].frame_sequence)

    local hover = set.idle_with_gun_in_air
    assert.are.equal("__gopher__/graphics/knight-hover.png", hover.layers[1].filename)
    assert.are.equal(5, hover.layers[1].frame_count)
    assert.are.equal(5, hover.layers[1].line_length)
    assert.are.equal(0.2, hover.layers[1].animation_speed)
    assert.is_nil(hover.layers[1].frame_sequence)
    assert.is_nil(set.smoke_in_air)
  end)

  it("preserves smoke on armour sets without in-air animations", function()
    local smoke = { "third-party-smoke" }
    local set = load_data_updates({ smoke_in_air = smoke })
    assert.are.same(smoke, set.smoke_in_air)
  end)

  it("keeps every non-mech armour tier on the default gopher", function()
    local cases = {
      { label = "no armour" },
      { label = "light armour", armors = { "light-armor" } },
      { label = "heavy and modular armour", armors = { "heavy-armor", "modular-armor" } },
      { label = "power armour", armors = { "power-armor", "power-armor-mk2" } },
    }
    for _, case in ipairs(cases) do
      local set = load_data_updates({ armors = case.armors })
      assert.are.equal(
        "__gopher__/graphics/gopher-8dir.png",
        set.idle.layers[1].filename,
        case.label
      )
      assert.are.equal(
        "__gopher__/graphics/gopher-running.png",
        set.running.layers[1].filename,
        case.label
      )
    end
  end)

  it("does not invent optional airborne states", function()
    local set = load_data_updates({ armors = { "mech-armor" } })
    assert.is_nil(set.idle_in_air)
    assert.is_nil(set.flying)
    assert.is_nil(set.flying_with_gun)
  end)

  it("reskins every armour set, not just the first", function()
    -- Add a second armour set and re-run the file to confirm the loop walks
    -- the whole array.
    local sets = { {}, { armors = { "heavy-armor", "modular-armor" } } }
    _G.data = {
      raw = {
        character = {
          character = { animations = sets },
        },
        ["character-corpse"] = {
          ["character-corpse"] = {
            pictures = { "base" },
            armor_picture_mapping = {},
          },
        },
      },
    }
    dofile("mod/data-updates.lua")
    for i, set in ipairs(sets) do
      assert.is_not_nil(set.idle, "armour set " .. i .. " missing idle wiring")
      assert.is_not_nil(set.running, "armour set " .. i .. " missing running wiring")
    end
  end)

  it("replaces default and mapped corpses without deleting other variations", function()
    local _, corpse = load_data_updates()
    local gopher = corpse.pictures[1]
    assert.are.equal("__gopher__/graphics/gopher-corpse.png", gopher.layers[1].filename)
    assert.are.equal(2, gopher.layers[1].frame_count)
    assert.are.equal(2, gopher.layers[1].line_length)
    assert.is_true(gopher.layers[2].draw_as_shadow)
    assert.are.equal(1, corpse.armor_picture_mapping["light-armor"])
    assert.are.equal(1, corpse.armor_picture_mapping["third-party-armor"])

    assert.are.equal(5, #corpse.pictures)
    local knight = corpse.pictures[5]
    assert.are.equal("__gopher__/graphics/knight-corpse.png", knight.layers[1].filename)
    assert.are.equal(256, knight.layers[1].width)
    assert.are.equal(5, corpse.armor_picture_mapping["mech-armor"])
  end)

  it("does not add a knight corpse without mech armour", function()
    local mapping = { ["light-armor"] = 1 }
    local _, corpse = load_data_updates(nil, {
      pictures = { "base" },
      armor_picture_mapping = mapping,
    })
    assert.are.equal(1, #corpse.pictures)
    assert.is_nil(corpse.armor_picture_mapping["mech-armor"])
  end)

  it("normalizes a singular corpse picture", function()
    local mapping = { ["light-armor"] = 1 }
    local _, corpse = load_data_updates(nil, {
      picture = { layers = { "single" } },
      pictures = false,
      armor_picture_mapping = mapping,
    })

    assert.is_nil(corpse.picture)
    assert.are.equal(
      "__gopher__/graphics/gopher-corpse.png",
      corpse.pictures[1].layers[1].filename
    )
    assert.are.equal(1, corpse.armor_picture_mapping["light-armor"])
  end)

  for label, pictures in pairs({
    animation = { filename = "character-corpse.png", width = 64, height = 64 },
    sheet = { sheet = { filename = "character-corpses.png", variation_count = 4 } },
    sheets = { sheets = { { filename = "character-corpses.png", variation_count = 4 } } },
  }) do
    it("normalizes " .. label .. " corpse variations", function()
      local _, corpse = load_data_updates(nil, { pictures = pictures })

      assert.is_nil(corpse.pictures.filename)
      assert.is_nil(corpse.pictures.sheet)
      assert.is_nil(corpse.pictures.sheets)
      assert.are.equal(2, #corpse.pictures)
      assert.are.equal(
        "__gopher__/graphics/gopher-corpse.png",
        corpse.pictures[1].layers[1].filename
      )
      assert.are.equal(
        "__gopher__/graphics/knight-corpse.png",
        corpse.pictures[2].layers[1].filename
      )
    end)
  end

  it("supports a corpse without an armour picture mapping", function()
    local _, corpse = load_data_updates(nil, {
      pictures = { "base" },
      armor_picture_mapping = false,
    })

    assert.are.equal(
      "__gopher__/graphics/gopher-corpse.png",
      corpse.pictures[1].layers[1].filename
    )
    assert.is_nil(corpse.armor_picture_mapping)
  end)
end)
