--- Handles activating the steel skin prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_steel_skin(player, activate)
    local setting_id = 92
    
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 28, "You need prayer level 28 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return
        end

        -- disable conflicting prayers
        prayer_thick_skin(player, false)
        prayer_rock_skin(player, false)

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_STEEL_SKIN, 12)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_STEEL_SKIN)
        player:interface_setting(setting_id, 0)
    end
end
