--- Casts the Teleport to Lumbridge spell
-- @param player The player casting the spell
function spell_teleport_lumbridge(player)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 31, "You need magic level 31 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune, 3 air runes and 1 earth rune
    teleport_standard(player, 3222, 3218, 0, ITEM_LAW_RUNE, 1, ITEM_AIR_RUNE, 3, ITEM_EARTH_RUNE, 1)
end
