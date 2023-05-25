-------------------------------------
-- Spell: teleport to Ardougne
-------------------------------------
function spell_teleport_ardougne(player)
    -- require 2 law runes and 2 water runes
    ok = player:consume_runes(563, 2, 555, 2)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    player:teleport(2662, 3306, 0)
end
