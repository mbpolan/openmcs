--- Casts the Teleport to Falador spell
-- @param player The player casting the spell
function spell_teleport_falador(player)
    local ok = stat_level_minimum(player, SKILL_MAGIC, 37, "You need magic level 37 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune, 3 air runes and 1 water rune
    teleport_standard(player, 2963, 3379, 0, 48,
            ITEM_LAW_RUNE, 1,
            ITEM_AIR_RUNE, 3,
            ITEM_WATER_RUNE, 1)
end
