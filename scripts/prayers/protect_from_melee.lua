--- Handles activating the protect from melee prayer.
-- @param player The player who activated the prayer
-- @param activate true if the prayer should be activated, false if deactivated
function prayer_protect_from_melee(player, activate)
    local setting_id = 97
    
    if activate then
        local ok = skill_level_minimum(player, SKILL_PRAYER, 43, "You need prayer level 43 to use this prayer.")
        if not ok then
            player:interface_setting(setting_id, 0)
            return
        end

        -- disable conflicting prayers
        prayer_protect_from_missiles(player, false)
        prayer_protect_from_magic(player, false)
        prayer_retribution(player, false)
        prayer_redemption(player, false)
        prayer_smite(player, false)

        -- TODO: add buffs, effects, etc.
        player:overhead_icon(OVERHEAD_PROTECT_FROM_MELEE)

        player:activate_prayer(PRAYER_PROTECT_FROM_MELEE, 12)
        player:interface_setting(setting_id, 1)
    else
        -- TODO: remove buffs, effects, etc.
        player:overhead_icon(OVERHEAD_NONE)

        player:deactivate_prayer(PRAYER_PROTECT_FROM_MELEE)
        player:interface_setting(setting_id, 0)
    end
end
