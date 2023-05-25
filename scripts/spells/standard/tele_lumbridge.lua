-------------------------------------
-- Spell: teleport to Lumbridge
-------------------------------------
function spell_teleport_lumbridge(player)
    -- require 1 law rune, 3 air runes and 1 earth rune
    ok = player:consume_runes(563, 1, 556, 3, 557, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    player:teleport(3222, 3218, 0)
end
