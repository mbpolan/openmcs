-- map of quest identifiers to interface text
QUESTS_IDS_TO_INTERFACES = {
    [QUEST_ID_COOKS_ASSISTANT] = 7333,
}

--- Handles an action performed on the quest list interface.
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_638_on_action(player, interface)
    -- TODO
end

--- Handles updating the quest list interface.
-- @param player The player
function interface_638_on_update(player)
    for quest_id, interface_id in pairs(QUESTS_IDS_TO_INTERFACES) do
        local status = player:quest_status(quest_id)

        -- set an appropriate color for each quest based on its status
        if status == QUEST_STATUS_NOT_STARTED then
            player:interface_color(interface_id, 31, 0, 0)
        elseif status == QUEST_STATUS_IN_PROGRESS then
            player:interface_color(interface_id, 31, 31, 0)
        elseif status == QUEST_STATUS_COMPLETED then
            player:interface_color(interface_id, 0, 31, 0)
        end
    end
end
