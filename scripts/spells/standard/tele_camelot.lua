-------------------------------------
-- Spell: teleport to Camelot
-------------------------------------
function spell_teleport_camelot(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 45, "You need magic level 45 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune and 10 air runes
    teleport_standard(player, 2757, 3478, 0, 563, 1, 556, 5)
end
