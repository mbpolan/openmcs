--- Handles activating the clarity of thought prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_clarity_of_thought(player, activate)
    local setting_id = 85
    
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 7, "You need prayer level 7 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return
        end

        -- disable conflicting prayers
        prayer_improved_reflexes(player, false)
        prayer_incredible_reflexes(player, false)

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_CLARITY_OF_THOUGHT, 3)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_CLARITY_OF_THOUGHT)
        player:interface_setting(setting_id, 0)
    end
end
