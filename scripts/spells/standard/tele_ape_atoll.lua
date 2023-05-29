-------------------------------------
-- Spell: teleport to Ape Atoll
-------------------------------------
function spell_teleport_ape_atoll(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 64, "You need magic level 64 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes, 2 fire runes, 2 water runes and 1 banana
    teleport_standard(player, 2801, 2704, 0, 563, 2, 554, 2, 555, 2, 1963, 1)
end
