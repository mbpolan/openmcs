-------------------------------------
-- Player equipment interactions
-------------------------------------

--- Handles a player equipping an item.
-- @param player The player performing the action
-- @param item The item being equipped
function on_equip_item(player, item)
    -- choose the appropriate interface based on the item's weapon attack style
    if item:equipment_slot() == EQUIP_SLOT_WEAPON then
        local inf_id = 0

        local style = item:weapon_style()
        if style == WEAPON_STYLE_2H_SWORD then
            inf_id = 4705
        elseif style == WEAPON_STYLE_AXE then
            inf_id = 1698
        elseif style == WEAPON_STYLE_BOW then
            inf_id = 1764
        elseif style == WEAPON_STYLE_BLUNT then
            inf_id = 425
        elseif style == WEAPON_STYLE_CLAW then
            inf_id = 7762
        elseif style == WEAPON_STYLE_CROSSBOW then
            inf_id = 1749
        elseif style == WEAPON_STYLE_GUN then
            inf_id = 13975
        elseif style == WEAPON_STYLE_PICKAXE then
            inf_id = 5570
        elseif style == WEAPON_STYLE_POLEARM then
            inf_id = 8460
        elseif style == WEAPON_STYLE_POLESTAFF then
            inf_id = 6103
        elseif style == WEAPON_STYLE_SCYTHE then
            inf_id = 776
        elseif style == WEAPON_STYLE_SLASH_SWORD then
            inf_id = 2423
            inf_func = interface_2423_on_update
        elseif style == WEAPON_STYLE_SPEAR then
            inf_id = 4679
        elseif style == WEAPON_STYLE_SPIKED then
            inf_id = 3796
        elseif style == WEAPON_STYLE_STAB_SWORD then
            inf_id = 2276
        elseif style == WEAPON_STYLE_STAFF then
            inf_id = 328
        elseif style == WEAPON_STYLE_THROWN then
            inf_id = 4446
        elseif style == WEAPON_STYLE_WHIP then
            inf_id = 12290
        end

        if inf_id > 0 then
            player:sidebar_interface(CLIENT_TAB_EQUIPPED_ITEM, inf_id)
        end

        if inf_func ~= nil then
            inf_func(player, item)
        end
    end

    set_equip_stats(player)
end

--- Updates the text for each combat statistic.
-- @param player The player to format statistics for
function set_equip_stats(player)
    local stats = player:combat_stats()

    -- update combat stats text interfaces
    player:interface_text(1675, format_stat("Stab", stats[STAT_ATTACK_STAB]))
    player:interface_text(1676, format_stat("Slash", stats[STAT_ATTACK_SLASH]))
    player:interface_text(1677, format_stat("Crush", stats[STAT_ATTACK_CRUSH]))
    player:interface_text(1678, format_stat("Magic", stats[STAT_ATTACK_MAGIC]))
    player:interface_text(1679, format_stat("Range", stats[STAT_ATTACK_RANGE]))

    player:interface_text(1680, format_stat("Stab", stats[STAT_DEFENSE_STAB]))
    player:interface_text(1681, format_stat("Slash", stats[STAT_DEFENSE_SLASH]))
    player:interface_text(1682, format_stat("Crush", stats[STAT_DEFENSE_CRUSH]))
    player:interface_text(1683, format_stat("Magic", stats[STAT_DEFENSE_MAGIC]))
    player:interface_text(1684, format_stat("Range", stats[STAT_DEFENSE_RANGE]))

    player:interface_text(1686, format_stat("Strength", stats[STAT_STRENGTH]))
    player:interface_text(1687, format_stat("Prayer", stats[STAT_PRAYER]))
end

--- Formats a combat statistic for displaying on the equipment interface.
-- @param name The name of the statistic
-- @param value The statistic value
-- @return a formatted string
function format_stat(name, value)
    if value > 0 then
        return name .. ": +" .. value
    end

    return name .. ": " .. value
end

--- Handles a player unequipping an item.
-- @param player The player performing the action
-- @param item The item being equipped
function on_unequip_item(player, item)
    if item:equipment_slot() == EQUIP_SLOT_WEAPON then
        set_unarmed(player)
    end

    set_equip_stats(player)
end

--- Sets the equipped item interface to unarmed.
-- @param player The player
function set_unarmed(player)
    player:sidebar_interface(CLIENT_TAB_EQUIPPED_ITEM, 5855)
    interface_5855_on_update(player)
end

-------------------------------------
-- Interface: unarmed
-------------------------------------

--- Handles an action performed on the unarmed weapon interface.
-- @param player The player performing the action
-- @param interface The subinterface that received the action
-- @param op_code The op code from the interaction
function interface_5855_on_action(player, interface, op_code)
    local style = interface:id()

    -- change the player's current weapon attack style
    if style == 5860 then
        player:attack_style(ATTACK_STYLE_PUNCH)
    elseif style == 5862 then
        player:attack_style(ATTACK_STYLE_KICK)
    elseif style == 5861 then
        player:attack_style(ATTACK_STYLE_BLOCK)
    end
end

--- Handles updating the unarmed weapon interface.
-- @param player The player
function interface_5855_on_update(player)
    -- 2425 is the weapon name
    player:interface_text(5857, "none")

    local style = player:attack_style()

    local style_value = -1
    if style == ATTACK_STYLE_PUNCH then
        style_value = 0
    elseif style == ATTACK_STYLE_KICK then
        style_value = 1
    elseif style == ATTACK_STYLE_BLOCK then
        style_value = 2
    end

    if style_value > -1 then
        player:interface_setting(43, style_value)
    end
end

-------------------------------------
-- Interface: slash/sword weapon
-------------------------------------

--- Handles an action performed on the slash/sword weapon interface.
-- @param player The player performing the action
-- @param interface The subinterface that received the action
-- @param op_code The op code from the interaction
function interface_2423_on_action(player, interface, op_code)
    local style = interface:id()

    -- change the player's current weapon attack style
    if style == 2429 then
        player:attack_style(ATTACK_STYLE_CHOP)
    elseif style == 2432 then
        player:attack_style(ATTACK_STYLE_SLASH)
    elseif style == 2431 then
        player:attack_style(ATTACK_STYLE_LUNGE)
    elseif style == 2430 then
        player:attack_style(ATTACK_STYLE_BLOCK)
    end
end

--- Handles updating the slash/sword weapon interface.
-- @param player The player
function interface_2423_on_update(player, item)
    -- 2424 is the weapon model
    player:interface_model(2424, item:id(), 169)

    -- 2425 is the weapon name
    player:interface_text(2426, " " .. item:name())

    local style = player:attack_style()

    -- setting id 43 toggles the appropriate attack style button
    local style_value = -1
    if style == ATTACK_STYLE_CHOP then
        style_value = 0
    elseif style == ATTACK_STYLE_SLASH then
        style_value = 1
    elseif style == ATTACK_STYLE_LUNGE then
        style_value = 2
    elseif style == ATTACK_STYLE_BLOCK then
        style_value = 3
    end

    if style_value > -1 then
        player:interface_setting(43, style_value)
    end
end
