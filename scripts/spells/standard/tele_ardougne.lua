--- Casts the Teleport to Ardougne spell
-- @param player The player casting the spell
function spell_teleport_ardougne(player)
    local ok = skill_level_minimum(player, SKILL_MAGIC, 51, "You need magic level 51 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes and 2 water runes
    teleport_standard(player, 2662, 3306, 0, ITEM_LAW_RUNE, 2, ITEM_WATER_RUNE, 2)
end
