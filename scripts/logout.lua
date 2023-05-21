-------------------------------------
-- Interface: logout tab
-------------------------------------
function interface_2449_on_action(player, interface)
    -- player clicked on the logout button
    if interface:id() == 2458 then
        player:disconnect()
    end
end
