-------------------------------------
-- Spell: teleport to Trollheim
-------------------------------------
function spell_teleport_trollheim(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 61, "You need magic level 61 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes and 2 fire runes
    teleport_standard(player, 2890, 3678, 0, 563, 2, 554, 2)
end
