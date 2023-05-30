--- Casts the Teleport to Watchtower spell
-- @param player The player casting the spell
function spell_teleport_watchtower(player)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 58, "You need magic level 58 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes and 2 earth runes
    teleport_standard(player, 2546, 3113, 2, ITEM_LAW_RUNE, 2, ITEM_EARTH_RUNE, 2)
end
