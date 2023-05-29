-------------------------------------
-- Spell: teleport to Watchtower
-------------------------------------
function spell_teleport_watchtower(player)
    ok = skill_level_minimum(player, SKILL_MAGIC, 58, "You need magic level 58 to cast this spell.")
    if not ok then
        return
    end

    -- require 2 law runes and 2 earth runes
    teleport_standard(player, 2546, 3113, 2, 563, 2, 557, 2)
end
