-------------------------------------
-- Interface: logout tab
-------------------------------------

--- Handles an action performed on the logout tab interface.
-- @param player The player performing the action
-- @param interface The subinterface that received the action
function interface_2449_on_action(player, interface)
    -- player clicked on the logout button
    if interface:id() == 2458 then
        player:disconnect()
    end
end
