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

    -- TODO: check if player meets level requirements
    -- TODO: activate prayer
end

--- Handles updating the prayer interface.
-- @param player The player
function interface_5608_on_update(player)
    -- TODO
end