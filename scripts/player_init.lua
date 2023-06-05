-------------------------------------
-- Player initialization
-------------------------------------

--- Handles initializing a player's dynamic parameters when they log into the game.
-- @param player The player to initialize
function init_player_tabs(player)
    -- initialize game options
    init_player_game_options(player)

    local low_memory = player:low_memory()

    -- set initial sidebar interfaces
    player:sidebar_interface(CLIENT_TAB_EQUIPMENT, 1644)
    player:sidebar_interface(CLIENT_TAB_INVENTORY, 3213)
    player:sidebar_interface(CLIENT_TAB_SKILLS, 3917)
    player:sidebar_interface(CLIENT_TAB_LOGOUT, 2449)
    player:sidebar_interface(CLIENT_TAB_FRIENDS_LIST, 5065)
    player:sidebar_interface(CLIENT_TAB_IGNORE_LIST, 5715)
    player:sidebar_interface(CLIENT_TAB_SPELLS, 1151)
    player:sidebar_interface(CLIENT_TAB_CONTROLS, 147)
    player:sidebar_interface(CLIENT_TAB_QUESTS, 638)

    -- set conditional sidebar interfaces
    if low_memory then
        player:sidebar_interface(CLIENT_TAB_SETTINGS, 4445)
        player:sidebar_interface(CLIENT_TAB_MUSIC, 6299)
        interface_4445_on_update(player)
    else
        player:sidebar_interface(CLIENT_TAB_SETTINGS, 904)
        player:sidebar_interface(CLIENT_TAB_MUSIC, 962)
        interface_904_on_update(player)
        interface_962_on_update(player)
    end

    -- set the equipped item interface based on the currently equipped weapon
    local item = player:equipped_item(EQUIP_SLOT_WEAPON)
    if item == nil then
        set_unarmed(player)
        set_equip_stats(player)
    else
        on_equip_item(player, item)
    end

    -- update other interfaces
    interface_638_on_update(player)

    -- TODO: not yet supported by game engine
    player:sidebar_clear(CLIENT_TAB_PRAYERS)
end

--- Initializes a player's game option preferences.
-- @param player The player
function init_player_game_options(player)
    local brightness = player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS)
    if brightness == "" then
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_NORMAL)
    end

    local chat_effects = player:game_option(GAME_OPTION_CHAT_EFFECTS)
    if chat_effects == "" then
        player:game_option(GAME_OPTION_CHAT_EFFECTS, CHAT_EFFECTS_ON)
    end

    local split_private = player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT)
    if split_private == "" then
        player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT, SPLIT_PRIVATE_CHAT_ON)
    end

    local mouse_buttons = player:game_option(GAME_OPTION_MOUSE_BUTTONS)
    if mouse_buttons == "" then
        player:game_option(GAME_OPTION_MOUSE_BUTTONS, MOUSE_BUTTONS_TWO)
    end

    local accept_aid = player:game_option(GAME_OPTION_ACCEPT_AID)
    if accept_aid == "" then
        player:game_option(GAME_OPTION_ACCEPT_AID, ACCEPT_AID_YES)
    end

    local music_volume = player:game_option(GAME_OPTION_MUSIC_VOLUME)
    if music_volume == "" then
        player:game_option(GAME_OPTION_MUSIC_VOLUME, "4")
    end

    local effects_volume = player:game_option(GAME_OPTION_EFFECTS_VOLUME)
    if effects_volume == "" then
        player:game_option(GAME_OPTION_EFFECTS_VOLUME, "4")
    end
end