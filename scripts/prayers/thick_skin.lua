--- Handles activating the thick spin prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_thick_skin(player, activate)
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 1, "You need prayer level 1 to use this prayer.")
        if not ok then
            return true
        end

        player:activate_prayer(PRAYER_THICK_SKIN, 3)
    else
        player:deactivate_prayer(PRAYER_THICK_SKIN)
    end
end
