--- Handles activating the smite prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_smite(player, activate)
    local setting_id = 100

    if activate then
        if not player:has_membership() then
            player:server_message("You need to be a member to use this prayer.")
            player:interface_setting(setting_id, 0)
            return
        end

        local ok = skill_level_minimum(player, SKILL_PRAYER, 52, "You need prayer level 52 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return
        end

        -- TODO: add buffs, effects, etc.
        player:overhead_icon(OVERHEAD_SMITE)

        player:activate_prayer(PRAYER_SMITE, 18)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.
        player:overhead_icon(OVERHEAD_NONE)

        player:deactivate_prayer(PRAYER_SMITE)
        player:interface_setting(setting_id, 0)
    end
end
