-- Test spec for mod/data-updates.lua. Stubs the bare minimum of Factorio's
-- data table that data-updates.lua mutates, sources the file, then asserts
-- the resulting per-armour-set wiring matches the intended layer layout.

describe("data-updates", function()
  -- Build a fresh `data` table for each test so mutation can't leak between
  -- specs. The helper returns the lone armour set so assertions stay terse.
  local function load_data_updates(fields)
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
    _G.data = {
      raw = {
        character = {
          character = { animations = { armour_set } },
        },
      },
    }
    -- Reset the package cache so repeated dofile/require land cleanly.
    dofile("mod/data-updates.lua")
    return armour_set
  end

  it("wires idle with body and shadow layers", function()
    local set = load_data_updates()
    assert.are.equal(2, #set.idle.layers)
    assert.are.equal("__gopher__/graphics/gopher-8dir.png", set.idle.layers[1].filename)
    assert.is_true(set.idle.layers[2].draw_as_shadow)
    assert.are.equal("__gopher__/graphics/gopher-shadow-8dir.png", set.idle.layers[2].filename)
  end)

  it("wires idle_with_gun the same as idle", function()
    local set = load_data_updates()
    assert.are.same(set.idle, set.idle_with_gun)
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

  it("running_with_gun uses 18 directions in a 6x3 grid", function()
    local set = load_data_updates()
    local body = set.running_with_gun.layers[1]
    assert.are.equal("__gopher__/graphics/gopher-running-with-gun.png", body.filename)
    assert.are.equal(18, body.direction_count)
    assert.are.equal(6, body.line_length)
    assert.is_true(set.running_with_gun.layers[2].draw_as_shadow)
  end)

  it("clears flipped_shadow_running_with_gun to avoid duplicate gopher", function()
    local set = load_data_updates()
    assert.is_nil(set.flipped_shadow_running_with_gun)
  end)

  it("mining_with_tool uses the single south-facing pose", function()
    local set = load_data_updates()
    local body = set.mining_with_tool.layers[1]
    assert.are.equal("__gopher__/graphics/gopher-s.png", body.filename)
    assert.are.equal(1, body.direction_count)
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
    assert.are.same(set.idle, set.idle_with_gun)
    assert.are.equal(
      "__gopher__/graphics/knight-idle.png",
      set.mining_with_tool.layers[1].filename
    )
    assert.are.equal(27, set.mining_with_tool.layers[1].repeat_count)
    assert.are.equal(0.45, set.mining_with_tool.layers[1].animation_speed)
    assert.are.equal(
      "__gopher__/graphics/knight-running.png",
      set.running.layers[1].filename
    )
    assert.are.equal(8, set.running.layers[1].frame_count)
    assert.are.equal(
      "__gopher__/graphics/knight-running-with-gun.png",
      set.running_with_gun.layers[1].filename
    )
    assert.are.equal(18, set.running_with_gun.layers[1].direction_count)
    assert.is_nil(set.flipped_shadow_running_with_gun)

    local take_off = set.take_off
    assert.are.equal(2, #take_off.layers)
    assert.are.equal("__gopher__/graphics/knight-take-off.png", take_off.layers[1].filename)
    assert.are.equal(512, take_off.layers[1].width)
    assert.are.equal(16, take_off.layers[1].frame_count)
    assert.are.equal(16, take_off.layers[1].line_length)
    assert.are.equal(8, take_off.layers[1].direction_count)
    assert.are.equal(0.6, take_off.layers[1].animation_speed)
    assert.are.equal(0.075, take_off.layers[1].scale)
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
      { label = "heavy armour", armors = { "heavy-armor" } },
      { label = "modular armour", armors = { "modular-armor" } },
      { label = "power armour", armors = { "power-armor" } },
      { label = "power armour MK2", armors = { "power-armor-mk2" } },
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

  it("does not assign knight flight to a different flying armour", function()
    local stub = { layers = { "third-party-flight" } }
    local smoke = { "third-party-smoke" }
    local set = load_data_updates({
      armors = { "jetpack-armor" },
      take_off = stub,
      landing = stub,
      idle_with_gun_in_air = stub,
      smoke_in_air = smoke,
    })
    assert.are.equal("__gopher__/graphics/gopher-8dir.png", set.idle.layers[1].filename)
    assert.are.same(stub, set.take_off)
    assert.are.same(stub, set.landing)
    assert.are.same(stub, set.idle_with_gun_in_air)
    assert.are.same(smoke, set.smoke_in_air)
  end)

  it("reskins every armour tier, not just the first", function()
    -- Add a second armour set and re-run the file to confirm the loop walks
    -- the whole array.
    local sets = { {}, {} }
    _G.data = {
      raw = {
        character = {
          character = { animations = sets },
        },
      },
    }
    dofile("mod/data-updates.lua")
    for i, set in ipairs(sets) do
      assert.is_not_nil(set.idle, "armour set " .. i .. " missing idle wiring")
      assert.is_not_nil(set.running, "armour set " .. i .. " missing running wiring")
    end
  end)
end)
