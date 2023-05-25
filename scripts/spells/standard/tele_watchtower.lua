-------------------------------------
-- Spell: teleport to Watchtower
-------------------------------------
function spell_teleport_watchtower(player)
    -- require 2 law runes and 2 earth runes
    ok = player:consume_runes(563, 2, 557, 2)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    player:teleport(2546, 3113, 2)
end
