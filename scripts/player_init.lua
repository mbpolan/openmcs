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

    -- TODO: not yet supported by game engine
    player:sidebar_clear(CLIENT_TAB_EQUIPPED_ITEM)
    player:sidebar_clear(CLIENT_TAB_QUESTS)
    player:sidebar_clear(CLIENT_TAB_PRAYERS)
    player:sidebar_clear(CLIENT_TAB_SPELLS)
    player:sidebar_clear(CLIENT_TAB_SETTINGS)
    player:sidebar_clear(CLIENT_TAB_CONTROLS)
    player:sidebar_clear(CLIENT_TAB_MUSIC)
end
