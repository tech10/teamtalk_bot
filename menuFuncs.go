package main

import (
	"strconv"
	"strings"
)

// Functions for generating menues.

func server_active_check(prompt string) *tt_server {
	if s := c.Server_active_read(); s != nil {
		return s
	}
	if prompt == "" {
		prompt = "No active server is currently selected. You must select a server for this command."
	}
	console_write(prompt)
	return server_menu(c.Servers_read())
}

func server_menu(servers []*tt_server) *tt_server {
	if len(servers) == 0 {
		console_write("No servers available for selection. Please add a server before continuing.")
		return nil
	}
	if len(servers) == 1 {
		console_write("Only 1 server is available. Automatically selecting " + servers[0].DisplayName_read())
		return servers[0]
	}
	prompt := "Please select a server.\r\n"
	menu := []string{}
	for _, server := range servers {
		menu = append(menu, server.DisplayName_read())
	}
	index, aborted := console_read_menu(prompt, menu)
	if aborted || index == -1 {
		return nil
	}
	return servers[index]
}

func server_modify_menu_prompt(server *tt_server) bool {
	if server == nil {
		return true
	}
	menu := []string{}
	funcs := []func(*tt_server) bool{}
	// 0 is all server options.
	menu = append(menu, "All server options.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_modify_prompt(server, true)
	})
	// 1 is the display name.
	menu = append(menu, "Display name.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_displayname(server, true)
	})
	// 2 is connection information.
	menu = append(menu, "Connection information.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_conn_info(server, true)
	})
	// 3 is account information.
	menu = append(menu, "Account information.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_account_info(server, true)
	})
	// 4 is the nickname.
	menu = append(menu, "Nickname.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_nickname(server, true)
	})
	// 5 is the autoconnect information.
	menu = append(menu, "Autoconnect information.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_autoconnect_info(server, true)
	})
	// 6 is the autosubscription information.
	menu = append(menu, "Automatic subscriptions information.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_autosubscriptions(server, true)
	})
	// 7 is the display information.
	menu = append(menu, "Displayed events information.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_events_info(server, true)
	})
	// 8 is the logging information.
	menu = append(menu, "Log information.")
	funcs = append(funcs, func(server *tt_server) bool {
		return c.Server_prompt_log_info(server, true)
	})
	// 8 is done.
	menu = append(menu, "Done.")
	funcs = append(funcs, func(server *tt_server) bool {
		console_write("Command complete.")
		return true
	})
	res, aborted := console_read_menu("Please select the options you wish to modify.\r\n", menu)
	if aborted || res == -1 {
		return true
	}
	return funcs[res](server)
}

func servers_modify_menu_prompt() bool {
	if len(c.Servers_read()) == 0 {
		console_write("No servers are available to modify configuration options for. Aborted.")
		return true
	}
	menu := []string{}
	funcs := []func() bool{}
	// 0 is autoconnect on start.
	menu = append(menu, "Autoconnect on program start")
	funcs = append(funcs, func() bool {
		return c.Servers_AutoConnectOnStart_prompt()
	})
	// 1 is autoconnect on connection loss.
	menu = append(menu, "Reconnect on connection loss")
	funcs = append(funcs, func() bool {
		return c.Servers_AutoConnectOnDisconnect_prompt()
	})
	// 2 is reconnect on kick.
	menu = append(menu, "Reconnect if kicked")
	funcs = append(funcs, func() bool {
		return c.Servers_AutoConnectOnKick_prompt()
	})
	// 3 is autosubscriptions.
	menu = append(menu, "Set automatic user subscriptions")
	funcs = append(funcs, func() bool {
		return c.Servers_AutoSubscriptions_prompt()
	})
	// 4 is display server events.
	menu = append(menu, "Display server events")
	funcs = append(funcs, func() bool {
		return c.Servers_DisplayEvents_prompt()
	})
	// 5 is display extended connection information.
	menu = append(menu, "Display extended connection information")
	funcs = append(funcs, func() bool {
		return c.Servers_DisplayExtendedConnInfo_prompt()
	})
	// 6 is display status updates.
	menu = append(menu, "Display status updates")
	funcs = append(funcs, func() bool {
		return c.Servers_DisplayStatusUpdates_prompt()
	})
	// 7 is display subscription updates.
	menu = append(menu, "Display subscription updates")
	funcs = append(funcs, func() bool {
		return c.Servers_DisplaySubscriptionUpdates_prompt()
	})
	// 8 is beep on critical events.
	menu = append(menu, "Beep on critical events")
	funcs = append(funcs, func() bool {
		return c.Servers_BeepOnCriticalEvents_prompt()
	})
	// 9 is log information.
	menu = append(menu, "Logging information")
	funcs = append(funcs, func() bool {
		return c.Servers_LogInfo_prompt()
	})
	// 10 is use global nickname.
	menu = append(menu, "Use global nickname")
	funcs = append(funcs, func() bool {
		return c.Servers_UseGlobalNickName_prompt()
	})
	// 11 is done.
	menu = append(menu, "Done.")
	funcs = append(funcs, func() bool {
		return false
	})

	res, aborted := console_read_menu("Please select the options you wish to modify.\r\n", menu)
	if aborted || res == -1 {
		return true
	}
	return funcs[res]()
}

func conf_modify_menu_prompt() bool {
	menu := []string{}
	funcs := []func() bool{}
	// 0 is all default configuration options.
	menu = append(menu, "All default options.")
	funcs = append(funcs, func() bool {
		return c.Defaults_prompt()
	})
	// 1 is the default autoconnect information.
	menu = append(menu, "Autoconnect information.")
	funcs = append(funcs, func() bool {
		return c.AutoConnectInfo_prompt()
	})
	// 2 is default autosubscription information.
	menu = append(menu, "Autosubscriptions.")
	funcs = append(funcs, func() bool {
		return c.AutoSubscriptions_prompt()
	})
	// 3 is default displayed events information.
	menu = append(menu, "Displayed information.")
	funcs = append(funcs, func() bool {
		return c.EventsInfo_prompt()
	})
	// 4 is logging information.
	menu = append(menu, "Log information.")
	funcs = append(funcs, func() bool {
		return c.LogInfo_prompt()
	})
	// 5 is the use defaults prompt.
	menu = append(menu, "Use defaults on server creation.")
	funcs = append(funcs, func() bool {
		return c.UseDefaults_prompt()
	})
	res, aborted := console_read_menu("Please select the options you wish to modify.\r\n", menu)
	if aborted || res == -1 {
		return true
	}
	return funcs[res]()
}

func server_menu_src(server *tt_server, src string) (*tt_user, *tt_channel, bool) {
	if server == nil {
		return nil, nil, true
	}
	var usr *tt_user
	var ch *tt_channel
	if src == "" {
		val, err := console_read_prompt("Please enter a nickname, username, user ID, channel name, channel ID, or nothing to be guided through individual prompts for user or channel selection.")
		if err != nil {
			return nil, nil, true
		}
		src = val
	}
	aborted := false
	if strings.Contains(src, "/") {
		ch, aborted = server_menu_channel(server, src)
		if aborted {
			return nil, nil, true
		}
		if ch != nil {
			return nil, ch, false
		}
		usr, aborted = server_menu_user(server, src)
		return usr, nil, aborted
	}
	usr, aborted = server_menu_user(server, src)
	if aborted {
		return nil, nil, true
	}
	if usr != nil {
		return usr, nil, false
	}
	ch, aborted = server_menu_channel(server, src)
	return nil, ch, aborted
}

func server_menu_option_user(usr *tt_user) string {
	str := ""
	if nickname := usr.NickName_read(); nickname != "" {
		str += nickname
	}
	if username := usr.UserName_read(); username != "" {
		if str != "" {
			str += " (" + username + ")"
		} else {
			str += username
		}
	}
	id_str := strconv.Itoa(usr.Uid_read())
	if str == "" {
		str += id_str
	} else {
		str += " (" + id_str + ")"
	}
	return str
}

func server_menu_select_users(server *tt_server, usrval string) ([]*tt_user, bool) {
	users := []*tt_user{}
	if server == nil {
		return users, true
	}
	if usrval == "" {
		val, err := console_read_prompt("Please enter a nickname, username, or user id. Press enter for all users.")
		if err != nil {
			return users, true
		}
		usrval = val
	}
	uid, err := strconv.Atoi(usrval)
	if err == nil && uid != 0 {
		answer, aborted := console_read_confirm("Do you wish to use " + usrval + " as a user ID?")
		if aborted {
			return users, true
		}
		if answer {
			usr := server.User_find_id(uid)
			if usr == nil {
				console_write("User " + usrval + " doesn't exist.")
				return users, false
			}
			return []*tt_user{usr}, false
		}
	}
	users = server.User_find_all(usrval)
	return users, false
}

func server_menu_user(server *tt_server, usrval string) (*tt_user, bool) {
	if server == nil {
		return nil, true
	}
	var usr *tt_user
	users, aborted := server_menu_select_users(server, usrval)
	if aborted {
		return nil, true
	}
	if len(users) == 0 {
		return nil, false
	}
	count := len(users)
	prompt := strconv.Itoa(count) + " users were found. Would you like to select one of the users"
	if count == 1 {
		usr = users[0]
		return usr, false
	}
	answer, aborted := console_read_confirm(prompt + "?\r\n")
	if aborted {
		return nil, true
	}
	if !answer {
		return nil, false
	}
	menu := []string{}
	for _, u := range users {
		menu = append(menu, server_menu_option_user(u))
	}
	res, aborted := console_read_menu("Please select a user.\r\n", menu)
	if aborted || res == -1 {
		return nil, true
	}
	usr = users[res]
	return usr, false
}

func server_menu_select_channels(server *tt_server, chval string) ([]*tt_channel, bool) {
	channels := []*tt_channel{}
	if server == nil {
		return channels, true
	}
	if chval == "" {
		val, err := console_read_prompt("Please enter a channel name or channel id. Press enter for all channels.")
		if err != nil {
			return channels, true
		}
		chval = val
	}
	cid, err := strconv.Atoi(chval)
	if err == nil && cid != 0 {
		answer, aborted := console_read_confirm("Do you wish to use " + chval + " as a channel ID?")
		if aborted {
			return channels, true
		}
		if answer {
			ch := server.Channel_find_id(cid)
			if ch == nil {
				console_write("channel " + chval + " doesn't exist.")
				return channels, false
			}
			return []*tt_channel{ch}, false
		}
	}
	channels = server.Channel_find_path(chval)
	return channels, false
}

func server_menu_channel(server *tt_server, chval string) (*tt_channel, bool) {
	if server == nil {
		return nil, true
	}
	var ch *tt_channel
	channels, aborted := server_menu_select_channels(server, chval)
	if aborted {
		return nil, true
	}
	if len(channels) == 0 {
		return nil, false
	}
	count := len(channels)
	if chval != "" && count != 1 {
		prompt := strconv.Itoa(count) + " channels were found. Would you like to select one of the channels"
		answer, aborted := console_read_confirm(prompt + "?\r\n")
		if aborted {
			return nil, true
		}
		if !answer {
			return nil, false
		}
	}
	if count == 1 {
		ch = channels[0]
		return ch, false
	}
	menu := []string{}
	for _, c := range channels {
		menu = append(menu, c.Path_read())
	}
	res, aborted := console_read_menu("Please select a channel.\r\n", menu)
	if aborted || res == -1 {
		return nil, true
	}
	ch = channels[res]
	return ch, false
}
