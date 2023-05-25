-------------------------------------
-- Spell: teleport to Varrock
-------------------------------------
function spell_teleport_varrock(player)
    -- require 1 law rune, 3 air runes and 1 fire rune
    ok = player:consume_runes(563, 1, 556, 3, 554, 1)
    if not ok then
        player:server_message("You do not have enough runes to cast this spell.")
        return
    end

    player:teleport(3213, 3424, 0)
end
