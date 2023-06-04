--- Returns the quest name for this quest.
-- @return The quest name
function quest_1_name()
    return "Cook's Assistant"
end

--- Returns the quest journal entries for this quest.
-- @return A table of journal entry identifiers to text.
function quest_1_entries()
    return {
        [0] = "@dbl@Go see the cook at the castle",
        [1] = "@dbl@Find the ingredients he's asking for",
        [2] = "@dbl@Bake a cake using the flour and bowl",
    }
end