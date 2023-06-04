-------------------------------------
-- Quests
-------------------------------------

-- identifiers for quests
QUEST_ID_COOKS_ASSISTANT = 1

-------------------------------------

--- Returns the name for a quest.
-- @param quest_id The quest identifier
-- @return The name of the quest
function quest_get_name(quest_id)
    local func_name = "quest_" .. quest_id .. "_name"

    return _G[func_name]()
end

--- Returns the journal entries for a quest.
-- @param quest_id The quest identifier
-- @return A table of journal entry identifiers to text
function quest_get_entries(quest_id)
    local func_name = "quest_" .. quest_id .. "_entries"

    return _G[func_name]()
end