-------------------------------------
-- Handles a player equipping items
-------------------------------------
function on_equip_item(player, item)
    -- choose the appropriate interface based on the item's weapon attack style
    inf_id = 0
    style = item:weapon_style()
    if style == WEAPON_STYLE_2H_SWORD then
        inf_id = 5855
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
end

-------------------------------------
-- Handles a player unequipping items
-------------------------------------
function on_unequip_item(player)
    player:sidebar_interface(CLIENT_TAB_EQUIPPED_ITEM, 5855)
end
