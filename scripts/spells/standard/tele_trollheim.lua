-------------------------------------
-- Spell: teleport to Trollheim
-------------------------------------
function spell_teleport_trollheim(player)
    -- require 2 law runes and 2 fire runes
    ok = player:consume_runes(563, 2, 555, 2)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    player:teleport(2890, 3678, 0)
end
