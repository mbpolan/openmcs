-- map of prayer identifiers for interface buttons
PRAYER_IDS_TO_INTERFACES = {
    [PRAYER_THICK_SKIN] = 5609,
    [PRAYER_BURST_OF_STRENGTH] = 5610,
    [PRAYER_CLARITY_OF_THOUGHT] = 5611,
    [PRAYER_ROCK_SKIN] = 5612,
    [PRAYER_SUPERHUMAN_STRENGTH] = 5613,
    [PRAYER_IMPROVED_REFLEXES] = 5614,
    [PRAYER_RAPID_RESTORE] = 5615,
    [PRAYER_RAPID_HEAL] = 5616,
    [PRAYER_PROTECT_ITEMS] = 5617,
    [PRAYER_STEEL_SKIN] = 5618,
    [PRAYER_ULTIMATE_STRENGTH] = 5619,
    [PRAYER_INCREDIBLE_REFLEXES] = 5620,
    [PRAYER_PROTECT_FROM_MAGE] = 5621,
    [PRAYER_PROTECT_FROM_MISSILES] = 5622,
    [PRAYER_PROTECT_FROM_MELEE] = 5623,
    [PRAYER_RETRIBUTION] = 683,
    [PRAYER_REDEMPTION] = 684,
    [PRAYER_SMITE] = 685
}

-- map of prayer button interfaces to prayer identifiers
PRAYER_INTERFACES_TO_IDS = {}
for k, v in pairs(PRAYER_IDS_TO_INTERFACES) do
    PRAYER_INTERFACES_TO_IDS[v] = k
end

--- Handles an action performed on the prayer interface.
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_5608_on_action(player, interface)
    local id = interface:id()

    -- find the prayer id that was clicked
    local prayer_id = PRAYER_INTERFACES_TO_IDS[id]
    if prayer_id == nil then
        player:server_message("This prayer is not yet available!")
        return
    end

    -- determine if this prayer is to be activated or deactivated
    local activate = player:has_prayer_active(prayer_id) == false

    -- ensure the player has at least one prayer point before activating a prayer
    if activate and player:stat_level(SKILL_PRAYER) == 0 then
        player:server_message("You need to recharge your Prayer at an altar.")
        activate = false
    end

    if prayer_id == PRAYER_THICK_SKIN then
        prayer_thick_skin(player, activate)
    elseif prayer_id == PRAYER_BURST_OF_STRENGTH then
        prayer_burst_of_strength(player, activate)
    elseif prayer_id == PRAYER_CLARITY_OF_THOUGHT then
        prayer_clarity_of_thought(player, activate)
    elseif prayer_id == PRAYER_ROCK_SKIN then
        prayer_rock_skin(player, activate)
    elseif prayer_id == PRAYER_SUPERHUMAN_STRENGTH then
        prayer_superhuman_strength(player, activate)
    elseif prayer_id == PRAYER_IMPROVED_REFLEXES then
        prayer_improved_reflexes(player, activate)
    elseif prayer_id == PRAYER_RAPID_RESTORE then
        prayer_rapid_restore(player, activate)
    elseif prayer_id == PRAYER_RAPID_HEAL then
        prayer_rapid_heal(player, activate)
    elseif prayer_id == PRAYER_PROTECT_ITEMS then
        prayer_protect_items(player, activate)
    elseif prayer_id == PRAYER_STEEL_SKIN then
        prayer_steel_skin(player, activate)
    elseif prayer_id == PRAYER_ULTIMATE_STRENGTH then
        prayer_ultimate_strength(player, activate)
    elseif prayer_id == PRAYER_INCREDIBLE_REFLEXES then
        prayer_incredible_reflexes(player, activate)
    elseif prayer_id == PRAYER_PROTECT_FROM_MAGE then
        prayer_protect_from_magic(player, activate)
    elseif prayer_id == PRAYER_PROTECT_FROM_MISSILES then
        prayer_protect_from_missiles(player, activate)
    elseif prayer_id == PRAYER_PROTECT_FROM_MELEE then
        prayer_protect_from_melee(player, activate)
    elseif prayer_id == PRAYER_RETRIBUTION then
        prayer_retribution(player, activate)
    elseif prayer_id == PRAYER_REDEMPTION then
        prayer_redemption(player, activate)
    elseif prayer_id == PRAYER_SMITE then
        prayer_smite(player, activate)
    else
        player:server_message("This prayer is not yet available!")
    end
end

--- Handles updating the prayer interface.
-- @param player The player
function interface_5608_on_update(player)
    -- synchronize all prayers with the player's set of active prayers
    prayer_thick_skin(player, player:has_prayer_active(PRAYER_THICK_SKIN))
    prayer_burst_of_strength(player, player:has_prayer_active(PRAYER_BURST_OF_STRENGTH))
    prayer_clarity_of_thought(player, player:has_prayer_active(PRAYER_CLARITY_OF_THOUGHT))
    prayer_rock_skin(player, player:has_prayer_active(PRAYER_ROCK_SKIN))
    prayer_superhuman_strength(player, player:has_prayer_active(PRAYER_SUPERHUMAN_STRENGTH))
    prayer_improved_reflexes(player, player:has_prayer_active(PRAYER_IMPROVED_REFLEXES))
    prayer_rapid_restore(player, player:has_prayer_active(PRAYER_RAPID_RESTORE))
    prayer_rapid_heal(player, player:has_prayer_active(PRAYER_RAPID_HEAL))
    prayer_protect_items(player, player:has_prayer_active(PRAYER_PROTECT_ITEMS))
    prayer_steel_skin(player, player:has_prayer_active(PRAYER_STEEL_SKIN))
    prayer_ultimate_strength(player, player:has_prayer_active(PRAYER_ULTIMATE_STRENGTH))
    prayer_incredible_reflexes(player, player:has_prayer_active(PRAYER_INCREDIBLE_REFLEXES))
    prayer_protect_from_magic(player, player:has_prayer_active(PRAYER_PROTECT_FROM_MAGE))
    prayer_protect_from_missiles(player, player:has_prayer_active(PRAYER_PROTECT_FROM_MISSILES))
    prayer_protect_from_melee(player, player:has_prayer_active(PRAYER_PROTECT_FROM_MELEE))
    prayer_retribution(player, player:has_prayer_active(PRAYER_RETRIBUTION))
    prayer_redemption(player, player:has_prayer_active(PRAYER_REDEMPTION))
    prayer_smite(player, player:has_prayer_active(PRAYER_SMITE))
end