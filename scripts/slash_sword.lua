function interface_2423_on_update(player, item)
    -- 2424 is the weapon model
    player:interface_model(2424, item:id(), 169)

    -- 2425 is the weapon name
    player:interface_text(2426, " " .. item:name())
end
