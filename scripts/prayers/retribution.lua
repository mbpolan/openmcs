--- Handles activating the retribution prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_retribution(player, activate)
    local setting_id = 98
    
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 46, "You need prayer level 46 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return
        end

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_RETRIBUTION, 3)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_RETRIBUTION)
        player:interface_setting(setting_id, 0)
    end
end
