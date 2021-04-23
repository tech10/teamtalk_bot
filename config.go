package main

import (
	"encoding/xml"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type config struct {
	sync.Mutex
	XMLName                    xml.Name `xml:"config"`
	Nickname                   string   `xml:"NickName,omitempty"`
	DisplayTimestamp           bool     `xml:"displayEventTimestamp"`
	ActiveServer               string   `xml:"ActiveServer,omitempty"`
	AutoConnectOnStart         bool     `xml:"defaults>autoConnectOnStart"`
	AutoConnectOnDisconnect    bool     `xml:"defaults>autoConnectOnDisconnect"`
	AutoConnectOnKick          bool     `xml:"defaults>autoConnectOnKick"`
	kicked                     bool
	AutoSubscriptions          int          `xml:"defaults>automaticSubscriptions,omitempty"`
	DisplayExtendedConnInfo    bool         `xml:"defaults>displayExtendedConnInfo"`
	DisplayStatusUpdates       bool         `xml:"defaults>displayStatusUpdates"`
	DisplaySubscriptionUpdates bool         `xml:"defaults>displaySubscriptionUpdates"`
	DisplayEvents              bool         `xml:"defaults>displayServerEventsIfInactive"`
	BeepOnCriticalEvents       bool         `xml:"defaults>beepOnCriticalServerEvents"`
	LogEvents                  bool         `xml:"defaults>logServerEvents"`
	LogEventsAccount           bool         `xml:"defaults>logServerEventsPerUserAccount"`
	UseGlobalNickName          bool         `xml:"defaults>useGlobalNickName"`
	UseDefaults                bool         `xml:"defaults>useOnServerCreate"`
	Servers                    []*tt_server `xml:"servers>server,omitempty"`
	cfile                      string
	timestamp_console          string
	logged_console             string
	wg                         sync.WaitGroup
	server                     *tt_server
}

var c *config

func NewConfig(fname string) *config {
	newconf := &config{
		cfile: fname,
	}
	f, err := os.Open(fname)
	if err != nil {
		res := newconf.Write()
		if res == false {
			console_close()
			os.Exit(1)
		}
		console_write("Creating configuration file " + fname)
		f.Close()
		if newconf.Init_prompt() {
			os.Remove(fname)
			console_close()
			os.Exit(1)
		}
		return newconf
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	err = dec.Decode(newconf)
	if err != nil {
		console_write("Error reading XML data in configuration file: " + err.Error())
		console_close()
		os.Exit(1)
	}
	console_write("Loaded configuration from " + fname)
	return newconf
}

func (conf *config) Write() bool {
	defer conf.Unlock()
	conf.Lock()
	fname := conf.cfile
	f, err := os.Create(fname)
	if err != nil {
		console_write("Error creating configuration file " + fname + ": " + err.Error())
		return false
	}
	defer time.Sleep(time.Millisecond * 50)
	defer f.Close()
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	err = enc.Encode(conf)
	if err != nil {
		console_write("Error writing xml: " + err.Error())
		return false
	}
	return true
}

func (conf *config) NickName_read() string {
	defer conf.Unlock()
	conf.Lock()
	return conf.Nickname
}

func (conf *config) NickName_set(name string) {
	if name == conf.NickName_read() {
		return
	}
	conf.Lock()
	conf.Nickname = name
	conf.Unlock()
	conf.Write()
}

func (conf *config) Servers_read() []*tt_server {
	defer conf.Unlock()
	conf.Lock()
	return conf.Servers
}

func (conf *config) Server_exists(server *tt_server) bool {
	defer conf.Unlock()
	conf.Lock()
	for _, cs := range conf.Servers {
		if cs == server {
			return true
		}
	}
	return false
}

func (conf *config) Server_add(server *tt_server) bool {
	if conf.Server_exists(server) {
		return false
	}
	conf.Lock()
	conf.Servers = append(conf.Servers, server)
	conf.Unlock()
	return conf.Write()
}

func (conf *config) Server_remove(server *tt_server) bool {
	if !conf.Server_exists(server) {
		return false
	}
	servers := conf.Servers_read()
	index := -1
	for i, sc := range servers {
		if sc != server {
			continue
		}
		index = i
		break
	}
	if index == -1 {
		return false
	}
	conf.Lock()
	copy(conf.Servers[index:], conf.Servers[index+1:])
	conf.Servers, conf.Servers[len(conf.Servers)-1] = append(conf.Servers[:index], conf.Servers[index+1:]...), nil
	conf.Unlock()
	server.Shutdown()
	return conf.Write()
}

func (conf *config) DisplayTimestamp_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.DisplayTimestamp
}

func (conf *config) DisplayTimestamp_set(displayTimestamp bool) {
	if displayTimestamp == conf.DisplayTimestamp_read() {
		return
	}
	conf.Lock()
	conf.DisplayTimestamp = displayTimestamp
	conf.Unlock()
	conf.Write()
}

func (conf *config) Log_timestamp_init() string {
	return time.Now().Format("2006-01-02  15:04:05 (-0700 MST)")
}

func (conf *config) Logged_console_read() string {
	defer conf.Unlock()
	conf.Lock()
	return conf.logged_console
}

func (conf *config) Logged_console_set(str string) {
	defer conf.Unlock()
	conf.Lock()
	conf.logged_console = str
}

func (conf *config) Timestamp_console_read() string {
	defer conf.Unlock()
	conf.Lock()
	return conf.timestamp_console
}

func (conf *config) Timestamp_console_set(str string) {
	defer conf.Unlock()
	conf.Lock()
	conf.timestamp_console = str
}

func conf_init(cname string) {
	console_write("Welcome to the " + bot_name + ", version " + Version + ".")
	if wd != "" {
		err := os.Chdir(wd)
		if err != nil {
			console_write("Error changing working directory to " + wd + ".\r\n" + err.Error())
			console_close()
			os.Exit(1)
		}
	}
	wdt, wdterr := os.Getwd()
	if wdterr == nil && wdt != "" {
		usr, usrerr := user.Current()
		if usrerr != nil {
			wd = wdt
		} else {
			usrhome := usr.HomeDir
			if usrhome != "" && wdt == usrhome {
				console_write("WARNING: Current working directory and user home directory are identical:\r\n" + wdt + "\r\nOperation not permitted. Attempting home directory change.")
				wd = wdt + string(filepath.Separator) + bot_name
				if cderr := dir_create(wd); cderr != nil {
					console_write("Home directory change failed.\r\n" + cderr.Error())
					console_close()
					os.Exit(1)
				}
				cherr := os.Chdir(wd)
				if cherr != nil {
					console_write("Home directory change failed.\r\n" + cherr.Error())
					console_close()
					os.Exit(1)
				}
				console_write("Home directory change successful.")
			} else {
				wd = wdt
			}
		}
	}
	console_write("Current working directory:\r\n" + wd)
	c = NewConfig(cname)
	servers := c.Servers_read()
	if len(servers) != 0 {
		//Inicial hostname resolution for duplicate server checking.
		//Check the display name here, too, as it can't be empty.
		for _, server := range servers {
			msg := ""
			if server.DisplayName_read() == "" {
				if server.Host_read() != "" || server.Tcpport_read() != "" {
					msg = "Error: the server connecting to " + server.Host_read() + ":" + server.Tcpport_read() + " has no display name."
				} else {
					msg = "Error: a server is missing connection information and a display name."
				}
			} else {
				if server.Host_read() == "" || server.Tcpport_read() == "" {
					msg = "Error: the server called " + server.DisplayName_read() + " has no connection information."
				}
			}
			if msg != "" {
				console_write(msg + " Operation not permitted.")
				if server.DisplayName_read() == "" {
					if c.Server_prompt_displayname(server, true) {
						console_close()
						os.Exit(1)
					}
					c.Write()
				}
				if server.Host_read() == "" || server.Tcpport_read() == "" {
					if c.Server_prompt_conn_info(server, true) {
						console_close()
						os.Exit(1)
					}
					c.Write()
				}
			}
			server.resolve()
		}
		//Check for things like duplicates here.
		if len(servers) != 1 {
			//Check for a duplicate server by name.
		name:
			for {
				for _, server := range c.Servers_read() {
					name := server.DisplayName_read()
					servers := c.Server_find_name(name)
					if len(servers) == 1 {
						continue
					}
					aborted := c.Duplicate_servers_by_name(servers)
					if aborted {
						console_close()
						os.Exit(1)
					}
					continue name
				}
				break name
			}
			//Check for a duplicate server by its information.
		info:
			for {
				for _, server := range c.Servers_read() {
					ip := server.Ip_read()
					port := server.Tcpport_read()
					servers := c.Server_find_info(ip, port)
					if len(servers) <= 1 {
						continue
					}
					aborted := c.Duplicate_servers_by_info(servers)
					if aborted {
						console_close()
						os.Exit(1)
					}
					continue info
				}
				break info
			}
		}
		//If configured, set the active server.
		name := c.Server_active_read_name()
		if name != "" {
			if servers := c.Server_find_name(name); len(servers) != 0 {
				c.Server_active_set(servers[0])
			} else {
				console_write("Notice: unable to find server " + name + ". Clearing active server.")
				c.Server_active_clear()
			}
		}
		return
	}
	if len(c.Servers_read()) == 0 {
		if c.Init_servers_prompt() {
			console_close()
			os.Exit(1)
		}
	}
}

func (conf *config) Init_servers_prompt() bool {
	console_write("There are currently no configured servers, and at least one is required.")
	return conf.Server_add_prompt()
}

func (conf *config) Init_prompt() bool {
	answer, aborted := console_read_confirm("Each server can individually have its own nickname set if desired, and such nicknames will be used in place of the global nickname used across all servers without a nickname set. Do you wish to set the global nickname?")
	if aborted {
		return true
	}
	if !answer {
		console_write("No global nickname set.")
	} else {
		if conf.NickName_prompt() {
			return true
		}
	}
	if conf.DisplayTimestamp_prompt() {
		return true
	}
	if conf.Defaults_prompt() {
		return true
	}
	if conf.Init_servers_prompt() {
		return true
	}
	return false
}

func (conf *config) Defaults_prompt() bool {
	console_write("Setting default options for server creation.")
	if conf.AutoConnectInfo_prompt() {
		return true
	}
	if conf.AutoSubscriptions_prompt() {
		return true
	}
	if conf.EventsInfo_prompt() {
		return true
	}
	if conf.LogInfo_prompt() {
		return true
	}
	if conf.UseDefaults_prompt() {
		return true
	}
	return false
}

func (conf *config) AutoConnectInfo_prompt() bool {
	autoConnectOnStart, aborted := console_read_confirm("Would you like to automatically connect to servers when the program starts?\r\n")
	if aborted {
		return true
	}
	conf.AutoConnectOnStart_set(autoConnectOnStart)
	autoConnectOnDisconnect, aborted := console_read_confirm("Would you like to automatically reconnect to servers when the connection is lost?\r\n")
	if aborted {
		return true
	}
	conf.AutoConnectOnDisconnect_set(autoConnectOnDisconnect)
	autoConnectOnKick, aborted := console_read_confirm("Would you like to automatically reconnect to servers when you are kicked?\r\n")
	if aborted {
		return true
	}
	conf.AutoConnectOnKick_set(autoConnectOnKick)
	return false
}

func (conf *config) AutoSubscriptions_prompt() bool {
	subs_answer, aborted := console_read_confirm("Would you like to set the default automatic user subscriptions?\r\n")
	if aborted {
		return true
	}
	if !subs_answer {
		console_write("Aborted.")
		return false
	}
	current_subs := conf.AutoSubscriptions_read()
	current_subs_str := conf.AutoSubscriptions_read_str()
	msg := "Modifying default automatic local subscriptions.\r\n"
	if current_subs_str != "" {
		msg += "Current default automatic local subscriptions: " + current_subs_str + "\r\n"
	} else {
		msg += "Default automatic local subscriptions disabled.\r\n"
	}
	console_write(msg)
	new_subs, aborted := teamtalk_flags_subscriptions_menu(current_subs)
	if aborted {
		return true
	}
	if new_subs == current_subs {
		console_write("Default automatic local subscriptions unchanged.")
		return false
	}
	conf.AutoSubscriptions_set(new_subs)
	return false
}

func (conf *config) EventsInfo_prompt() bool {
	conn_info, aborted := console_read_confirm("Would you like to display extended connection information from users on servers, such as their IP address, user id, client name and version, etc?\r\n")
	if aborted {
		return true
	}
	conf.DisplayExtendedConnInfo_set(conn_info)
	status, aborted := console_read_confirm("Would you like to receive status updates from users on servers?\r\n")
	if aborted {
		return true
	}
	conf.DisplayStatusUpdates_set(status)
	substatus, aborted := console_read_confirm("Would you like to receive subscription updates from users on servers?\r\n")
	if aborted {
		return true
	}
	conf.DisplaySubscriptionUpdates_set(substatus)
	display_events, aborted := console_read_confirm("Would you like to display server events for a server that isn't active?\r\n")
	if aborted {
		return true
	}
	conf.DisplayEvents_set(display_events)
	beep_events, aborted := console_read_confirm(TT_BEEP + "Would you like to hear a beep on critical events from servers like the one just sent to the console window?\r\n")
	if aborted {
		return true
	}
	conf.BeepOnCriticalEvents_set(beep_events)
	return false
}

func (conf *config) LogInfo_prompt() bool {
	log_events, aborted := console_read_confirm("Would you like to log server events in a log file?\r\n")
	if aborted {
		return true
	}
	conf.LogEvents_set(log_events)
	log_events_account := false
	if log_events {
		log_events_account, aborted = console_read_confirm("Would you like to log each event from a particular user in their own log file matching their username on the server?\r\n")
		if aborted {
			return true
		}
	}
	conf.LogEventsAccount_set(log_events_account)
	return false
}

func (conf *config) UseGlobalNickName_prompt() bool {
	if conf.NickName_read() == "" {
		conf.UseGlobalNickName_set(false)
		return false
	}
	answer, aborted := console_read_confirm("Do you wish to use the global nickname by default on server creation?\r\n")
	if aborted {
		return true
	}
	conf.UseGlobalNickName_set(answer)
	return false
}

func (conf *config) UseDefaults_prompt() bool {
	answer, aborted := console_read_confirm("Do you wish to use the default options on server creation when no values are entered?\r\n")
	if aborted {
		return true
	}
	conf.UseDefaults_set(answer)
	return false
}

func (conf *config) SetDefaultValue_prompt() (bool, bool) {
	return console_read_confirm("Would you like to set the default option to this value?\r\n")
}

func (conf *config) SetDefaultValues_prompt() (bool, bool) {
	return console_read_confirm("Would you like to set the default options to these values?\r\n")
}

func (conf *config) Servers_AutoConnectOnStart_prompt() bool {
	autoConnectOnStart, aborted := console_read_confirm("Would you like to automatically connect to all servers when the program starts?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.AutoConnectOnStart_set(autoConnectOnStart)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.AutoConnectOnStart_set(autoConnectOnStart)
	}
	return false
}

func (conf *config) Servers_AutoConnectOnDisconnect_prompt() bool {
	autoConnectOnDisconnect, aborted := console_read_confirm("Would you like to automatically reconnect to all servers when the connection is lost?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.AutoConnectOnDisconnect_set(autoConnectOnDisconnect)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.AutoConnectOnDisconnect_set(autoConnectOnDisconnect)
	}
	return false
}

func (conf *config) Servers_AutoConnectOnKick_prompt() bool {
	autoConnectOnKick, aborted := console_read_confirm("Would you like to automatically reconnect to all servers when you are kicked?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.AutoConnectOnKick_set(autoConnectOnKick)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.AutoConnectOnKick_set(autoConnectOnKick)
	}
	return false
}

func (conf *config) Servers_AutoSubscriptions_prompt() bool {
	subs_answer, aborted := console_read_confirm("Would you like to set automatic user subscriptions for all servers?\r\n")
	if aborted {
		return true
	}
	if !subs_answer {
		console_write("Aborted.")
		return false
	}
	current_subs := conf.AutoSubscriptions_read()
	current_subs_str := conf.AutoSubscriptions_read_str()
	msg := "Modifying subscriptions for all servers.\r\n"
	if current_subs_str != "" {
		msg += "Current default automatic local subscriptions: " + current_subs_str + "\r\n"
	} else {
		msg += "Default automatic local subscriptions disabled.\r\n"
	}
	console_write(msg)
	new_subs, aborted := teamtalk_flags_subscriptions_menu(current_subs)
	if aborted {
		return true
	}
	if new_subs != 0 {
		update_subs, aborted := console_read_confirm("Would you like to update the subscription settings for available users on connected servers?\r\n")
		if aborted {
			return true
		}
		console_write("Updating subscriptions in a background task.")
		go func(new_subs int) {
			updated_servers := 0
			updated_users := 0
			for _, server := range conf.Servers_read() {
				server.AutoSubscriptions_set(new_subs)
				if update_subs {
					if !server.connected() {
						continue
					}
					updated_servers++
					for _, usr := range server.Users_sort(0) {
						if new_subs == usr.Subscriptions_local_read() {
							continue
						}
						if server.cmd_changesubscriptions(usr.Uid_read(), new_subs) {
							updated_users++
						}
					}
				}
			}
			if updated_servers != 0 {
				msg := "Local subscriptions updated for " + strconv.Itoa(updated_users) + " user"
				if updated_users != 1 {
					msg += "s"
				}
				msg += " on " + strconv.Itoa(updated_servers) + " server"
				if updated_servers != 1 {
					msg += "s"
				}
				console_write(msg + ".")
			}
		}(new_subs)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.AutoSubscriptions_set(new_subs)
	}
	return false
}

func (conf *config) Servers_DisplayExtendedConnInfo_prompt() bool {
	conn_info, aborted := console_read_confirm("Would you like to display extended connection information from the users on all servers, such as their IP address, user id, client name and version, etc?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.DisplayExtendedConnInfo_set(conn_info)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.DisplayExtendedConnInfo_set(conn_info)
	}
	return false
}

func (conf *config) Servers_DisplayStatusUpdates_prompt() bool {
	status, aborted := console_read_confirm("Would you like to receive status updates from the users on all servers?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.DisplayStatusUpdates_set(status)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.DisplayStatusUpdates_set(status)
	}
	return false
}

func (conf *config) Servers_DisplaySubscriptionUpdates_prompt() bool {
	substatus, aborted := console_read_confirm("Would you like to receive subscription updates from the users on all servers?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.DisplaySubscriptionUpdates_set(substatus)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.DisplaySubscriptionUpdates_set(substatus)
	}
	return false
}

func (conf *config) Servers_DisplayEvents_prompt() bool {
	display_events, aborted := console_read_confirm("Would you like to display events from all servers, even if they aren't selected as the active server?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.DisplayEvents_set(display_events)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.DisplayEvents_set(display_events)
	}
	return false
}

func (conf *config) Servers_BeepOnCriticalEvents_prompt() bool {
	beep_events, aborted := console_read_confirm(TT_BEEP + "Would you like to hear a beep on critical events for all servers like the one just sent to the console window?\r\n")
	if aborted {
		return true
	}
	for _, server := range conf.Servers_read() {
		server.BeepOnCriticalEvents_set(beep_events)
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.BeepOnCriticalEvents_set(beep_events)
	}
	return false
}

func (conf *config) Servers_LogInfo_prompt() bool {
	log_events, aborted := console_read_confirm("Would you like to log server events for all servers in log files?\r\n")
	if aborted {
		return true
	}
	log_events_account := false
	if log_events {
		log_events_account, aborted = console_read_confirm("Would you like to log each event from a particular user in their own log file matching their username on servers?\r\n")
		if aborted {
			return true
		}
	}
	for _, server := range conf.Servers_read() {
		server.LogEvents_set(log_events)
		server.LogEventsAccount_set(log_events_account)
	}
	answer, aborted := conf.SetDefaultValues_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.LogEvents_set(log_events)
		conf.LogEventsAccount_set(log_events_account)
	}
	return false
}

func (conf *config) Servers_UseGlobalNickName_prompt() bool {
	oldnick := conf.NickName_read()
	if oldnick == "" {
		answer, aborted := console_read_confirm("No global nickname is currently set. Would you like to set one now?")
		if aborted {
			return true
		}
		if !answer {
			console_write("Aborted.")
			return true
		}
		if conf.NickName_prompt() {
			return true
		}
	}
	nickname := conf.NickName_read()
	if nickname == "" {
		console_write("No global nickname set. Updating all servers cannot be achieved. Aborted.")
		return true
	}
	use_global_nick, aborted := console_read_confirm("Would you like to use the global nickname for all servers, currently set to " + nickname + "?\r\n")
	if aborted {
		return true
	}
	console_write("Updating nicknames, please wait.")
	updated := 0
	updatemsg := ""
	for _, server := range conf.Servers_read() {
		switch use_global_nick {
		case true:
			if server.connected() {
				usr := server.User_find_id(server.Uid_read())
				if usr != nil && usr.NickName_read() != nickname {
					if server.cmd_changenick(nickname) {
						updated++
					}
				}
			}
			server.UseGlobalNickName_set(true)
			server.NickName_set("")
		case false:
			if server.connected() {
				usr := server.User_find_id(server.Uid_read())
				if usr != nil {
					server.NickName_set(usr.NickName_read())
					updatemsg += server.DisplayName_read() + " using "
					if usr.NickName_read() == "" {
						updatemsg += "no nickname"
					} else {
						updatemsg += "the nickname " + usr.NickName_read()
					}
					updatemsg += ".\r\n"
					updated++
				}
			}
			server.UseGlobalNickName_set(false)
		}
	}
	if updated != 0 {
		msg := strconv.Itoa(updated) + "server"
		if updated != 1 {
			msg += "s"
		}
		msg += " were connected and updated to use "
		if updatemsg != "" {
			msg += "no global nickname. Settings changed detailed below.\r\n" + updatemsg
		} else {
			msg += "the global nickname " + nickname
		}
		console_write(msg + ".")
	}
	if oldnick != nickname {
		return false
	}
	answer, aborted := conf.SetDefaultValue_prompt()
	if aborted {
		return true
	}
	if answer {
		conf.UseGlobalNickName_set(use_global_nick)
	}
	return false
}

func (conf *config) DisplayTimestamp_prompt() bool {
	answer, aborted := console_read_confirm("Do you wish to display timestamps on displayed server events?\r\n")
	if aborted {
		return true
	}
	conf.DisplayTimestamp_set(answer)
	return false
}

func (conf *config) NickName_prompt() bool {
	nickname, err := console_read_prompt("Enter the global nickname.")
	if err != nil {
		return true
	}
	if nickname == "" {
		console_write("No global nickname set.")
		return false
	}
	answer, aborted := console_read_confirm("Do you wish to set the global nickname to " + nickname + "?\r\n")
	if aborted {
		return true
	}
	if answer {
		conf.NickName_set(nickname)
		if conf.UseGlobalNickName_prompt() {
			return true
		}
	} else {
		console_write("No global nickname set.")
	}
	return false
}

func (conf *config) Server_modify_prompt(server *tt_server, changeprompt bool) bool {
	for {
		if conf.Server_prompt_displayname(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_conn_info(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_account_info(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_nickname(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_autosubscriptions(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_autoconnect_info(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_events_info(server, changeprompt) {
			return true
		}
		if conf.Server_prompt_log_info(server, changeprompt) {
			return true
		}
		res, aborted := console_read_confirm("You have entered the following information for this server.\r\n" + server.Info_str() + "Is this correct?\r\n")
		if aborted {
			return true
		}
		if !res {
			changeprompt = true
			continue
		}
		break
	}
	return false
}

func (conf *config) Server_add_prompt() bool {
	server := NewServer(conf)
	if conf.Server_modify_prompt(server, false) {
		return true
	}
	conf.Server_add(server)
	return false
}

func (conf *config) Server_prompt_displayname(server *tt_server, changeprompt bool) bool {
	oldname := server.DisplayName_read()
	if !changeprompt && oldname != "" {
		servers := conf.Server_find_name(oldname)
		if len(servers) != 0 && servers[0] != server {
			console_write("Error: A server with the name " + oldname + " already exists.")
			oldname = ""
		} else {
			return false
		}
	}
	if changeprompt && oldname != "" {
		answer, aborted := console_read_confirm("The servers display name is currently " + oldname + ". Would you like to change it?\r\n")
		if aborted {
			return true
		}
		if !answer {
			return false
		}
	}
	var name string
	var err error
	for {
		name, err = console_read_prompt("Enter the display name for the server. This will be used when displaying logged events, and if the server is inactive, to set the server to be the active server, or select the server in certain commands.")
		if err != nil {
			return true
		}
		if name == "" {
			console_write("Empty value not accepted.")
			continue
		}
		servers := conf.Server_find_name(name)
		if len(servers) != 0 && servers[0] != server {
			console_write("Error: A server with the name " + name + " already exists.")
			continue
		}
		prompt := ""
		if oldname != "" {
			prompt = "Would you like to change the server's display name from " + oldname + " to " + name
		} else {
			prompt = "Do you wish the server's display name to be " + name
		}
		if name == oldname {
			console_write("Display name unchanged.")
			return false
		}
		answer, aborted := console_read_confirm(prompt + "?\r\n")
		if aborted {
			return true
		}
		if !answer {
			continue
		}
		break
	}
	server.DisplayName_set(name)
	return false
}

func (conf *config) Server_prompt_conn_info(server *tt_server, changeprompt bool) bool {
	host := server.Host_read()
	port := server.Tcpport_read()
	for {
		changed_host, aborted := conf.Server_prompt_host(server, changeprompt)
		if aborted {
			return true
		}
		changed_port, aborted := conf.Server_prompt_port(server, changeprompt)
		if aborted {
			return true
		}
		if changed_host || changed_port {
			if server.connected() {
				console_write("Disconnecting.")
				server.Shutdown()
			}
			console_write("Attempting test connection.")
			ip := ""
			connerr := server.connect_silent()
			if connerr != nil {
				console_write("Warning: failed to connect to server. The server may be temporarily unavailable.\r\nError: " + connerr.Error())
				answer, aborted := console_read_confirm("Do you wish to continue?\r\n")
				if aborted {
					server.Host_set(host)
					server.Tcpport_set(port)
					return true
				}
				if answer {
					break
				}
				changeprompt = true
				server.Host_set(host)
				server.Tcpport_set(port)
				continue
			} else {
				console_write("Test connection successful.")
				ip, _, _ = net.SplitHostPort(server.conn.RemoteAddr().String())
				server.disconnect_silent()
			}
			servers := conf.Server_find_info(ip, port)
			if len(servers) != 0 {
				if len(servers) != 1 || servers[0] != server {
					console_write("A server already exists to connect to the same location provided.")
					changeprompt = true
					server.Host_set(host)
					server.Tcpport_set(port)
					continue
				}
			}
		}
		break
	}
	return false
}

func (conf *config) Server_prompt_host(server *tt_server, changeprompt bool) (bool, bool) {
	//Return values are as follows.
	//First is changed,
	//but return this one as true
	//even if unchanged if dealing with the function non-interactively.
	//Second is aborted.
	oldhost := server.Host_read()
	changed := false
	answer := false
	aborted := false
	for {
		if oldhost != "" {
			if changeprompt {
				answer, aborted = console_read_confirm("The current hostname is " + oldhost + ". Would you like to change it?\r\n")
				if aborted {
					return false, true
				}
			} else {
				changed = true
			}
		} else {
			answer = true
		}
		if answer {
			host, err := console_read_prompt("Enter the server's hostname.")
			if err != nil {
				return false, true
			}
			if host == "" {
				console_write("Empty value not accepted.")
				continue
			}
			if host != oldhost {
				changed = true
				server.Host_set(host)
			} else {
				console_write("The two hostnames are identical. Value unchanged.")
			}
		}
		if changed {
			reserror := server.resolve()
			if reserror != nil {
				console_write("Error: inicial hostname resolution failed. You will be unable to connect to the server.\r\nError: " + reserror.Error())
				answer, aborted = console_read_confirm("Do you wish to continue?")
				if aborted {
					server.Host_set(oldhost)
					return false, true
				}
				if !answer {
					changed = false
					changeprompt = true
					server.Host_set(oldhost)
					continue
				}
			}
		}
		break
	}
	return changed, aborted
}

func (conf *config) Server_prompt_port(server *tt_server, changeprompt bool) (bool, bool) {
	oldport := server.Tcpport_read()
	answer := false
	aborted := false
	changed := false
	//First return value is changed.
	//Return true if non-interactive.
	//Second return value is aborted.
	for {
		if oldport != "" {
			if changeprompt {
				answer, aborted = console_read_confirm("The servers current TCP port is " + oldport + ". Would you like to change it?\r\n")
				if aborted {
					return false, true
				}
			} else {
				changed = true
			}
		} else {
			answer = true
		}
		if answer {
			changeprompt = true
			port, err := console_read_prompt("Enter the servers TCP port.")
			if err != nil {
				server.Tcpport_set(oldport)
				return false, true
			}
			num, err := strconv.Atoi(port)
			if err != nil || num < 1 || num > 65535 {
				console_write("Invalid port number: " + strconv.Itoa(num) + ".")
				changed = false
				changeprompt = true
				continue
			}
			if port != oldport {
				changed = true
				server.Tcpport_set(port)
			} else {
				console_write("The two port numbers are identical. Value unchanged.")
			}
		}
		if !changeprompt {
			num, err := strconv.Atoi(oldport)
			if err != nil || num < 1 || num > 65535 {
				console_write("Invalid port number: " + oldport + ".")
				oldport = ""
				changed = false
				changeprompt = true
				continue
			}
		}
		break
	}
	return changed, aborted
}

func (conf *config) Server_prompt_account_info(server *tt_server, changeprompt bool) bool {
	oldusername := server.AccountName_read()
	oldpassword := server.AccountPassword_read()
	answer := false
	aborted := false
	if oldusername != "" {
		if changeprompt {
			answer, aborted = console_read_confirm("The current username is " + oldusername + ". Would you like to change it?\r\n")
			if aborted {
				return true
			}
		}
	} else {
		answer = true
	}
	if answer {
		username, err := console_read_prompt("Enter the account username. Press enter for no account.")
		if err != nil {
			return true
		}
		if username != oldusername {
			server.AccountName_set(username)
		} else {
			if oldusername != "" {
				console_write("The usernames are identical. Value unchanged.")
			}
		}
	}
	answer = false
	if oldpassword != "" {
		if changeprompt {
			answer, aborted = console_read_confirm("The current password is " + oldpassword + ". Would you like to change it?\r\n")
			if aborted {
				return true
			}
		}
	} else {
		answer = true
	}
	if answer {
		password, err := console_read_prompt("Enter the account password. Press enter for none.")
		if err != nil {
			return true
		}
		if password != oldpassword {
			server.AccountPassword_set(password)
		} else {
			if oldpassword != "" {
				console_write("The passwords are identical. Value unchanged.")
			}
		}
	}
	return false
}

func (conf *config) Server_prompt_nickname(server *tt_server, changeprompt bool) bool {
	answer := false
	aborted := false
	oldnickname := server.NickName_read()
	if oldnickname != "" {
		if changeprompt {
			answer, aborted = console_read_confirm("The current nickname is " + oldnickname + ". Would you like to change it?\r\n")
			if aborted {
				return true
			}
			if !answer {
				return false
			}
		}
	} else {
		if changeprompt {
			answer = true
		}
	}
	if answer {
		changeprompt = true
		nickname, err := console_read_prompt("Enter the nickname that you wish to use for this server. Press enter for no nickname.")
		if err != nil {
			return true
		}
		if nickname != oldnickname {
			server.NickName_set(nickname)
		} else {
			if oldnickname != "" {
				if changeprompt {
					console_write("The nickname's are identical. Value unchanged.")
				}
			}
		}
	}
	globalnick := conf.NickName_read()
	if server.NickName_read() == "" && globalnick != "" {
		if conf.UseGlobalNickName_read() && !changeprompt {
			server.UseGlobalNickName_set(true)
		} else {
			answer, aborted := console_read_confirm("Would you like to connect to the server using the global nickname, currently set to " + globalnick + "?\r\n")
			if aborted {
				return true
			}
			server.UseGlobalNickName_set(answer)
		}
	}
	return false
}

func (conf *config) Server_prompt_autoconnect_info_prompts(server *tt_server) bool {
	autoConnectOnStart, aborted := console_read_confirm("Would you like to automatically connect to the server when the program starts?\r\n")
	if aborted {
		return true
	}
	server.AutoConnectOnStart_set(autoConnectOnStart)
	autoConnectOnDisconnect, aborted := console_read_confirm("Would you like to automatically reconnect to the server when the connection is lost?\r\n")
	if aborted {
		return true
	}
	server.AutoConnectOnDisconnect_set(autoConnectOnDisconnect)
	autoConnectOnKick, aborted := console_read_confirm("Would you like to automatically reconnect to the server when you are kicked?\r\n")
	if aborted {
		return true
	}
	server.AutoConnectOnKick_set(autoConnectOnKick)
	return false
}

func (conf *config) Server_prompt_autoconnect_info(server *tt_server, changeprompt bool) bool {
	if changeprompt {
		if conf.Server_prompt_autoconnect_info_prompts(server) {
			return true
		}
		return false
	}
	if !conf.UseDefaults_read() {
		if conf.Server_prompt_autoconnect_info_prompts(server) {
			return true
		}
		return false
	}
	server.AutoConnectOnStart_set(conf.AutoConnectOnStart_read())
	server.AutoConnectOnDisconnect_set(conf.AutoConnectOnDisconnect_read())
	server.AutoConnectOnKick_set(conf.AutoConnectOnKick_read())
	return false
}

func (conf *config) Server_prompt_autosubscriptions(server *tt_server, changeprompt bool) bool {
	if changeprompt {
		if conf.Server_prompt_autosubscriptions_prompts(server) {
			return true
		}
		return false
	}
	if !conf.UseDefaults_read() {
		if conf.Server_prompt_autosubscriptions_prompts(server) {
			return true
		}
		return false
	}
	server.AutoSubscriptions_set(conf.AutoSubscriptions_read())
	return false
}

func (conf *config) Server_prompt_autosubscriptions_prompts(server *tt_server) bool {
	answer, aborted := console_read_confirm("Would you like to set automatic subscriptions for users that log in to this server?")
	if aborted {
		return true
	}
	if !answer {
		return false
	}
	current_subs := server.AutoSubscriptions_read()
	current_subs_str := server.AutoSubscriptions_read_str()
	msg := "Modifying automatic local subscriptions for " + server.DisplayName_read() + ".\r\n"
	if current_subs_str != "" {
		msg += "Current automatic local subscriptions: " + current_subs_str + "\r\n"
	} else {
		msg += "Current automatic local subscriptions disabled.\r\n"
	}
	console_write(msg)
	new_subs, aborted := teamtalk_flags_subscriptions_menu(current_subs)
	if aborted {
		return true
	}
	if new_subs == current_subs {
		console_write("Automatic local subscriptions unchanged.")
		return false
	}
	server.AutoSubscriptions_set(new_subs)
	sub_str := server.AutoSubscriptions_read_str()
	if sub_str != "" {
		console_write("Automatic local subscriptions changed to " + sub_str)
	} else {
		console_write("Automatic local subscriptions disabled.")
	}
	if !server.connected() {
		return false
	}
	if new_subs != 0 {
		answer, aborted = console_read_confirm("Would you like to update the subscriptions for the users on this server?")
		if aborted {
			return true
		}
		if !answer {
			return false
		}
		console_write("Updating subscriptions as a background task.")
		go func(new_subs int) {
			updated := 0
			for _, usr := range server.Users_sort(0) {
				if usr.Subscriptions_local_read() == new_subs {
					continue
				}
				if server.cmd_changesubscriptions(usr.Uid_read(), new_subs) {
					updated++
				}
			}
			if updated != 0 {
				msg := "Local subscriptions updated for " + strconv.Itoa(updated) + " user"
				if updated != 1 {
					msg += "s"
				}
				console_write(msg + ".")
			} else {
				console_write("Automatic local subscriptions not updated for any users.")
			}
		}(new_subs)
	}
	return false
}

func (conf *config) Server_prompt_events_info_prompts(server *tt_server) bool {
	conn_info, aborted := console_read_confirm("Would you like to display extended connection information from the users, such as their IP address, user id, client name and version, etc?\r\n")
	if aborted {
		return true
	}
	server.DisplayExtendedConnInfo_set(conn_info)
	status, aborted := console_read_confirm("Would you like to receive status updates from the users on the server?\r\n")
	if aborted {
		return true
	}
	server.DisplayStatusUpdates_set(status)
	substatus, aborted := console_read_confirm("Would you like to receive subscription updates from the users on the server?\r\n")
	if aborted {
		return true
	}
	server.DisplaySubscriptionUpdates_set(substatus)
	display_events, aborted := console_read_confirm("Would you like to display the server events if the server is not selected as the active server?\r\n")
	if aborted {
		return true
	}
	server.DisplayEvents_set(display_events)
	beep_events, aborted := console_read_confirm(TT_BEEP + "Would you like to hear a beep on critical events like the one just sent to the console window?\r\n")
	if aborted {
		return true
	}
	server.BeepOnCriticalEvents_set(beep_events)
	return false
}

func (conf *config) Server_prompt_events_info(server *tt_server, changeprompt bool) bool {
	if changeprompt {
		if conf.Server_prompt_events_info_prompts(server) {
			return true
		}
		return false
	}
	if !conf.UseDefaults_read() {
		if conf.Server_prompt_events_info_prompts(server) {
			return true
		}
		return false
	}

	server.DisplayExtendedConnInfo_set(conf.DisplayExtendedConnInfo_read())

	server.DisplayStatusUpdates_set(conf.DisplayStatusUpdates_read())

	server.DisplaySubscriptionUpdates_set(conf.DisplaySubscriptionUpdates_read())

	server.DisplayEvents_set(conf.DisplayEvents_read())

	server.BeepOnCriticalEvents_set(conf.BeepOnCriticalEvents_read())

	return false
}

func (conf *config) Server_prompt_log_info_prompts(server *tt_server) bool {
	log_events, aborted := console_read_confirm("Would you like to log the server events in a log file?\r\n")
	if aborted {
		return true
	}
	server.LogEvents_set(log_events)
	log_events_account := false
	if log_events {
		log_events_account, aborted = console_read_confirm("Would you like to log each event from a particular user in their own log file matching their username on the server?\r\n")
		if aborted {
			return true
		}
	}
	server.LogEventsAccount_set(log_events_account)
	return false
}

func (conf *config) Server_prompt_log_info(server *tt_server, changeprompt bool) bool {
	if changeprompt {
		if conf.Server_prompt_log_info_prompts(server) {
			return true
		}
		return false
	}
	if !conf.UseDefaults_read() {
		if conf.Server_prompt_log_info_prompts(server) {
			return true
		}
		return false
	}
	log_events := conf.LogEvents_read()
	server.LogEvents_set(log_events)
	if log_events {
		server.LogEventsAccount_set(conf.LogEventsAccount_read())
	}
	return false
}

func (conf *config) Server_active_read() *tt_server {
	defer conf.Unlock()
	conf.Lock()
	return conf.server
}

func (conf *config) Server_active_read_name() string {
	defer conf.Unlock()
	conf.Lock()
	return conf.ActiveServer
}

func (conf *config) Server_active_set(s *tt_server) {
	conf.Lock()
	conf.server = s
	if s == nil {
		conf.ActiveServer = ""
	} else {
		conf.ActiveServer = s.DisplayName_read()
	}
	conf.Unlock()
	conf.Write()
}

func (conf *config) Server_active_clear() {
	conf.Server_active_set(nil)
}

func (conf *config) Server_find_name(name string) []*tt_server {
	servers := []*tt_server{}
	defer conf.Unlock()
	conf.Lock()
	if len(conf.Servers) == 0 {
		return servers
	}
	for _, server := range conf.Servers {
		if strings.ToLower(server.DisplayName_read()) == strings.ToLower(name) {
			servers = append(servers, server)
		}
	}
	return servers
}

func (conf *config) Server_find_info(ip, port string) []*tt_server {
	servers := []*tt_server{}
	defer conf.Unlock()
	conf.Lock()
	if len(conf.Servers) == 0 || ip == "" {
		return servers
	}
	for _, server := range conf.Servers {
		if server.Ip_read() == ip && server.Tcpport_read() == port {
			servers = append(servers, server)
		}
	}
	return servers
}

func (conf *config) Duplicate_servers_by_name(servers []*tt_server) bool {
	console_write(strconv.Itoa(len(servers)) + " servers have been found that are named " + servers[0].DisplayName_read() + "\r\nThe duplicates must be renamed or removed, as all server names must be unique. Server name checking is case insensitive.")
	menu := []string{}
	for _, s := range servers {
		menu = append(menu, s.DisplayName_read()+" ("+s.Host_read()+":"+s.Tcpport_read()+")")
	}
menuloop:
	for {
		res, aborted := console_read_menu("Please select a server to rename or remove from the following menu.\r\n", menu)
		if aborted {
			return true
		}
		server := servers[res]
		serverinfo := server.Info_str()
		console_write("This server has the following information:\r\n" + serverinfo)
		answer, aborted := console_read_confirm("Would you like to rename the server?\r\n")
		if aborted {
			return true
		}
		if answer {
		loop:
			for {
				name, err := console_read_prompt("Enter a new name for the server, or abort to cancel.")
				if err != nil {
					return true
				}
				switch strings.ToLower(name) {
				case "":
					console_write("An empty value is not supported.")
					continue loop
				case "abort":
					console_write("Aborted.")
					return true
				default:
					if len(conf.Server_find_name(name)) != 0 {
						console_write("A server with the name " + name + " already exists.")
						continue loop
					}
					res, aborted := console_read_confirm("Do you wish to rename the server to " + name + "?\r\n")
					if aborted {
						return true
					}
					if !res {
						continue menuloop
					}
					server.DisplayName_set(name)
					console_write("Server renamed to " + name)
					conf.Write()
					break loop
				}
			}
			return false
		}
		answer, aborted = console_read_confirm("Do you wish to remove the server?")
		if aborted {
			return true
		}
		if answer {
			conf.Server_remove(server)
			console_write("Server removed.")
		}
	}
	return false
}

func (conf *config) Duplicate_servers_by_info(servers []*tt_server) bool {
	console_write(strconv.Itoa(len(servers)) + " servers have been found that are connecting to the same place.\r\nThe duplicates must be removed, as multiple connections to the same server are not permitted.")
	prompt := "Please select a server to keep. The "
	switch len(servers) {
	case 1:
		prompt += "other servers"
	default:
		prompt += "other server"
	}
	prompt += " will be removed.\r\n"
	menu := []string{}
	for _, s := range servers {
		menu = append(menu, s.DisplayName_read()+" ("+s.Host_read()+":"+s.Tcpport_read()+")")
	}
menuloop:
	for {
		res, aborted := console_read_menu(prompt, menu)
		if aborted {
			return true
		}
		server := servers[res]
		serverinfo := server.Info_str()
		console_write("This server has the following information:\r\n" + serverinfo)
		answer, aborted := console_read_confirm("Is this the server you wish to keep?\r\n")
		if aborted {
			return true
		}
		if !answer {
			continue menuloop
		}
		count := 0
		for _, s := range servers {
			if s == server {
				continue
			}
			conf.Server_remove(s)
			count++
		}
		msg := strconv.Itoa(count) + " server"
		if count > 1 {
			msg += "s"
		}
		console_write(msg + " removed.")
		break menuloop
	}
	return false
}

func (conf *config) AutoConnectOnStart_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.AutoConnectOnStart
}

func (conf *config) AutoConnectOnStart_set(autoConnect bool) {
	if autoConnect == conf.AutoConnectOnStart_read() {
		return
	}
	conf.Lock()
	conf.AutoConnectOnStart = autoConnect
	conf.Unlock()
	conf.Write()
}

func (conf *config) AutoConnectOnDisconnect_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.AutoConnectOnDisconnect
}

func (conf *config) AutoConnectOnDisconnect_set(autoConnect bool) {
	if autoConnect == conf.AutoConnectOnDisconnect_read() {
		return
	}
	conf.Lock()
	conf.AutoConnectOnDisconnect = autoConnect
	conf.Unlock()
	conf.Write()
}

func (conf *config) AutoConnectOnKick_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.AutoConnectOnKick
}

func (conf *config) AutoConnectOnKick_set(autoConnect bool) {
	if autoConnect == conf.AutoConnectOnKick_read() {
		return
	}
	conf.Lock()
	conf.AutoConnectOnKick = autoConnect
	conf.Unlock()
	conf.Write()
}

func (conf *config) AutoSubscriptions_read() int {
	defer conf.Unlock()
	conf.Lock()
	return conf.AutoSubscriptions
}

func (conf *config) AutoSubscriptions_read_str() string {
	return teamtalk_flags_subscriptions_str(conf.AutoSubscriptions_read())
}

func (conf *config) AutoSubscriptions_set(subs int) {
	defer conf.Unlock()
	conf.Lock()
	conf.AutoSubscriptions = subs
}

func (conf *config) DisplayExtendedConnInfo_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.DisplayExtendedConnInfo
}

func (conf *config) DisplayExtendedConnInfo_set(info bool) {
	if info == conf.DisplayExtendedConnInfo_read() {
		return
	}
	conf.Lock()
	conf.DisplayExtendedConnInfo = info
	conf.Unlock()
	conf.Write()
}

func (conf *config) DisplayStatusUpdates_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.DisplayStatusUpdates
}

func (conf *config) DisplayStatusUpdates_set(status bool) {
	if status == conf.DisplayStatusUpdates_read() {
		return
	}
	conf.Lock()
	conf.DisplayStatusUpdates = status
	conf.Unlock()
	conf.Write()
}

func (conf *config) DisplaySubscriptionUpdates_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.DisplaySubscriptionUpdates
}

func (conf *config) DisplaySubscriptionUpdates_set(substatus bool) {
	if substatus == conf.DisplaySubscriptionUpdates_read() {
		return
	}
	conf.Lock()
	conf.DisplaySubscriptionUpdates = substatus
	conf.Unlock()
	conf.Write()
}

func (conf *config) BeepOnCriticalEvents_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.BeepOnCriticalEvents
}

func (conf *config) BeepOnCriticalEvents_set(beepOnCriticalEvents bool) {
	if beepOnCriticalEvents == conf.BeepOnCriticalEvents_read() {
		return
	}
	conf.Lock()
	conf.BeepOnCriticalEvents = beepOnCriticalEvents
	conf.Unlock()
	conf.Write()
}

func (conf *config) DisplayEvents_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.DisplayEvents
}

func (conf *config) DisplayEvents_set(displayEvents bool) {
	if displayEvents == conf.DisplayEvents_read() {
		return
	}
	conf.Lock()
	conf.DisplayEvents = displayEvents
	conf.Unlock()
	conf.Write()
}

func (conf *config) LogEvents_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.LogEvents
}

func (conf *config) LogEvents_set(logEvents bool) {
	if logEvents == conf.LogEvents_read() {
		return
	}
	conf.Lock()
	conf.LogEvents = logEvents
	conf.Unlock()
	conf.Write()
}

func (conf *config) LogEventsAccount_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.LogEventsAccount
}

func (conf *config) LogEventsAccount_set(logEventsAccount bool) {
	if logEventsAccount == conf.LogEventsAccount_read() {
		return
	}
	conf.Lock()
	conf.LogEventsAccount = logEventsAccount
	conf.Unlock()
	conf.Write()
}

func (conf *config) UseGlobalNickName_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.UseGlobalNickName
}

func (conf *config) UseGlobalNickName_set(UseGlobalNickName bool) {
	if UseGlobalNickName == conf.UseGlobalNickName_read() {
		return
	}
	conf.Lock()
	conf.UseGlobalNickName = UseGlobalNickName
	conf.Unlock()
	conf.Write()
}

func (conf *config) UseDefaults_read() bool {
	defer conf.Unlock()
	conf.Lock()
	return conf.UseDefaults
}

func (conf *config) UseDefaults_set(UseDefaults bool) {
	if UseDefaults == conf.UseDefaults_read() {
		return
	}
	conf.Lock()
	conf.UseDefaults = UseDefaults
	conf.Unlock()
	conf.Write()
}
