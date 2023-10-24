--- Handles activating the incredible reflexes prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_incredible_reflexes(player, activate)
    local setting_id = 94
    
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 34, "You need prayer level 34 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
           return
        end

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_INCREDIBLE_REFLEXES, 12)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_INCREDIBLE_REFLEXES)
        player:interface_setting(setting_id, 0)
    end
end
