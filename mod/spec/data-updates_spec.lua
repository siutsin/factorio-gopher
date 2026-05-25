-- Test spec for mod/data-updates.lua. Stubs the bare minimum of Factorio's
-- data table that data-updates.lua mutates, sources the file, then asserts
-- the resulting per-armour-set wiring matches the intended layer layout.

describe("data-updates", function()
  -- Build a fresh `data` table for each test so mutation can't leak between
  -- specs. The helper returns the lone armour set so assertions stay terse.
  local function load_data_updates()
    local armour_set = {
      idle = nil,
      idle_with_gun = nil,
      running = nil,
      running_with_gun = nil,
      flipped_shadow_running_with_gun = { layers = { "stub" } },
      mining_with_tool = nil,
    }
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
