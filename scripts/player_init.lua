-------------------------------------
-- Initialize a player on log in
-------------------------------------
function init_player_tabs(player)
    -- set initial sidebar interfaces
    player:sidebar_interface(CLIENT_TAB_EQUIPMENT, 1644)
    player:sidebar_interface(CLIENT_TAB_INVENTORY, 3213)
    player:sidebar_interface(CLIENT_TAB_SKILLS, 3917)
    player:sidebar_interface(CLIENT_TAB_LOGOUT, 2449)
    player:sidebar_interface(CLIENT_TAB_FRIENDS_LIST, 5065)
    player:sidebar_interface(CLIENT_TAB_IGNORE_LIST, 5715)

    -- set the equipped item interface based on the currently equipped weapon
    item = player:equipped_item(EQUIP_SLOT_WEAPON)
    if item == nil then
        on_unequip_item(player)
    else
        on_equip_item(player, item)
    end

    -- TODO: not yet supported by game engine
    player:sidebar_clear(CLIENT_TAB_QUESTS)
    player:sidebar_clear(CLIENT_TAB_PRAYERS)
    player:sidebar_clear(CLIENT_TAB_SPELLS)
    player:sidebar_clear(CLIENT_TAB_SETTINGS)
    player:sidebar_clear(CLIENT_TAB_CONTROLS)
    player:sidebar_clear(CLIENT_TAB_MUSIC)
end