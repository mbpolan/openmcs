-- initialize the interface identifiers for each quest journal entry
INTERFACE_QUEST_LOG_ENTRY_LINE_IDS = { 8145 }
for i = 8147, 8195, 1 do
    table.insert(INTERFACE_QUEST_LOG_ENTRY_LINE_IDS, i)
end
for i = 12174, 12223, 1 do
    table.insert(INTERFACE_QUEST_LOG_ENTRY_LINE_IDS, i)
end

--- Shows the quest journal interface and its contents.
-- @param player The player to show the interface to
-- @param quest_name The name of the quest
-- @param entries A table of quest entry text mapping to a boolean if that entry has been completed
function interface_quest_log_show(player, quest_name, entries)
    -- show the primary interface
    player:show_interface(8134)

    -- update the quest name line
    player:interface_text(8144, "@dre@" .. quest_name)

    -- update journal entries starting from the first line
    local text_interface_idx = 1
    for k, v in pairs(entries) do
        local entry_interface = INTERFACE_QUEST_LOG_ENTRY_LINE_IDS[text_interface_idx]
        player:interface_text(entry_interface, v)

        text_interface_idx = text_interface_idx + 1
    end

    -- clear the remaining entry interfaces
    for i = text_interface_idx, 100, 1 do
        local entry_interface = INTERFACE_QUEST_LOG_ENTRY_LINE_IDS[i]
        if entry_interface == nil then
            break
        end

        player:interface_text(entry_interface, "")
    end
end