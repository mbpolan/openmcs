--- Casts the Teleport to Ape Atoll spell
-- @param player The player casting the spell
function spell_teleport_ape_atoll(player)
    local ok = stat_level_minimum(player, SKILL_MAGIC, 64, "You need magic level 64 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes, 2 fire runes, 2 water runes and 1 banana
    teleport_standard(player, 2801, 2704, 0, 74,
            ITEM_LAW_RUNE, 2,
            ITEM_FIRE_RUNE, 2,
            ITEM_WATER_RUNE, 2,
            ITEM_BANANA, 1)
end
