-- Casts the Teleport to Varrock spell
-- @param player The player casting the spell
function spell_teleport_varrock(player)
    local ok = stat_level_minimum(player, SKILL_MAGIC, 25, "You need magic level 25 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune, 3 air runes and 1 fire rune
    teleport_standard(player, 3213, 3424, 0, 35,
            ITEM_LAW_RUNE, 1,
            ITEM_AIR_RUNE, 3,
            ITEM_FIRE_RUNE, 1)
end
