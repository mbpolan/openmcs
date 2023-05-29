-------------------------------------
-- Spell: teleport to Varrock
-------------------------------------
function spell_teleport_varrock(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 25, "You need magic level 25 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune, 3 air runes and 1 fire rune
    teleport_standard(player, 3213, 3424, 0, 563, 1, 556, 3, 554, 1)
end
