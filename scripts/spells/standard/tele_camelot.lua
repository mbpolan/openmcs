--- Casts the Teleport to Camelot spell
-- @param player The player casting the spell
function spell_teleport_camelot(player)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 45, "You need magic level 45 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune and 5 air runes
    teleport_standard(player, 2757, 3478, 0, ITEM_LAW_RUNE, 1, ITEM_AIR_RUNE, 5)
end
