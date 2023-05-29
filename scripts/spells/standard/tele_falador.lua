-------------------------------------
-- Spell: teleport to Falador
-------------------------------------
function spell_teleport_falador(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 37, "You need magic level 37 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune, 3 air runes and 1 water rune
    teleport_standard(player, 2963, 3379, 0, 563, 1, 556, 3, 555, 1)
end
