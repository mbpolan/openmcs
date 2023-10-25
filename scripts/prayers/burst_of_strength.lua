--- Handles activating the burst of strength prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_burst_of_strength(player, activate)
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 4, "You need prayer level 4 to use this prayer.")
        if not ok then
            player:interface_setting(84, 0)
            return
        end

        -- disable conflicting prayers
        prayer_superhuman_strength(player, false)
        prayer_ultimate_strength(player, false)

        -- TODO: add buffs, effects, etc.

        player:activate_prayer(PRAYER_BURST_OF_STRENGTH, 3)
        player:interface_setting(84, 1)
    else
        -- TODO: remove buffs, effects, etc.

        player:deactivate_prayer(PRAYER_BURST_OF_STRENGTH)
        player:interface_setting(84, 0)
    end
end
