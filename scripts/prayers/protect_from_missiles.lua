--- Handles activating the protect from missiles prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_protect_from_missiles(player, activate)
    local setting_id = 96
    
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 40, "You need prayer level 40 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return
        end

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_PROTECT_FROM_MISSILES, 12)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_PROTECT_FROM_MISSILES)
        player:interface_setting(setting_id, 0)
    end
end
