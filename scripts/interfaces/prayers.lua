-- map of prayer identifiers for interface buttons
PRAYER_IDS_TO_INTERFACES = {
    [PRAYER_THICK_SKIN] = 5609
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
        return
    end

    if prayer_id == PRAYER_THICK_SKIN then
        prayer_thick_skin(player, activate)
    else
        player:server_message("This prayer is not yet available!")
    end
end

--- Handles updating the prayer interface.
-- @param player The player
function interface_5608_on_update(player)
    -- synchronize all prayers with the player's set of active prayers
    prayer_thick_skin(player, player:has_prayer_active(PRAYER_THICK_SKIN))
end