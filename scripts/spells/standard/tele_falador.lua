-------------------------------------
-- Spell: teleport to Falador
-------------------------------------
function spell_teleport_falador(player)
    -- require 1 law rune, 3 air runes and 1 water rune
    ok = player:consume_runes(563, 1, 556, 3, 555, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    player:teleport(2963, 3379, 0)
end
