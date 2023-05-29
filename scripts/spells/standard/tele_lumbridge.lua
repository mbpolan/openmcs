-------------------------------------
-- Spell: teleport to Lumbridge
-------------------------------------
function spell_teleport_lumbridge(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 31, "You need magic level 31 to cast this spell.")
    if not ok then
        return
    end

    -- require 1 law rune, 3 air runes and 1 earth rune
    teleport_standard(player, 3222, 3218, 0, 563, 1, 556, 3, 557, 1)
end
