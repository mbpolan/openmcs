--- Handles activating the thick spin prayer.
-- @param player The player who activated the prayer
function prayer_thick_skin(player)
    player:toggle_prayer(PRAYER_THICK_SKIN, 3)
end
