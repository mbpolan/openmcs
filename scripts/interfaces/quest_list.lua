-- map of quest identifiers to interface text
QUESTS_IDS_TO_INTERFACES = {
    [QUEST_ID_COOKS_ASSISTANT] = 7333,
}

-- map of quest text interfaces to quest identifiers
QUEST_INTERFACES_TO_IDS = {}
for k, v in pairs(QUESTS_IDS_TO_INTERFACES) do
    QUEST_INTERFACES_TO_IDS[v] = k
end

--- Handles an action performed on the quest list interface.
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_638_on_action(player, interface)
    local id = interface:id()

    -- find the quest id based on the interface that was clicked
    local quest_id = QUEST_INTERFACES_TO_IDS[id]
    if quest_id == nil then
        player:server_message("This quest is not yet available!")
        return
    end

    -- get the quest name and entries
    local name = quest_get_name(quest_id)
    local entries = quest_get_entries(quest_id)

    interface_quest_log_show(player, name, entries)
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
