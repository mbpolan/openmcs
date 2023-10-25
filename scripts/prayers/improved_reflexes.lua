--- Handles activating the improved reflexes prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_improved_reflexes(player, activate)
    local setting_id = 88

    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 16, "You need prayer level 16 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return true
        end

        -- disable conflicting prayers
        prayer_clarity_of_thought(player, false)
        prayer_incredible_reflexes(player, false)

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_IMPROVED_REFLEXES, 6)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_IMPROVED_REFLEXES)
        player:interface_setting(setting_id, 0)
    end
end
