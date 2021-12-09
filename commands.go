package main

import (
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var commands Commands

func init() {
	commands = NewCommands()

	commands.AddHelp("help",
		"Provides a list of available commands, or help with a specific command or list of commands.",
		"help\nProvides a list of commands.",
		"help quit\nProvides help for the quit command.",
		"help quit join server\nProvides help for the quit, join, and server commands.")
	commands.Add("help",
		func(cmd string) {
			if cmd != "" {
				cmdlist := strings.Split(cmd, " ")
				helpmsg := []string{}
				for _, cmd := range cmdlist {
					noexist := "The command " + cmd + " doesn't exist."
					if cmd == "" {
						continue
					}
					if !commands.Exists(cmd) {
						helpmsg = append(helpmsg, noexist, "")
						continue
					}
					msg := commands.HelpText(cmd)
					if msg == "" {
						helpmsg = append(helpmsg, "No help is available for the command "+cmd+".", "")
						continue
					}
					helpmsg = append(helpmsg, msg, "")
				}
				console_write(strings.TrimSuffix(strings.Join(helpmsg, "\r\n"), "\r\n"))
				return
			}
			helpmsg := []string{"For more information on an available command, type \"help <command>\".\r\nFor example, to obtain information on the command quit, type \"help quit\".\r\n", "Available commands:"}
			helpmsg = append(helpmsg, commands.cmdorder...)
			console_write(strings.TrimSuffix(strings.Join(helpmsg, "\r\n"), "\r\n"))
		})

	commands.AddHelp("quit",
		"Disconnects from all servers and shuts down the program.\r\nYou can also quit the program with CTRL+C, or through any means your operating system has to terminate programs.\r\nCTRL+D works as well.")
	commands.Add("quit",
		func(param string) {
			console_close()
			quit <- true
			if !console_use_readline() {
				c.wg.Done()
				console_close()
				runtime.Goexit()
			}
		})

	commands.AddHelp("exit",
		"Same as the quit command.")
	commands.Add("exit",
		func(param string) {
			commands.Exec("quit", "")
		})

	commands.AddHelp("config",
		"Configures the server's various settings.",
		"config\r\nProvides prompts to modify the main configuration.",
		"config nick\r\nWill modify the global nickname.",
		"config nick myname\r\nWill set the global nickname to myname.",
		"config defaults\r\nWill configure defaults for servers created using values entered into the server command, and additionally, will add servers with no values entered using those defaults.",
		"config timestamp\r\nWill configure whether or not to display timestamps on the console.",
		"config timestamp y\r\nWill automatically set this value to yes.")
	commands.Add("config",
		func(param string) {
			opt := ""
			value := ""
			if param == "" {
				if c.NickName_prompt() {
					return
				}
				if c.DisplayTimestamp_prompt() {
					return
				}
				if c.Defaults_prompt() {
					return
				}
				console_write("Command complete.")
				return
			}
			params := strings.Split(param, " ")
			opt = params[0]
			if len(params) > 1 {
				value = strings.Join(params[1:], " ")
			}
			if opt == "" {
				console_write("Error: empty option value unsupported.")
				console_write(commands.HelpText("config"))
				return
			}
			switch strings.ToLower(opt) {
			case "nick", "nickname":
				nickname := c.NickName_read()
				msg := ""
				if value == "" {
					if c.NickName_prompt() {
						return
					}
				} else {
					if value == nickname {
						console_write("Global nickname already set to " + value + ". Unchanged.")
						return
					} else {
						c.NickName_set(value)
						if c.UseGlobalNickName_prompt() {
							return
						}
					}
				}
				newnick := c.NickName_read()
				if nickname == "" {
					if newnick != "" {
						msg = "The global nickname is now " + newnick + "."
					}
				} else {
					if newnick != "" {
						msg = "The global nickname has been changed from " + nickname + " to " + newnick + "."
					} else {
						msg = "The global nickname has been changed to an empty value from " + nickname + "."
					}
				}
				console_write(msg)
				if newnick == nickname {
					return
				}
				servers := []*tt_server{}
				for _, server := range c.Servers_read() {
					if !server.connected() {
						continue
					}
					if server.NickName_read() == "" && server.UseGlobalNickName_read() && c.NickName_read() != "" {
						servers = append(servers, server)
					}
				}
				if len(servers) == 0 {
					return
				}
				count := len(servers)
				msg = strconv.Itoa(count)
				if count != 1 {
					msg += " servers are"
				} else {
					msg += " server is"
				}
				msg += " connected and "
				if count == 1 {
					msg += "is"
				} else {
					msg += "are"
				}
				msg += " configured to use the global nickname. Would you like to update "
				if count == 1 {
					msg += "it"
				} else {
					msg += "them"
				}
				answer, aborted := console_read_confirm(msg + " now?\r\n")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
				}
				success_count := 0
				for _, server := range servers {
					if server.cmd_changenick(newnick) {
						success_count++
					}
				}
				if success_count == 0 {
					console_write("The nickname wasn't changed on any of the servers.")
				}
				msg = "The nickname was successfully changed on " + strconv.Itoa(success_count) + " server"
				if success_count != 1 {
					msg += "s"
				}
				console_write(msg + ".")
				return
			case "timestamp", "timestamps":
				oldtimestamp := c.DisplayTimestamp_read()
				if value == "" {
					if c.DisplayTimestamp_prompt() {
						return
					}
				} else {
					switch strings.ToLower(value) {
					case "y", "yes":
						if c.DisplayTimestamp_read() {
							console_write("Timestamps being displayed already.")
							return
						}
						c.DisplayTimestamp_set(true)
					case "n", "no":
						if !c.DisplayTimestamp_read() {
							console_write("Timestamps aren't being displayed already.")
							return
						}
						c.DisplayTimestamp_set(false)
					default:
						console_write("Unrecognized value: " + value)
						console_write(commands.HelpText("config"))
						return
					}
				}
				if c.DisplayTimestamp_read() {
					if oldtimestamp {
						console_write("Timestamps being displayed already.")
					} else {
						console_write("Displaying timestamps on server events.")
					}
				} else {
					if !oldtimestamp {
						console_write("Timestamps aren't being displayed already.")
					} else {
						console_write("No longer displaying timestamps on server events.")
					}
				}
				return
			case "default", "defaults":
				if conf_modify_menu_prompt() {
					return
				}
				console_write("Command complete.")
				return
			default:
				console_write("Unrecognized option: " + opt)
				console_write(commands.HelpText("config"))
				return
			}
		})

	commands.AddHelp("conf",
		"Same as config.")
	commands.Add("conf",
		func(param string) {
			commands.Exec("config", param)
		})

	commands.AddHelp("configure",
		"Same as config.")
	commands.Add("configure",
		func(param string) {
			commands.Exec("config", param)
		})

	commands.AddHelp("active",
		"This command will set and clear the active server.",
		"active test\r\nWill set your active server to test, if available.",
		"active clear\r\nWill clear the active server.",
		"active\r\nWill give you a menu of available servers to make the active server, or if only one server is available and active, will clear that server from being the active server.")
	commands.Add("active",
		func(param string) {
			reset := false
			server := c.Server_active_read()
			if param == "" {
				if server != nil {
					if len(c.Servers_read()) == 1 {
						c.Server_active_clear()
						console_write("Only one server available. Active server was " + server.DisplayName_read() + ". Active server cleared.")
						return
					}
					answer, aborted := console_read_confirm("The currently active server is " + server.DisplayName_read() + ". Do you wish to clear it?\r\n")
					if aborted {
						return
					}
					if answer {
						c.Server_active_clear()
						console_write("Active server cleared.")
						return
					} else {
						reset, aborted = console_read_confirm("Do you wish to set it to another server?\r\n")
						if aborted {
							return
						}
						if !reset {
							console_write("Canceled.")
							return
						}
					}
					servers := []*tt_server{}
					for _, s := range c.Servers_read() {
						if s == server {
							continue
						}
						servers = append(servers, s)
					}
					newserver := server_menu(servers)
					if server == nil {
						return
					}
					c.Server_active_set(newserver)
					console_write("Active server changed from " + server.DisplayName_read() + " to " + c.Server_active_read_name())
					return
				}
				server = server_menu(c.Servers_read())
				if server != nil {
					c.Server_active_set(server)
					console_write("Active server set to " + c.Server_active_read_name())
				}
				return
			}
			if param == "clear" {
				if server != nil {
					c.Server_active_clear()
					console_write("Active server cleared.")
				} else {
					console_write("Active server already cleared.")
				}
				return
			}
			servers := c.Server_find_name(param)
			if len(servers) == 0 {
				console_write("Unable to set the active server to " + param + ". Server doesn't exist.")
				return
			}
			msg := ""
			if server != nil {
				if server != servers[0] {
					msg = "Active server switched from " + server.DisplayName_read() + " to "
				} else {
					console_write("Active server unchanged.")
					return
				}
			} else {
				msg = "Active server set to "
			}
			msg += servers[0].DisplayName_read()
			c.Server_active_set(servers[0])
			console_write(msg)
		})

	commands.AddHelp("debug",
		"Enable debugging for a server.\r\nWhen this is enabled, you will receive the raw data from each sending and receiving of commands as critical log messages. They will be logged along with the standard data received.",
		"debug\r\nEnables debugging for the active server, or a selected server.",
		"debug off\r\nDisables debugging for the active server or a selected one.",
		"debug\r\nWill toggle the debug state for the active server or a selected one.")
	commands.Add("debug",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if param == "" {
				switch server.Debug_read() {
				case false:
					server.Debug_set(true)
					console_write("Debugging enabled.")
				case true:
					server.Debug_set(false)
					console_write("Debugging disabled.")
				}
				c.Write()
				return
			}
			debug := server.Debug_read()
			switch strings.ToLower(param) {
			case "on", "enable":
				if debug {
					console_write("Debugging already enabled.")
					return
				}
				server.Debug_set(true)
				console_write("Debugging enabled.")
			case "off", "disable":
				if !debug {
					console_write("Debugging already disabled.")
					return
				}
				server.Debug_set(false)
				console_write("Debugging disabled.")
			default:
				console_write("Unrecognized parameter: " + param)
				console_write(commands.HelpText("debug"))
				return
			}
			c.Write()
		})

	commands.AddHelp("raw",
		"Send a raw command to the active or a selected server.\r\nThis command is intended to be used when debugging mode is enabled for the server the command is being sent to.",
		"raw logout\r\nWill log out the client.",
		"raw ping id=3\r\nWill send the ping command to the server, with the command id of 3.")
	commands.Add("raw",
		func(param string) {
			if param == "" {
				console_write("This command requires data.")
				console_write(commands.HelpText("raw"))
				return
			}
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to send raw command. Not connected.")
				return
			}
			if !server.Debug_read() {
				answer, aborted := console_read_confirm("This command is intended to be used with debugging enabled, and debugging is currently disabled. You will not receive any error relating to the command once sent. Do you wish to continue?")
				if aborted {
					return
				}
				if !answer {
					return
				}
			}
			err := server.Write(param + "\r\n")
			if err != nil {
				console_write("Command failed to send.\r\n" + err.Error())
				return
			}
			console_write("Command successfully sent.")
		})

	commands.AddHelp("connect",
		"Will connect you to the active or selected server, or a server of your choice.",
		"connect\r\nConnects you to the active server, or one you select.",
		"connect test\r\nWill connect you to the server named test.",
		"connect test check\r\nWill connect you to the servers named test and check.")
	commands.Add("connect",
		func(param string) {
			params := stringSeperateParam(param, " ", "\"")
			if len(params) > 0 {
				for _, param := range params {
					servers := c.Server_find_name(param)
					if len(servers) == 0 {
						console_write("Unable to connect to " + param + ". The server doesn't exist.")
						continue
					}
					server := servers[0]
					if server.connected() {
						console_write("Already connected to " + param + ".")
						continue
					}
					server.Startup(true)
				}
				return
			}
			server := server_active_check("")
			if server == nil {
				return
			}
			if server.connected() {
				console_write("Already connected.")
				return
			}
			server.Startup(true)
		})

	commands.AddHelp("disconnect",
		"Will disconnect you from the active or selected server, or a server or multiple servers of your choice.",
		"disconnect\r\nWill disconnect you from the active or selected server.",
		"disconnect test\r\nWill disconnect you from the server test.",
		"disconnect \"public 1\" public2\r\nWill disconnect you from the servers named public 1 and public2.")
	commands.Add("disconnect",
		func(param string) {
			params := stringSeperateParam(param, " ", "\"")
			if len(params) > 0 {
				for _, param := range params {
					servers := c.Server_find_name(param)
					if len(servers) == 0 {
						console_write("Unable to disconnect from " + param + ". The server doesn't exist.")
						continue
					}
					server := servers[0]
					if !server.connected() {
						console_write("Already disconnected from " + param + ".")
						continue
					}
					server.Shutdown()
				}
				return
			}
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Already disconnected.")
				return
			}
			server.Shutdown()
		})

	commands.AddHelp("login",
		"Will log you in to the active or selected server, or any servers of your choice.",
		"login\r\nWill log you in to the active or selected server.",
		"login test\r\nWill log you in to the server named test.",
		"login test \"test 2\"\r\nWill log you in to servers test and test 2.")
	commands.Add("login",
		func(param string) {
			params := stringSeperateParam(param, " ", "\"")
			if len(params) > 0 {
				for _, param := range params {
					servers := c.Server_find_name(param)
					if len(servers) == 0 {
						console_write("Unable to log in to " + param + ". The server doesn't exist.")
						continue
					}
					server := servers[0]
					if !server.connected() {
						server.Startup(true)
					} else {
						console_write("Error, already connected to " + server.DisplayName_read())
					}
					return
				}
			}
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				server.Startup(true)
			} else {
				console_write("Already connected.")
			}
		})

	commands.AddHelp("logout",
		"Will log you out of the active or selected server.",
		"logout\r\nWill log you out of the active or selected server.",
		"logout test1 \"test 2\"\r\nWill log you out of servers test1 and test 2.")
	commands.Add("logout",
		func(param string) {
			params := stringSeperateParam(param, " ", "\"")
			if len(params) > 0 {
				for _, param := range params {
					servers := c.Server_find_name(param)
					if len(servers) == 0 {
						console_write("Unable to log out of " + param + ". The server doesn't exist.")
						continue
					}
					server := servers[0]
					_, err := server.Logout()
					if err != nil {
						console_write("Failed to log out of " + param + ".\r\nError: " + err.Error())
					}
				}
				return
			}
			server := server_active_check("")
			if server == nil {
				return
			}
			server.Logout()
		})

	commands.AddHelp("server",
		"Commands to add, change, or remove data for a server.",
		"server\r\nWill give you a list for the available commands.",
		"server  test\r\nWill give you the option to either modify or remove the server named test.",
		"server add\r\nWill guide you through the prompts to add a server.",
		"server add <name> <hostname> <port> <username> <password> <nickname>\r\nAdds a server with any of the above parameters, which can be surrounded by quotes (\") to include spaces within a parameter.",
		"server create\r\nSame as server add.",
		"server change\r\nWill modify data for a server you select.",
		"server change test\r\nWill modify the server named test.",
		"server change all\r\nWill change a setting for all servers.",
		"server modify\r\nSame as the change parameter.",
		"server mod\r\nSame as the change parameter.",
		"server del\r\nWill remove a server you select.",
		"server del test\r\nWill remove the server named test.",
		"server delete\r\nserver remove\r\nSame as server del.")
	commands.Add("server",
		func(param string) {
			cmd := ""
			sname := ""
			sparams := []string{}
			params := strings.Split(param, " ")
			var server *tt_server
			if len(params) >= 1 {
				cmd = strings.ToLower(params[0])
			}
			if len(params) > 1 {
				sname = strings.Join(params[1:], " ")
			}
			if cmd == "add" || cmd == "create" {
				if len(params) >= 2 {
					sname = ""
					sparams = stringSeperateParam(strings.Join(params[1:], " "), " ", "\"")
				}
			}
			if cmd == "" {
				menu := []string{}
				if sname == "" {
					menu = append(menu, "Add a server")
				}
				menu = append(menu, "Modify one server or configuration options for all servers", "Remove server")
				res, aborted := console_read_menu("Please select from the available options.\r\n", menu)
				if aborted {
					return
				}
				cmd = strings.ToLower(strings.Split(menu[res], " ")[0])
			}
			if sname != "" && sname != "all" {
				servers := c.Server_find_name(sname)
				if len(servers) == 0 {
					console_write("The server " + sname + " doesn't exist.")
					return
				}
				server = servers[0]
			}
			switch cmd {
			case "add", "create":
				if len(sparams) == 0 {
					if c.Server_add_prompt() {
						return
					}
					console_write("Server added.")
					return
				}
				new_server := NewServer(c)
				if len(sparams) >= 1 {
					new_server.DisplayName_set(sparams[0])
				}
				if len(sparams) >= 2 {
					new_server.Host_set(sparams[1])
				}
				if len(sparams) >= 3 {
					new_server.Tcpport_set(sparams[2])
				}
				if len(sparams) >= 4 {
					new_server.AccountName_set(sparams[3])
				}
				if len(sparams) >= 5 {
					new_server.AccountPassword_set(sparams[4])
				}
				if len(sparams) >= 6 {
					new_server.NickName_set(sparams[5])
				}
				if c.Server_modify_prompt(new_server, false) {
					return
				}
				if c.Server_add(new_server) {
					console_write("Server added.")
					new_server.Startup(new_server.AutoConnectOnStart_read())
				} else {
					console_write("Failed to add server.")
				}
				return
			case "change", "modify", "mod":
				modify_all := false
				if sname == "all" {
					modify_all = true
				}
				if server == nil {
					if !modify_all {
						answer, aborted := console_read_confirm("Do you wish to modify configuration values for all servers?\r\n")
						if aborted {
							return
						}
						if answer {
							modify_all = true
						}
					}
					if !modify_all {
						server = server_menu(c.Servers_read())
					}
				}
				if modify_all {
					answer, aborted := console_read_confirm("WARNING: modifying all servers configuration values will change the selected value to the setting you specify for all available servers. Are you sure this is what you want to do?\r\n")
					if aborted {
						return
					}
					if !answer {
						console_write("Aborted.")
						return
					}
					if servers_modify_menu_prompt() {
						return
					}
					c.Write()
					console_write("Command complete.")
					return
				}
				if server == nil {
					return
				}
				console_write("This server has the following information.\r\n" + server.Info_str())
				answer, aborted := console_read_confirm("Is this the server you wish to modify?\r\n")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
					return
				}
				if server_modify_menu_prompt(server) {
					return
				}
				c.Write()
				console_write("Command complete.")
				return
			case "del", "delete", "remove":
				if server == nil {
					server = server_menu(c.Servers_read())
				}
				if server == nil {
					return
				}
				console_write("This server has the following information.\r\n" + server.Info_str())
				answer, aborted := console_read_confirm("Is this the server you wish to remove?\r\n")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
					return
				}
				c.Server_remove(server)
				console_write("Server removed.")
				return
			default:
				console_write("Unrecognized parameter: " + cmd)
				console_write(commands.HelpText("server"))
			}
		})

	commands.AddHelp("servers",
		"Lists the available servers.")
	commands.Add("servers",
		func(param string) {
			servers := c.Servers_read()
			if len(servers) == 0 {
				console_write("No servers available.")
			}
			count := len(servers)
			str := ""
			for _, server := range servers {
				str += server.DisplayName_read() + " (" + server.Host_read() + ":" + server.Tcpport_read() + ") "
				if !server.connected() {
					str += "(not connected)\r\n"
					continue
				}
				str += "(connected) "
				usrcount := len(server.Users_read())
				str += "(" + strconv.Itoa(usrcount) + " user"
				if usrcount != 1 {
					str += "s"
				}
				str += " connected) "
				chancount := len(server.Channels_read())
				str += "(" + strconv.Itoa(chancount) + " channel"
				if chancount != 1 {
					str += "s"
				}
				str += " found)\r\n"
			}
			prompt := strconv.Itoa(count) + " server"
			if count != 1 {
				prompt += "s are"
			} else {
				prompt += " is"
			}
			console_write(prompt + " available.\r\n" + str)
		})

	commands.AddHelp("message",
		"Sends a message to a channel or a user.",
		"message\r\nWill guide you through prompts.",
		"message tech\r\nWill prompt you for a message to send to tech.",
		"message tech Testing.\r\nWill send \"Testing.\" to tech.",
		"message \"/test channel/1\"\r\nWill prompt you for a message to send to channel 1 under a channel called test channel.",
		"message This is a test.\r\nWill send the message after guiding you through the prompts for a channel or user, unless you're in a channel already, in which case, the message will be sent to the channel you are in.")
	commands.Add("message",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to send message. Not connected.")
				return
			}
			var usr *tt_user
			var ch *tt_channel
			bot_usr := server.User_find_id(server.Uid_read())
			aborted := false
			dest := ""
			message := ""
			params := stringSeperateParam(param, " ", "\"")
			if len(params) >= 1 {
				dest = params[0]
			}
			usr, ch, aborted = server_menu_src(server, dest)
			if aborted {
				return
			}
			if len(params) >= 2 {
				if usr != nil || ch != nil {
					message = strings.Join(restoreParams(params[1:], " ", "\""), " ")
				} else {
					message = param
					if bot_usr != nil && bot_usr.Channel_read() == nil {
						usr, ch, aborted = server_menu_src(server, "")
						if aborted {
							return
						}
					} else {
						if bot_usr != nil {
							ch = bot_usr.Channel_read()
						}
					}
				}
			}
			if ch == nil && usr == nil {
				if message == "" {
					message = param
				}
				if bot_usr.Channel_read() != nil {
					ch = bot_usr.Channel_read()
				} else {
					usr, ch, aborted = server_menu_src(server, "")
				}
			}
			if message == "" {
				after := ""
				if usr != nil {
					after = usr.NickName_log()
				} else if ch != nil {
					after = ch.Path_read()
				}
				if after != "" {
					after = " Sending to " + after
				}
				msg, err := console_read_prompt("Please enter the message you want to send." + after)
				if err != nil {
					return
				}
				if msg == "" {
					console_write("Empty message unsupported. Aborted.")
					return
				}
				message = msg
			}
			res := false
			if usr != nil {
				if usr.Uid_read() == server.Uid_read() {
					console_write("Sending a message to the bot isn't supported.")
					return
				}
				res = server.cmd_message_user(usr.Uid_read(), message)
			} else if ch != nil {
				if bot_usr != nil && bot_usr.UserType_read() != TT_USERTYPE_ADMIN && bot_usr.Channel_read() != ch {
					console_write("You cannot send a message to a channel you aren't in. Aborted.")
					return
				}
				res = server.cmd_message_channel(ch.Id_read(), message)
			} else {
				console_write("Unable to send message. A user or channel wasn't selected.")
				return
			}
			if res {
				console_write("Command successful.")
			} else {
				console_write("Command unsuccessful.")
			}
		})

	commands.AddHelp("msg",
		"Same as message.")
	commands.Add("msg",
		func(param string) {
			commands.Exec("message", param)
		})

	commands.AddHelp("broadcast",
		"Will send a broadcast message.",
		"broadcast\r\nWill prompt you for a message to send.",
		"broadcast testing\r\nWill send testing as a broadcast message.")
	commands.Add("broadcast",
		func(message string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to send message. Not connected.")
				return
			}
			if !server.User_rights_check(TT_USERRIGHT_TEXT_MESSAGE_BROADCAST) {
				console_write("You cannot send broadcast messages.")
				return
			}
			var err error
			if message == "" {
				message, err = console_read_prompt("Please enter the message you want to send.")
				if err != nil {
					return
				}
				if message == "" {
					console_write("Empty message unsupported.")
					return
				}
			}
			res := server.cmd_message_broadcast(message)
			if res {
				console_write("Command successful.")
			} else {
				console_write("Command unsuccessful.")
			}
		})

	commands.AddHelp("join",
		"Will join a channel.",
		"join\r\nWill prompt you for a channel to join.",
		"join /\r\nWill join the root channel.")
	commands.Add("join",
		func(channel string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to join channel. Not connected.")
				return
			}
			ch, aborted := server_menu_channel(server, channel)
			if aborted {
				return
			}
			usr := server.User_find_id(server.Uid_read())
			if ch == nil {
				console_write("The channel " + channel + " doesn't exist.")
				return
			}
			if usr != nil {
				if usr.Channel_read() == ch {
					console_write("You are already in this channel.")
					return
				}
				if usr.Channel_read() != nil {
					cpath := usr.Channel_read().Path_read()
					answer, aborted := console_read_confirm("You are already in " + cpath + ". Would you like to leave this channel and join another?\r\n")
					if aborted {
						return
					}
					if !answer {
						console_write("Aborted.")
						return
					}
					server.cmd_leave()
				}
			}
			password := ""
			if ch.Protected_read() != 0 {
				password = ch.Password_read()
				if password == "" {
					var err error
					password, err = console_read_prompt("Please enter the channel password for " + ch.Path_read())
					if err != nil {
						return
					}
				}
			}
			res := server.cmd_join(ch.Id_read(), password)
			if res {
				console_write("Command successful.")
			} else {
				console_write("Command unsuccessful.")
			}
		})

	commands.AddHelp("leave",
		"Will leave a channel.")
	commands.Add("leave",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to leave channel. Not connected.")
				return
			}
			ch := server.User_find_id(server.Uid_read()).Channel_read()
			if ch == nil {
				console_write("You aren't in a channel.")
				return
			}
			res := server.cmd_leave()
			if res {
				console_write("Command successful.")
			} else {
				console_write("Command unsuccessful.")
			}
		})

	commands.AddHelp("move",
		"Move one user to a channel, or everyone in a channel to another channel.",
		"move\r\nWill guide you through prompts to move users.",
		"move / /away\r\nWill move all users in the root channel to the away channel.",
		"move tech /admin\r\nWill move tech to the admin channel.")
	commands.Add("move",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to move users. Not connected.")
				return
			}
			if !server.User_rights_check(TT_USERRIGHT_MOVE_USERS) {
				console_write("You don't have permission to move users. Command unsuccessful.")
				return
			}
			var usr_src *tt_user
			var ch_src *tt_channel
			var ch_dest *tt_channel
			aborted := false
			src := ""
			dest := ""
			params := stringSeperateParam(param, " ", "\"")
			if len(params) >= 1 {
				src = params[0]
			}
			if len(params) >= 2 {
				dest = strings.Join(restoreParams(params[1:], " ", "\""), " ")
			}
			usr_src, ch_src, aborted = server_menu_src(server, src)
			if aborted {
				return
			}
			ch_dest, aborted = server_menu_channel(server, dest)
			if aborted {
				return
			}
			if ch_dest != nil && ch_dest == server.Channel_find_id(server.AutoMoveFrom_read()) {
				answer, aborted := console_read_confirm("You are already automoving from " + ch_src.Path_read() + ". Would you like to disable automatic user moving and continue?")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
					return
				}
				server.AutoMove_clear()
				console_write("Automatic user moving disabled.")
			}
			if ch_src == nil && ch_dest == nil && usr_src == nil {
				console_write("Source and destination unselected for user moving. Command unsuccessful.")
				return
			}
			if usr_src != nil && ch_dest != nil {
				res := server.cmd_move_user(usr_src.Uid_read(), ch_dest.Id_read())
				if res {
					console_write(usr_src.NickName_log() + " moved to " + ch_dest.Path_read() + ".")
				} else {
					console_write("Command unsuccessful.")
				}
				return
			}
			if ch_src != nil && ch_dest != nil {
				count := 0
				users := ch_src.Users_read()
				cid := ch_dest.Id_read()
				for i, u := range users {
					if u == nil {
						console_write("Nil user found at " + strconv.Itoa(i))
						continue
					}
					if server.cmd_move_user(u.Uid_read(), cid) {
						count++
					}
				}
				if count == 0 {
					console_write("No users were moved.")
					return
				}
				prompt := "Successfully moved " + strconv.Itoa(count) + " user"
				if count != 1 {
					prompt += "s"
				}
				console_write(prompt + " to " + ch_dest.Path_read())
				return
			}
			prompt := ""
			if usr_src == nil || ch_src == nil {
				prompt += "Source unselected for user moving."
			}
			if ch_dest == nil {
				if prompt != "" {
					prompt += " "
				}
				prompt += "Destination unselected for user moving."
			}
			console_write("Command unsuccessful. " + prompt)
		})

	commands.AddHelp("nick",
		"Will allow you to change your nickname.",
		"nick\r\nWill prompt you for a nickname.",
		"nick test\r\nWill change your nickname to test.")
	commands.Add("nick",
		func(nick string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to change nickname. Not connected.")
				return
			}
			asked := false
			if nick == "" {
				asked = true
				if c.Server_prompt_nickname(server, true) {
					return
				}
				nick = server.NickName_read()
				if nick == "" && server.UseGlobalNickName_read() {
					nick = c.NickName_read()
				}
			}
			usr := server.User_find_id(server.Uid_read())
			if nick == usr.NickName_read() {
				console_write("Nicknames are identical. Your nickname is already " + nick + ".")
				return
			}
			res := server.cmd_changenick(nick)
			if !res {
				console_write("Command unsuccessful.")
				return
			}
			if !server.UseGlobalNickName_read() {
				server.NickName_set(nick)
			} else {
				if !asked {
					server.UseGlobalNickName_set(false)
					server.NickName_set(nick)
				}
			}
			c.Write()
			console_write("Command successful. Nickname changed to " + nick)
		})

	commands.AddHelp("status",
		"Updates or displays your status.",
		"status\r\nWill display your status.",
		"status online\r\nWill update your status to online.",
		"status away testing\r\nWill update your status to away with the message testing.")
	commands.Add("status",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to change nickname. Not connected.")
				return
			}
			usr := server.User_find_id(server.Uid_read())
			mode := usr.StatusMode_read_str()
			msg := usr.StatusMsg_read()
			if param == "" {
				output := ""
				if mode != "" {
					output += "Current status mode: " + mode
				}
				if msg != "" {
					output += "\r\nCurrent status message: " + msg
				}
				if output == "" {
					output = "No status information available."
				}
				console_write(output)
				return
			}
			params := strings.Split(param, " ")
			status_mode := 0
			status_mode_str := ""
			status_msg := ""
			if len(params) >= 1 {
				status_mode_str = strings.ToLower(params[0])
			}
			if len(params) >= 2 {
				status_msg = strings.Join(params[1:], " ")
			}
			if status_mode_str == "" {
				status_mode_str = TT_USERSTATUS_NONE_STR
			}
			if mode == status_mode_str && msg == status_msg {
				console_write("Status unchanged, already set to entered parameters.")
				return
			}
			switch status_mode_str {
			case TT_USERSTATUS_NONE_STR:
				status_mode = TT_USERSTATUS_NONE
			case TT_USERSTATUS_AWAY_STR:
				status_mode = TT_USERSTATUS_AWAY
			default:
				console_write("Unrecognized parameter.")
				console_write(commands.HelpText("status"))
				return
			}
			res := server.cmd_changestatus(status_mode, status_msg)
			if res {
				console_write("Command successful.")
				commands.Exec("status", "")
			} else {
				console_write("Command unsuccessful.")
			}
		})

	commands.AddHelp("who",
		"Will provide information about a user either entered or selected.",
		"who\r\nWill list all users to select one to obtain information about.",
		"who test\r\nWill obtain information on the user test, or if multiple users were found with test in their nicknames or usernames, will list them so you can make a selection.")
	commands.Add("who",
		func(user string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to obtain user information. Not connected.")
				return
			}
			usr, aborted := server_menu_user(server, user)
			if aborted {
				return
			}
			if usr == nil {
				console_write("No user selected. Aborted.")
				return
			}
			info := ""
			lnickname := usr.NickName_log()
			nickname := usr.NickName_read()
			info += "Displaying information for " + lnickname + "\r\n"
			if nickname == "" {
				info += "User has no nickname\r\n"
			}
			username := usr.UserName_read()
			if username != "" {
				info += "Username: " + username
			} else {
				info += "User has no username."
			}
			info += "\r\n"
			usertype := usr.UserType_read_str()
			if usertype != "" {
				info += "User type: " + usertype
			} else {
				info += "No user type available."
			}
			info += "\r\nUser ID: " + strconv.Itoa(usr.Uid_read()) + "\r\n"
			if usr.Conntime_isSet() {
				info += "Connected for " + usr.Conntime_read_str() + ".\r\n"
			}
			ip := usr.Ip_read()
			if ip != "" {
				info += "IP: " + ip + "\r\n"
			}
			version := usr.Version_read()
			if version != "" {
				info += "Client version: " + version + "\r\n"
			}
			clname := usr.ClientName_read()
			if clname != "" {
				info += "Client name: " + clname + "\r\n"
			}
			info += "Current local subscriptions: " + usr.Subscriptions_local_read_str() + "\r\n"
			info += "Current remote subscriptions: " + usr.Subscriptions_remote_read_str() + "\r\n"
			status_mode := usr.StatusMode_read_str()
			if status_mode != "" {
				info += "Current status mode: " + status_mode + "\r\n"
			}
			status_msg := usr.StatusMsg_read()
			if status_msg != "" {
				info += "Status message: " + status_msg + "\r\n"
			}
			ch := usr.Channel_read()
			if ch != nil {
				info += "Currently in channel " + ch.Path_read() + "\r\n"
			}
			console_write(info)
		})

	commands.AddHelp("version",
		"Displays the name, version number, and build time of the bot if available.")
	commands.Add("version",
		func(param string) {
			str := bot_name + ", version " + Version + "\r\n"
			if BuildTime != "" {
				btime, err := time.Parse("2006-01-02__15:04:05_(MST)", BuildTime)
				if err != nil {
					str += "Unable to retrieve build time.\r\nError: " + err.Error() + "\r\nThis is a bug, report it."
				} else {
					btime_local := btime.Local()
					ftime := btime_local.Format("Monday, January 2, 2006  03:04:05 PM (-0700 MST)")
					fwhen := time_duration_str(time.Since(btime)) + " ago"
					str += "Built " + ftime + ", which was " + fwhen + ".\r\n"
				}
			} else {
				str += "No build information available.\r\n"
			}
			console_write(str)
		})

	commands.AddHelp("vlist",
		"Sorts users on the active or selected server by version number from oldest to newest, or alphabetically, and displays them, along with their client name in parenthesis, if available.")
	commands.Add("vlist",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to obtain user information. Not connected.")
				return
			}
			users := server.Users_sort(server.Uid_read())
			versions_duplicates := []string{}
			for _, usr := range users {
				versions_duplicates = append(versions_duplicates, usr.Version_read())
			}
			sort.Strings(versions_duplicates)
			versions := []string{}
			duplicated := false
			for _, version := range versions_duplicates {
				duplicated = false
				for _, v_dup := range versions {
					if version == v_dup {
						duplicated = true
						break
					}
				}
				if !duplicated {
					versions = append(versions, version)
				}
			}
			str := ""
			prompt := strconv.Itoa(len(users)) + " user"
			if len(users) != 1 {
				prompt += "s"
			}
			prompt += " found.\r\n"
			count := 0
			vstring := ""
			for _, version := range versions {
				count = 0
				vstring = ""
				for _, usr := range users {
					if usr.Version_read() == version {
						vstring += usr.NickName_log()
						if clientname := usr.ClientName_read(); clientname != "" {
							vstring += " (" + clientname + ")"
						}
						vstring += ", "
						count++
					}
				}
				vstring = strings.TrimSuffix(vstring, ", ")
				str += strconv.Itoa(count) + " client"
				if count != 1 {
					str += "s"
				}
				str += " running " + version + ":\r\n" + vstring + "\r\n"
			}

			console_write(prompt + str)
		})

	commands.AddHelp("clear",
		"Clears the output console of all data. This command will not notify you of success.")
	commands.Add("clear",
		func(param string) {
			console_clear()
		})

	commands.AddHelp("ping",
		"Will ping a single server you enter as a parameter, or all servers. Reported results will be estimates only, and shouldn't be taken as completely accurate.",
		"ping test\r\nWill ping the test server and return the time taken.",
		"ping\r\nWill ping each server and return statistics.")
	commands.Add("ping",
		func(param string) {
			ping_times := 4
			if param != "" {
				servers := c.Server_find_name(param)
				if len(servers) == 0 {
					console_write("Unable to ping " + param + ". Server doesn't exist.")
					return
				}
				server := servers[0]
				if !server.connected() {
					console_write("Unable to ping " + param + ". Not connected.")
					return
				}
				console_write("Pinging " + param + ", please wait.")
				msecs_min := 0
				msecs_max := 0
				msecs_avg := 0
				for i := 1; i <= ping_times; i++ {
					start := time.Now()
					server.cmd_ping()
					end := time.Since(start)
					msecs := int(end.Nanoseconds()) / int(time.Millisecond) / 3
					msecs_avg += msecs
					if i == 1 {
						msecs_min = msecs
						msecs_max = msecs
						continue
					}
					if msecs > msecs_max {
						msecs_max = msecs
					}
					if msecs < msecs_min {
						msecs_min = msecs
					}
				}
				msecs_total := msecs_avg
				msecs_avg = int(msecs_avg / ping_times)
				msg := "Ping complete.\r\nMinimum milliseconds: " + strconv.Itoa(msecs_min) + "\r\nMaximum milliseconds: " + strconv.Itoa(msecs_max) + "\r\nAverage milliseconds: " + strconv.Itoa(msecs_avg) + "\r\nTotal milliseconds: " + strconv.Itoa(msecs_total)
				console_write(msg)
				return
			}
			if len(c.Servers_read()) == 0 {
				console_write("No servers are available to ping.")
				return
			}
			go servers_ping(ping_times)
		})

	commands.AddHelp("subscriptions",
		"Manage a users subscriptions.",
		"subscriptions test\r\nWill manage the subscriptions for the user test.",
		"subscriptions\r\nWill present a list of users to select to manage the subscriptions of.")
	commands.Add("subscriptions",
		func(user string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to manage user subscriptions. Not connected.")
				return
			}
			usr, aborted := server_menu_user(server, user)
			if aborted {
				return
			}
			if usr == nil {
				console_write("No user selected. Aborted.")
				return
			}
			current_subs := usr.Subscriptions_local_read()
			console_write("Modifying subscriptions for " + usr.NickName_log() + ".\r\nCurrent local subscriptions: " + usr.Subscriptions_local_read_str())
			new_subs, aborted := teamtalk_flags_subscriptions_menu(current_subs)
			if aborted {
				return
			}
			if current_subs == new_subs {
				console_write("Subscriptions for " + usr.NickName_log() + " unchanged.\r\nCurrent local subscriptions: " + usr.Subscriptions_local_read_str())
				return
			}
			res := server.cmd_changesubscriptions(usr.Uid_read(), new_subs)
			if !res {
				console_write("Command unsuccessful.")
				return
			}
			sub_msg := ""
			sub_added_str := usr.Subscriptions_local_added_str(current_subs)
			sub_removed_str := usr.Subscriptions_local_removed_str(current_subs)
			if sub_added_str != "" {
				sub_msg += "Subscriptions added: " + sub_added_str + "\r\n"
			}
			if sub_removed_str != "" {
				sub_msg += "Subscriptions removed: " + sub_removed_str + "\r\n"
			}
			if sub_msg != "" {
				console_write("Local subscriptions changed for " + usr.NickName_log() + ".\r\n" + sub_msg)
				return
			}
			console_write("Local subscriptions for " + usr.NickName_log() + " unchanged.")
		})

	commands.AddHelp("sub",
		"Same as subscriptions.")
	commands.Add("sub",
		func(user string) {
			commands.Exec("subscriptions", user)
		})

	commands.AddHelp("subs",
		"Same as subscriptions.")
	commands.Add("subs",
		func(user string) {
			commands.Exec("subscriptions", user)
		})

	commands.AddHelp("autosubscriptions",
		"Manages automatic user subscriptions for the active or selected server.\r\nTo disable automatic subscribing and unsubscribing, disable all subscriptions in the menu.\r\nPlease note that, if you begin enabling subscriptions in the menu, anything disabled will automatically be unsubscribed from when a user connects. For example, if you want to automatically intercept private messages from a user, so you enable that subscription and complete your selection, you will be subscribed to intercept users private messages, but unsubscribed from everything else, which would prevent you from receiving any channel or broadcast messages from them.")
	commands.Add("autosubscriptions",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if c.Server_prompt_autosubscriptions(server, true) {
				return
			}
			console_write("Command complete.")
		})

	commands.AddHelp("autosubs",
		"Same as autosubscriptions.")
	commands.Add("autosubs", func(param string) {
		commands.Exec("autosubscriptions", param)
	})

	commands.AddHelp("autosub",
		"Same as autosubscriptions.")
	commands.Add("autosub", func(param string) {
		commands.Exec("autosubscriptions", param)
	})

	commands.AddHelp("automove",
		"Automatically move all users that connect to the server, or all users in a channel, to another channel.",
		"automove\r\nWill guide you through prompts to automatically move users.",
		"automove / /away\r\nWill configure automatic user moving for users that join the root channel to the away channel.",
		"automove /admin\r\nWill configure automatic user moving of all users that connect to the server to the admin channel.",
		"automove disable\r\nautomove off\r\nWill disable automoving if enabled.",
		"automove\r\nWill report on the status of automatic user moving, and if enabled, will give you the option to disable it. You will also get prompts to select a source and/or destination channel for automoving if you choose.")
	commands.Add("automove",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			if !server.connected() {
				console_write("Unable to set up automatic user moving. Not connected.")
				return
			}
			var ch_src *tt_channel
			var ch_dest *tt_channel
			answer := false
			aborted := false
			src := ""
			dest := ""
			if param == "" {
				if !server.User_rights_check(TT_USERRIGHT_MOVE_USERS) {
					console_write("Automove settings disabled. Insufficient user rights to enable.")
					return
				}
				msg := "Automove status: "
				if server.AutoMove_enabled() {
					msg += "enabled"
				} else {
					msg += "disabled"
				}
				msg += ".\r\n"
				ch_src = server.Channel_find_id(server.AutoMoveFrom_read())
				ch_dest = server.Channel_find_id(server.AutoMoveTo_read())
				if ch_src != nil || ch_dest != nil {
					msg += "Automatically moving "
					if ch_src != nil {
						msg += "users from " + ch_src.Path_read()
					} else {
						msg += "all users that connect"
					}
					msg += " to "
					if ch_dest != nil {
						msg += ch_dest.Path_read() + "\r\n"
					} else {
						console_write("Incorrect automove settings found. Disabling.")
						server.AutoMove_clear()
						ch_src = nil
						ch_dest = nil
						msg = ""
					}
				}
				if msg != "" {
					console_write(msg)
				}
				if ch_src != nil || ch_dest != nil {
					answer, aborted = console_read_confirm("Would you like to disable automatic user moving?\r\n")
					if aborted {
						return
					}
					if !answer {
						console_write("Automove settings will remain the same until disabled.")
						return
					}
					server.AutoMove_clear()
					console_write("Automove settings disabled.")
					ch_src = nil
					ch_dest = nil
				}
				if ch_src == nil && ch_dest == nil {
					answer, aborted = console_read_confirm("Would you like to set up automatic user moving now?\r\n")
					if aborted {
						return
					}
					if !answer {
						console_write("Aborted.")
						return
					}
					ch_src, aborted = server_menu_channel(server, "")
					if aborted {
						return
					}
					if ch_src == nil {
						console_write("Channel not found. Aborted.")
						return
					}
					console_write("Channel selected: " + ch_src.Path_read())
					answer, aborted = console_read_confirm("Would you like to move all users that connect to the server to this channel?\r\n")
					if aborted {
						return
					}
					if answer {
						ch_dest = ch_src
						ch_src = nil
					} else {
						console_write("Selecting destination channel.")
						ch_dest, aborted = server_menu_channel(server, "")
						if aborted {
							return
						}
						if ch_dest == nil {
							console_write("Channel not found. Aborted.")
							return
						}
					}
				}
			} else {
				if !server.User_rights_check(TT_USERRIGHT_MOVE_USERS) {
					console_write("You don't have permission to move users. Unable to set up automatic user moving. Command unsuccessful.")
					return
				}
				switch strings.ToLower(param) {
				case "off", "disable":
					if !server.AutoMove_enabled() {
						console_write("Automatic user moving already disabled.")
						return
					}
					server.AutoMove_clear()
					console_write("Automatic user moving disabled.")
					return
				}
				params := stringSeperateParam(param, " ", "\"")
				if len(params) >= 1 {
					src = params[0]
				}
				if len(params) >= 2 {
					dest = strings.Join(restoreParams(params[1:], " ", "\""), " ")
				}
				ch_src, aborted = server_menu_channel(server, src)
				if aborted {
					return
				}
				if len(params) == 1 {
					if ch_src == nil {
						console_write("Source and/or destination channel not found. Aborted.")
						return
					}
					ch_dest = ch_src
					ch_src = nil
				}
				if ch := server.Channel_find_id(server.AutoMoveFrom_read()); ch != nil {
					answer, aborted := console_read_confirm("You are already automoving from " + ch.Path_read() + ". Would you like to disable automatic user moving and continue?")
					if aborted {
						return
					}
					if !answer {
						console_write("Aborted.")
						return
					}
					server.AutoMove_clear()
					console_write("Automatic user moving disabled.")
				}
				if ch_dest == nil {
					if ch := server.Channel_find_id(server.AutoMoveTo_read()); ch != nil {
						answer, aborted := console_read_confirm("You are already automatically moving users to " + ch.Path_read() + ". Would you like to disable automatic user moving and continue?\r\n")
						if aborted {
							return
						}
						if !answer {
							console_write("Aborted.")
							return
						}
						server.AutoMove_clear()
						console_write("Automatic user moving disabled.")
						if dest == "" {
							console_write("Selecting destination channel for automatic user moving.")
						}
					}
					ch_dest, aborted = server_menu_channel(server, dest)
					if aborted {
						return
					}
				}
			}
			if ch_dest == nil {
				console_write("Destination unselected for automatic user moving. Command unsuccessful.")
				return
			}
			if ch := server.Channel_find_id(server.AutoMoveTo_read()); ch != nil {
				answer, aborted := console_read_confirm("You are already automatically moving users to " + ch.Path_read() + ". Would you like to disable automatic user moving and continue?\r\n")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
					return
				}
				server.AutoMove_clear()
				console_write("Automatic user moving disabled.")
			}
			if ch_src == nil && ch_dest != nil {
				answer, aborted = console_read_confirm("You have selected to move all users that connect to the server to " + ch_dest.Path_read() + ". Is this correct?")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
					return
				}
				server.AutoMoveTo_set(ch_dest.Id_read())
				answer, aborted = console_read_confirm("Would you like these settings to persist until you disable them? If you don't set this, you will have to set up automatic user moving again if the program terminates or you quit the program, then restart it.\r\n")
				if aborted {
					return
				}
				if answer {
					server.AutoMoveTo_config_set(ch_dest.Id_read())
					if c.Write() {
						console_write("Automatic move settings saved to configuration file.")
					}
				}
				console_write("Users will be moved to " + ch_dest.Path_read() + " as they connect.")
				return
			}
			if ch_src != nil && ch_dest != nil {
				answer, aborted = console_read_confirm("You have selected to move users from " + ch_src.Path_read() + " to " + ch_dest.Path_read() + ". Is this correct?")
				if aborted {
					return
				}
				if !answer {
					console_write("Aborted.")
					return
				}
				server.AutoMoveFrom_set(ch_src.Id_read())
				server.AutoMoveTo_set(ch_dest.Id_read())
				answer, aborted = console_read_confirm("Would you like these settings to persist until you disable them? If you don't set this, you will have to set up automatic user moving again if the program terminates or you quit the program, then restart it.\r\n")
				if aborted {
					return
				}
				if answer {
					server.AutoMoveFrom_config_set(ch_src.Id_read())
					server.AutoMoveTo_config_set(ch_dest.Id_read())
					if c.Write() {
						console_write("Automatic move settings saved to configuration file.")
					}
				}
				answer, aborted = console_read_confirm("Would you like to move the users now?\r\n")
				if aborted {
					return
				}
				if !answer {
					console_write("Users will be moved as they join " + ch_src.Path_read())
					return
				}
				count := 0
				users := ch_src.Users_read()
				cid := ch_dest.Id_read()
				for i, u := range users {
					if u == nil {
						console_write("Nil user found at " + strconv.Itoa(i))
						continue
					}
					if server.cmd_move_user(u.Uid_read(), cid) {
						count++
					}
				}
				if count == 0 {
					console_write("No users were moved.")
					return
				}
				prompt := "Successfully moved " + strconv.Itoa(count) + " user"
				if count != 1 {
					prompt += "s"
				}
				console_write(prompt + " to " + ch_dest.Path_read())
				return
			}
		})

	commands.AddHelp("history",
		"Will display a history of logged events that occurred on the active or selected server.",
		"history\r\nDisplays the history for the active server.",
		"history test1 test2\r\nDisplays history for test1 and test2.")
	commands.Add("history",
		func(param string) {
			params := stringSeperateParam(param, " ", "\"")
			if len(params) > 0 {
				for _, param := range params {
					servers := c.Server_find_name(param)
					if len(servers) == 0 {
						console_write("Unable to display events for " + param + ". The server doesn't exist.")
						continue
					}
					server := servers[0]
					history := server.Log_history_read()
					if len(history) == 0 {
						console_write("History of logged events unavailable for " + server.DisplayName_read() + ".")
						continue
					}
					msg := "Displaying history of " + strconv.Itoa(len(history)) + " event"
					if len(history) != 1 {
						msg += "s"
					}
					msg += " for " + server.DisplayName_read() + ":\r\n" + strings.Join(history, "\r\n")
					console_write(msg)
				}
				return
			}
			server := server_active_check("")
			if server == nil {
				return
			}
			history := server.Log_history_read()
			if len(history) == 0 {
				console_write("History of logged events unavailable for " + server.DisplayName_read() + ".")
				return
			}
			msg := "Displaying history of " + strconv.Itoa(len(history)) + " event"
			if len(history) != 1 {
				msg += "s"
			}
			msg += " for " + server.DisplayName_read() + ":\r\n" + strings.Join(history, "\r\n")
			console_write(msg)
		})

	commands.AddHelp("date",
		"Displays the time and date.")
	commands.Add("date",
		func(param string) {
			console_write(time.Now().Format("Monday, January 2, 2006  03:04:05 PM (-0700 MST)"))
		})

	commands.AddHelp("time",
		"Same as the date command.")
	commands.Add("time",
		func(param string) {
			commands.Exec("date", param)
		})

	commands.AddHelp("account",
		"Add, change, or remove user accounts.",
		"account add\r\nWill guide you through prompts to add a user account.",
		"account add test test\r\nWill add a user account with the username and password of test.",
		"account change\r\nWill give you a list of accounts to modify.",
		"account change test\r\nWill provide options and prompts to change the account test.",
		"account change test check\r\nWill change the password of test to check.",
		"account mod\r\naccount modify\r\nSame as accounts change.",
		"account del\r\nWill give you a list of accounts to delete.",
		"account del test\r\nWill delete the account named test.",
		"account delete\r\naccount remove\r\nSame as account del.")
	commands.Add("account",
		func(param string) {
			server := server_active_check("")
			if server == nil {
				return
			}
			cmd := ""
			params := strings.Split(param, " ")
			if len(params) >= 1 {
				cmd = params[0]
			}
			if cmd == "" {
				menu := []string{
					"Add a user account",
					"Change a user account",
					"Remove a user account",
				}
				res, aborted := console_read_menu("Please select your option.\r\n", menu)
				if aborted || res == -1 {
					return
				}
				cmd = strings.Split(menu[res], " ")[0]
			}
			switch strings.ToLower(cmd) {
			case "add":
				cmd_params := stringSeperateParam(strings.Join(params[1:], " "), " ", "\"")
				username := ""
				password := ""
				usertype := ""
				if len(cmd_params) >= 1 {
					username = cmd_params[0]
				}
				if len(cmd_params) >= 2 {
					password = cmd_params[1]
				}
				if len(cmd_params) >= 3 {
					usertype = cmd_params[2]
				}
				server.Account_add_prompt(username, password, usertype)
				return
			case "change", "mod", "modify":
				// cmd_params := stringSeperateParam(strings.Join(params[1:], " "), " ", "\"")

				console_write("Command not implemented.")
				return
			case "del", "delete", "remove":
				console_write("Command not implemented.")
				return
			case "update":
				res := server.cmd_list_accounts()
				if res {
					console_write("Command successful.")
				} else {
					console_write("Command unsuccessful.")
				}
				return
			}
		})

	commands.AddHelp("panic",
		"Initiates a runtime panic. You probably shouldn't use this.")
	commands.Add("panic",
		func(param string) {
			panic("This is a user initiated panic. Well done for initiating this panic attack! You have shut down everything in an unsafe manner with the exception of the console, which will close itself to prevent strange things from occurring. Goodbye!")
		})
}
