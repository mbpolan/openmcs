--- Handles an action performed on the game options interface (low-memory mode).
-- @param player The player performing the action
-- @param interface The interface that received the action
function interface_4445_on_action(player, interface)
    local id = interface:id()

    -- screen brightness
    if id == 5452 then
        -- dark
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_DARK)
    elseif id == 6273 then
        -- normal
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_NORMAL)
    elseif id == 6275 then
        -- bright
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_BRIGHT)
    elseif id == 6276 then
        -- very bright
        player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS, SCREEN_BRIGHTNESS_VERY_BRIGHT)
    end

    -- mouse buttons
    if id == 6278 then
        -- two
        player:game_option(GAME_OPTION_MOUSE_BUTTONS, MOUSE_BUTTONS_TWO)
    elseif id == 6279 then
        -- one
        player:game_option(GAME_OPTION_MOUSE_BUTTONS, MOUSE_BUTTONS_ONE)
    end

    -- chat effects
    if id == 6280 then
        -- on
        player:game_option(GAME_OPTION_CHAT_EFFECTS, CHAT_EFFECTS_ON)
    elseif id == 6281 then
        -- off
        player:game_option(GAME_OPTION_CHAT_EFFECTS, CHAT_EFFECTS_OFF)
    end

    -- split private chat
    if id == 952 then
        -- on
        player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT, SPLIT_PRIVATE_CHAT_ON)
    elseif id == 953 then
        -- off
        player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT, SPLIT_PRIVATE_CHAT_OFF)
    end

    -- accept aid
    if id == 12591 then
        -- yes
        player:game_option(GAME_OPTION_ACCEPT_AID, ACCEPT_AID_YES)
    elseif id == 12590 then
        -- no
        player:game_option(GAME_OPTION_ACCEPT_AID, ACCEPT_AID_NO)
    end
end

--- Handles updating the game options interface (low-memory mode).
-- @param player The player
function interface_4445_on_update(player)
    -- screen brightness: op code 166
    local brightness = player:game_option(GAME_OPTION_SCREEN_BRIGHTNESS)
    if brightness == SCREEN_BRIGHTNESS_DARK then
        player:interface_setting(166, 1)
    elseif brightness == SCREEN_BRIGHTNESS_NORMAL then
        player:interface_setting(166, 2)
    elseif brightness == SCREEN_BRIGHTNESS_BRIGHT then
        player:interface_setting(166, 3)
    elseif brightness == SCREEN_BRIGHTNESS_VERY_BRIGHT then
        player:interface_setting(166, 4)
    end

    -- chat effects: op code 171
    local chat_effects = player:game_option(GAME_OPTION_CHAT_EFFECTS)
    if chat_effects == CHAT_EFFECTS_ON then
        player:interface_setting(171, 0)
    elseif chat_effects == CHAT_EFFECTS_OFF then
        player:interface_setting(171, 1)
    end

    -- split private chat: op code 287
    local split_private = player:game_option(GAME_OPTION_SPLIT_PRIVATE_CHAT)
    if split_private == SPLIT_PRIVATE_CHAT_ON then
        player:interface_setting(287, 1)
    elseif split_private == SPLIT_PRIVATE_CHAT_OFF then
        player:interface_setting(287, 0)
    end

    -- mouse buttons: op code 170
    local mouse_buttons = player:game_option(GAME_OPTION_MOUSE_BUTTONS)
    if mouse_buttons == MOUSE_BUTTONS_ONE then
        player:interface_setting(170, 1)
    elseif mouse_buttons == MOUSE_BUTTONS_TWO then
        player:interface_setting(170, 0)
    end

    -- accept aid: op code 427
    local accept_aid = player:game_option(GAME_OPTION_ACCEPT_AID)
    if accept_aid == ACCEPT_AID_YES then
        player:interface_setting(427, 1)
    elseif accept_aid == ACCEPT_AID_NO then
        player:interface_setting(427, 0)
    end
end