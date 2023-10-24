-- Casts the Teleport to Trollheim spell
-- @param player The player casting the spell
function spell_teleport_trollheim(player)
    local ok = stat_level_minimum(player, SKILL_MAGIC, 61, "You need magic level 61 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes and 2 fire runes
    teleport_standard(player, 2890, 3678, 0, 68,
            ITEM_LAW_RUNE, 2,
            ITEM_FIRE_RUNE, 2)
end
