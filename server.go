package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Command ids.
const TT_CMD_NONE = 0

const (
	TT_CMD_LOGIN         = 1
	TT_CMD_LIST_ACCOUNTS = 2
	TT_CMD_LIST_BANS     = 3
	TT_CMD_MIN_ID        = 4
)

type tt_server struct {
	sync.Mutex
	XMLName                    xml.Name `xml:"server"`
	config                     *config
	DisplayName                string `xml:"name"`
	Host                       string `xml:"host"`
	ip                         string
	Tcpport                    string `xml:"port"`
	address                    string
	conn                       net.Conn
	reader                     *bufio.Reader
	writer                     *bufio.Writer
	keepalive                  *time.Ticker
	keepaliveon                bool
	keepalivedone              chan bool
	usertimeout                int
	cmd_sent                   bool
	cmdfinish                  chan int
	cl                         sync.Mutex
	checkevents                *time.Ticker
	checkeventson              bool
	checkeventsdone            chan bool
	uid                        int
	user_rights                int
	user_type                  int
	logged_in                  bool
	AccountName                string `xml:"username,omitempty"`
	AccountPassword            string `xml:"password,omitempty"`
	NickName                   string `xml:"nickname,omitempty"`
	UseGlobalNickName          bool   `xml:"useGlobalNickName,omitempty"`
	AutoConnectOnStart         bool   `xml:"autoConnectOnStart"`
	AutoConnectOnDisconnect    bool   `xml:"autoConnectOnDisconnect"`
	AutoConnectOnKick          bool   `xml:"autoConnectOnKick"`
	kicked                     bool
	AutoSubscriptions          int `xml:"automatic>subscriptions,omitempty"`
	AutoMoveFrom               int `xml:"automatic>moveFrom,omitempty"`
	autoMoveFrom               int
	AutoMoveTo                 int `xml:"automatic>moveTo,omitempty"`
	autoMoveTo                 int
	DisplayExtendedConnInfo    bool `xml:"displayExtendedConnInfo"`
	DisplayStatusUpdates       bool `xml:"displayStatusUpdates"`
	DisplaySubscriptionUpdates bool `xml:"displaySubscriptionUpdates"`
	DisplayEvents              bool `xml:"displayServerEventsIfInactive"`
	BeepOnCriticalEvents       bool `xml:"beepOnCriticalServerEvents"`
	LogEvents                  bool `xml:"logServerEvents"`
	LogEventsAccount           bool `xml:"logServerEventsPerUserAccount"`
	name                       string
	protocol                   string
	motd                       string
	version                    string
	maxusers                   int
	cmderror                   error
	cmdid                      int
	cmdid_add                  int
	shutdown                   bool
	users                      map[int]*tt_user
	channels                   map[int]*tt_channel
	accounts                   map[string]map[string]string
	accounts_cached            map[string]map[string]string
	bans                       map[string]map[string]string
	bans_cached                map[string]map[string]string
	log_username               string
	log_buffer                 string
	log_history                []string
	log_history_number         int
	log_timestamp              string
	log_timestamp_console      string
	log_timestamp_account      map[string]string
	Debug                      bool `xml:"debug,omitempty"`
}

func NewServer(conf *config) *tt_server {
	return &tt_server{
		config: conf,
	}
}

func (server *tt_server) resolve() error {
	defer server.Unlock()
	server.Lock()
	if server.ip == "" {
		ip6, err := net.ResolveIPAddr("ip6", server.Host)
		if err != nil {
			ip4, err := net.ResolveIPAddr("ip4", server.Host)
			if err != nil {
				return err
			} else {
				server.ip = ip4.String()
			}
		} else {
			server.ip = ip6.String()
		}
	}
	return nil
}

func (server *tt_server) Startup(autostart bool) {
	if server.connected() || !autostart {
		return
	}
	err := server.connect()
	if err != nil {
		return
	}
	server.Lock()
	ip, _, _ := net.SplitHostPort(server.conn.RemoteAddr().String())
	tmpip := server.ip
	server.ip = ip
	server.Unlock()
	if tmpip != "" && ip != tmpip {
		server.Log_write("Warning: Connected to "+ip+", but resolved to "+tmpip+".", true)
	}
	server.Config().wg.Add(1)
	go server.Process()
}

func (server *tt_server) connect_silent() error {
	server.Lock()
	address := server.address
	server.Unlock()
	if address == "" {
		server.resolve()
		server.Lock()
		address = server.Host + ":" + server.Tcpport
		server.Unlock()
	}
	timeout := time.Duration(5 * time.Second)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	server.init_vars()
	server.Lock()
	server.conn = conn
	if connaddress := conn.RemoteAddr().String(); server.address != connaddress {
		server.address = connaddress
	}
	server.writer = bufio.NewWriter(server.conn)
	server.reader = bufio.NewReader(server.conn)
	server.usertimeout = -1
	server.keepalivedone = make(chan bool)
	server.cmdfinish = make(chan int)
	server.checkeventsdone = make(chan bool)
	server.cmdid_add = TT_CMD_MIN_ID
	server.shutdown = false
	server.kicked = false
	server.Unlock()
	return nil
}

func (server *tt_server) init_vars() {
	server.Log_reset()
	server.Logged_in_set(false)
	defer server.Unlock()
	server.Lock()
	server.users = make(map[int]*tt_user, 0)
	server.channels = make(map[int]*tt_channel, 0)
	server.log_timestamp_account = make(map[string]string, 0)
	server.log_timestamp = ""
	server.uid = 0
}

func (server *tt_server) connect() error {
	err := server.connect_silent()
	if err != nil {
		server.Log_debug("Error connecting to " + server.Host_read() + ":" + server.Tcpport_read() + ": " + err.Error() + ".\r\nConnection failure.")
		return err
	}
	server.Log_write("Connected.", true)
	server.Log_debug("Debug mode enabled.")
	return nil
}

func (server *tt_server) autoconnect() {
	for {
		if server.Shutdown_read() {
			return
		}
		err := server.connect()
		if err == nil {
			return
		}
	}
}

func (server *tt_server) Read_line() (string, error) {
	line, err := server.reader.ReadString('\n')
	if err != nil {
		if server.Shutdown_read() {
			server.Log_debug("Error receiving data:\r\n" + err.Error())
		}
	} else {
		server.Log_debug("Received data:\r\n" + line)
	}
	return line, err
}

func (server *tt_server) Process() {
	defer server.Config().wg.Done()
	defer server.Shutdown()
	// Recover from a panic, which will shut down
	// the server on which it panicked,
	// and will log the panic.
	// This shouldn't happen, so if it does,
	// something is seriously wrong.
	// This is here to gracefully recover,
	// and avoid crashing the entire program.
	// If recovering from a panic,
	// you will be disconnected from
	// only the server on which the panic occurred.
	defer func() {
		if pd := recover(); pd != nil {
			server.Log_write("PANIC ERROR\r\n"+fmt.Sprint(pd), true)
		}
	}()
loop:
	for {
		cmdline, err := server.Read_line()
		if err != nil {
			if server.Shutdown_read() {
				break loop
			}
			server.disconnect()
			if server.Kicked_read() {
				if server.AutoConnectOnKick_read() {
					server.autoconnect()
					continue loop
				}
				break loop
			}
			if server.AutoConnectOnDisconnect_read() {
				server.autoconnect()
				continue loop
			}
			break loop
		}
		cmdline = strings.TrimSpace(cmdline)
		if cmdline == "" {
			continue
		}
		// Command processing.
		cmd := teamtalk_get_cmd(cmdline)
		params := teamtalk_get_params(cmdline)
		server.Log_reset()
		switch cmd {
		case "teamtalk":
			name := teamtalk_param_str(params, "servername")
			server.Name_set(name)
			maxusers, _ := teamtalk_param_int(params, "maxusers")
			server.MaxUsers_set(maxusers)
			protocol := teamtalk_param_str(params, "protocol")
			server.Protocol_set(protocol)
			uid, _ := teamtalk_param_int(params, "userid")
			server.Uid_set(uid)
			secs, _ := teamtalk_param_int(params, "usertimeout")
			if secs != server.UserTimeout_read() {
				if secs < 10 {
					server.Log_write("User timeout may be too low. Current value in seconds: "+strconv.Itoa(secs)+".", true)
				}
				go server.KeepAlive(secs)
			}
			go server.Login()
		case "accepted":
			server.Logged_in_set(true)
			msg := "Logged in.\r\n"
			user_rights, _ := teamtalk_param_int(params, "userrights")
			server.User_rights_set(user_rights)
			user_type, _ := teamtalk_param_int(params, "usertype")
			server.User_type_set(user_type)

			// Give the bot its own user based on the available information.

			uid, _ := teamtalk_param_int(params, "userid")
			usr := server.User_add(uid)
			usr.Conntime_set()
			nickname := teamtalk_param_str(params, "nickname")
			usr.NickName_set(nickname)
			username := teamtalk_param_str(params, "username")
			usr.UserName_set(username)
			usr.UserType_set(user_type)
			statusmode, _ := teamtalk_param_int(params, "statusmode")
			usr.StatusMode_set(statusmode)
			statusmsg := teamtalk_param_str(params, "statusmsg")
			usr.StatusMsg_set(statusmsg)
			ip := teamtalk_param_str(params, "ipaddr")
			usr.Ip_set(ip)
			usr.Version_set(Version)
			usr.ClientName_set(bot_name)

			// Warn of a few things.
			if !server.User_rights_check(TT_USERRIGHT_MULTI_LOGIN) {
				msg += "Warning: Unable to log in multiple times. You must log out of this user account before you can log in with a TeamTalk client.\r\n"
			}
			if !server.User_rights_check(TT_USERRIGHT_VIEW_ALL_USERS) {
				msg += "Warning: you cannot view any users unless you have joined a channel, and you will see only those users in the channel you have joined. Insufficient information about user login and logouts will be sent to the bot, which may cause problems and errors.\r\n"
			}
			if server.AutoMove_enabled() {
				msg += "Automatic moving of users enabled.\r\n"
			}
			server.Log_write(msg, true)
			go server.CheckEvents(10)
		case "serverupdate":
			msg := ""
			version := teamtalk_param_str(params, "version")
			if version != server.Version_read() {
				server.Version_set(version)
				msg += "Server version: " + version + "\r\n"
			}
			secs, _ := teamtalk_param_int(params, "usertimeout")
			if secs != server.UserTimeout_read() {
				if secs < 10 {
					server.Log_write("User timeout may be too low. Current value in seconds: "+strconv.Itoa(secs)+".", true)
				} else {
					msg += "User timeout in seconds: " + strconv.Itoa(secs) + ".\r\n"
				}
				go server.KeepAlive(secs)
			}
			motd := teamtalk_param_str(params, "motd")
			if motd != server.Motd_read() {
				server.Motd_set(motd)
				msg += "Message of the day updated:\r\n" + motd
			}
			if server.Cmdid_read() != TT_CMD_LOGIN {
				server.Log(msg)
			}
		// Do more here.
		case "addchannel":
			cid, _ := teamtalk_param_int(params, "chanid")
			pid, _ := teamtalk_param_int(params, "parentid")
			ch := server.Channel_add(cid, pid)
			if ch == nil {
				server.Log_write("Error adding channel "+strconv.Itoa(cid)+". Channel already exists.", true)
				continue loop
			}
			cname := teamtalk_param_str(params, "name")
			ch.Name_set(cname)
			cpassword := teamtalk_param_str(params, "password")
			ch.Password_set(cpassword)
			coppassword := teamtalk_param_str(params, "oppassword")
			ch.Oppassword_set(coppassword)
			cprotected, _ := teamtalk_param_int(params, "protected")
			ch.Protected_set(cprotected)
			ctopic := teamtalk_param_str(params, "topic")
			ch.Topic_set(ctopic)
			coperators := teamtalk_param_list(params, "operators")
			ch.Operators_set(coperators)
			cquota, _ := teamtalk_param_int(params, "diskquota")
			ch.Quota_set(cquota)
			cmaxusers, _ := teamtalk_param_int(params, "maxusers")
			ch.Maxusers_set(cmaxusers)
			coptions, _ := teamtalk_param_int(params, "type")
			ch.Options_set(coptions)
			if server.Cmdid_read() != TT_CMD_LOGIN {
				server.Log("Channel added.")
				server.Log("Name: " + cname)
				server.Log("ID: " + strconv.Itoa(cid))
				server.Log("Parent ID: " + strconv.Itoa(pid))
				if ctopic != "" {
					server.Log("Topic: " + ctopic)
				}
				if cprotected != 0 {
					server.Log("Password protected: yes")
				} else {
					server.Log("Password protected: no")
				}
				if cpassword != "" {
					server.Log("Password: " + cpassword)
				}
				if coppassword != "" {
					server.Log("Operator password: " + coppassword)
				}
				coperators_str := ch.Operators_read_str()
				if coperators_str != "" {
					server.Log("Operators: " + coperators_str)
				}
				coptions_str := ch.Options_read_str()
				if coptions_str != "" {
					server.Log("Options: " + coptions_str)
				}
				cquota_str := ch.Quota_read_str()
				if cquota_str != "" {
					server.Log("Disk quota: " + cquota_str)
				}
				server.Log("Maximum users: " + strconv.Itoa(cmaxusers))
			}
		case "removechannel":
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to remove channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			server.Channel_remove(cid)
			server.Log("Channel removed.\r\nChannel path: " + ch.Path_read())
		case "addfile":
			fname := teamtalk_param_str(params, "filename")
			fsize, _ := teamtalk_param_int(params, "filesize")
			fowner := teamtalk_param_str(params, "owner")
			fid, _ := teamtalk_param_int(params, "fileid")
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to add file to channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			if ch.File_add(fid, fname, fsize, fowner) {
				if server.Cmdid_read() != TT_CMD_LOGIN {
					server.Log_username_set(fowner)
					server.Log("File added to " + ch.Path_read() + ".\r\nFilename: " + fname + "\r\nFile owner: " + fowner)
				}
			}
		case "removefile":
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to remove file from channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			fname := teamtalk_param_str(params, "filename")
			if ch.File_remove(fname) {
				server.Log("File removed from " + ch.Path_read() + ".\r\nFilename: " + fname)
			}
		case "loggedin":
			uid, _ := teamtalk_param_int(params, "userid")
			usr := server.User_add(uid)
			if usr == nil {
				usr = server.User_find_id(uid)
			}
			if server.Cmdid_read() != TT_CMD_LOGIN {
				usr.Conntime_set()
			}
			nickname := teamtalk_param_str(params, "nickname")
			usr.NickName_set(nickname)
			username := teamtalk_param_str(params, "username")
			usr.UserName_set(username)
			subscriptions_remote, _ := teamtalk_param_int(params, "subpeer")
			usr.Subscriptions_remote_set(subscriptions_remote)
			subscriptions_local, _ := teamtalk_param_int(params, "sublocal")
			usr.Subscriptions_local_set(subscriptions_local)
			statusmode, _ := teamtalk_param_int(params, "statusmode")
			usr.StatusMode_set(statusmode)
			statusmsg := teamtalk_param_str(params, "statusmsg")
			statusmodestr := usr.StatusMode_read_str()
			usr.StatusMsg_set(statusmsg)
			ip := teamtalk_param_str(params, "ipaddr")
			usr.Ip_set(ip)
			version := teamtalk_param_str(params, "version")
			usr.Version_set(version)
			clientname := teamtalk_param_str(params, "clientname")
			usr.ClientName_set(clientname)
			usertype, _ := teamtalk_param_int(params, "usertype")
			usr.UserType_set(usertype)
			server.Log_username_set(username)
			conn_msg := usr.NickName_log() + " "
			if server.Cmdid_read() != TT_CMD_LOGIN {
				conn_msg += "has"
			} else {
				conn_msg += "is"
			}
			conn_msg += " connected.\r\n"
			conn_extended := "User ID: " + strconv.Itoa(uid) + "\r\n"
			if ip != "" {
				conn_extended += "IP: " + ip + "\r\n"
			}
			if version != "" {
				conn_extended += "Client version: " + version + "\r\n"
			}
			if clientname != "" {
				conn_extended += "Client name: " + clientname + "\r\n"
			}
			if username != "" {
				conn_extended += "Username: " + username + "\r\n"
			}
			conn_extended += "User type: " + usr.UserType_read_str() + "\r\n"
			sub_local_str := usr.Subscriptions_local_read_str()
			sub_remote_str := usr.Subscriptions_remote_read_str()
			sub_msg := ""
			if sub_local_str != "" && sub_remote_str != "" {
				sub_local_msg := "Current local subscriptions: " + sub_local_str + "\r\n"
				sub_remote_msg := "Current remote subscriptions: " + sub_remote_str + "\r\n"
				if sub_local_str != sub_remote_str {
					sub_msg = sub_local_msg + sub_remote_msg
				} else {
					sub_msg = "Current local and remote subscriptions: " + sub_local_str + "\r\n"
				}
			} else {
				if sub_local_str != "" {
					sub_msg = "Current local subscriptions: " + sub_local_str + "\r\n"
				}
				if sub_remote_str != "" {
					sub_msg = "Current remote subscriptions: " + sub_remote_str + "\r\n"
				}
			}
			status_msg := ""
			if statusmodestr != "" {
				status_msg += "Current status mode: " + statusmodestr + "\r\n"
			}
			if statusmsg != "" {
				status_msg += "Current status message: " + statusmsg + "\r\n"
			}
			server.Log_write_files(conn_msg + conn_extended + sub_msg + status_msg)
			console_msg := conn_msg
			if server.DisplayExtendedConnInfo_read() {
				console_msg += conn_extended
			}
			if server.DisplaySubscriptionUpdates_read() {
				console_msg += sub_msg
			}
			if server.DisplayStatusUpdates_read() {
				console_msg += status_msg
			}
			server.Log_console(console_msg, false)
			go server.autosubscribe(usr)
			if server.Cmdid_read() != TT_CMD_LOGIN {
				go server.automove(usr)
			}
		case "updateuser":
			uid, _ := teamtalk_param_int(params, "userid")
			usr := server.User_find_id(uid)
			if usr == nil {
				// server.Log_write("Error: failed to update user information. User ID " + strconv.Itoa(uid) + " doesn't exist.", true)
				continue loop
			}
			status_msg := ""
			sub_local_msg := ""
			sub_remote_msg := ""
			nick_msg := ""
			username := usr.UserName_read()
			server.Log_username_set(username)
			lnickname := usr.NickName_log()
			nickname := teamtalk_param_str(params, "nickname")
			if nickname != usr.NickName_read() {
				nick_msg += lnickname + " changed nickname"
				usr.NickName_set(nickname)
				lnickname = usr.NickName_log()
				if lnickname != nickname {
					nick_msg += ".\r\nThe nickname to identify this user for logging will be " + lnickname + "\r\n"
				} else {
					nick_msg += " to " + nickname + ".\r\n"
				}
			}

			subscriptions_local, _ := teamtalk_param_int(params, "sublocal")
			if old_subscriptions_local := usr.Subscriptions_local_read(); old_subscriptions_local != subscriptions_local {
				usr.Subscriptions_local_set(subscriptions_local)
				sub_local_msg += lnickname + ": local subscription change.\r\n"

				subs_local_added_str := usr.Subscriptions_local_added_str(old_subscriptions_local)

				subs_local_removed_str := usr.Subscriptions_local_removed_str(old_subscriptions_local)

				if subs_local_added_str != "" {
					sub_local_msg += "Subscriptions added: " + subs_local_added_str + "\r\n"
				}
				if subs_local_removed_str != "" {
					sub_local_msg += "Subscriptions removed: " + subs_local_removed_str + "\r\n"
				}
			}

			subscriptions_remote, _ := teamtalk_param_int(params, "subpeer")
			if old_subscriptions_remote := usr.Subscriptions_remote_read(); old_subscriptions_remote != subscriptions_remote {
				usr.Subscriptions_remote_set(subscriptions_remote)
				sub_remote_msg += lnickname + ": remote subscription change." + "\r\n"

				subs_remote_added_str := usr.Subscriptions_remote_added_str(old_subscriptions_remote)

				subs_remote_removed_str := usr.Subscriptions_remote_removed_str(old_subscriptions_remote)

				if subs_remote_added_str != "" {
					sub_remote_msg += "Subscriptions added: " + subs_remote_added_str + "\r\n"
				}
				if subs_remote_removed_str != "" {
					sub_remote_msg += "Subscriptions removed: " + subs_remote_removed_str + "\r\n"
				}
			}

			statusmode, _ := teamtalk_param_int(params, "statusmode")
			statusmsg := teamtalk_param_str(params, "statusmsg")
			if usr.StatusMode_read() != statusmode || usr.StatusMsg_read() != statusmsg {
				oldstatusmodestr := usr.StatusMode_read_str()
				oldstatusmsg := usr.StatusMsg_read()
				usr.StatusMode_set(statusmode)
				statusmodestr := usr.StatusMode_read_str()
				if statusmsg != oldstatusmsg {
					usr.StatusMsg_set(statusmsg)
				}
				if oldstatusmodestr != statusmodestr {
					status_msg += "Mode: " + statusmodestr + "\r\n"
				}
				if oldstatusmsg != statusmsg {
					if statusmsg != "" {
						status_msg += "Message: " + statusmsg + "\r\n"
					} else {
						status_msg += "No status message provided.\r\n"
					}
				}
				if status_msg != "" {
					status_msg = lnickname + ": status change.\r\n" + status_msg
				}
			}
			server.Log_write_files(nick_msg + status_msg + sub_local_msg + sub_remote_msg)
			console_msg := ""
			if nick_msg != "" {
				console_msg += nick_msg
			}
			if server.DisplaySubscriptionUpdates_read() {
				if sub_local_msg != "" {
					console_msg += sub_local_msg
				}
				if sub_remote_msg != "" {
					console_msg += sub_remote_msg
				}
			}
			if server.DisplayStatusUpdates_read() && status_msg != "" {
				console_msg += status_msg
			}
			if console_msg != "" {
				server.Log_console(console_msg, false)
			}
		case "adduser":
			uid, _ := teamtalk_param_int(params, "userid")
			usr := server.User_find_id(uid)
			if usr == nil {
				usr = server.User_add(uid)
				nickname := teamtalk_param_str(params, "nickname")
				usr.NickName_set(nickname)
				username := teamtalk_param_str(params, "username")
				usr.UserName_set(username)
				subscriptions_remote, _ := teamtalk_param_int(params, "subpeer")
				usr.Subscriptions_remote_set(subscriptions_remote)
				subscriptions_local, _ := teamtalk_param_int(params, "sublocal")
				usr.Subscriptions_local_set(subscriptions_local)
				statusmode, _ := teamtalk_param_int(params, "statusmode")
				usr.StatusMode_set(statusmode)
				statusmsg := teamtalk_param_str(params, "statusmsg")
				usr.StatusMsg_set(statusmsg)
				ip := teamtalk_param_str(params, "ipaddr")
				usr.Ip_set(ip)
				version := teamtalk_param_str(params, "version")
				usr.Version_set(version)
				clientname := teamtalk_param_str(params, "clientname")
				usr.ClientName_set(clientname)
				usertype, _ := teamtalk_param_int(params, "usertype")
				usr.UserType_set(usertype)
			}
			username := usr.UserName_read()
			server.Log_username_set(username)
			lnickname := usr.NickName_log()
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to add "+lnickname+" to channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			chanpath := ch.Path_read()
			res := ch.User_add(usr)
			if res {
				msghead := lnickname + " "
				if server.Cmdid_read() != TT_CMD_LOGIN {
					msghead += "has joined"
				} else {
					msghead += "is in"
				}
				server.Log(msghead + " " + chanpath)
			}
			go server.automove(usr)
		case "joined":
			server.Kicked_set(false)
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to join channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			chanpath := ch.Path_read()
			usr := server.User_find_id(server.Uid_read())
			if usr == nil {
				// server.Log_write("Error: unable to find the bot user for channel joining, id " + strconv.Itoa(server.Uid_read()) + ". User doesn't exist.", true)
				continue loop
			}
			res := ch.User_add(usr)
			if res {
				server.Log_console("Entered "+chanpath, false)
				lnickname := usr.NickName_log()
				server.Log_username_set(usr.UserName_read())
				server.Log_write_files(lnickname + " has joined " + chanpath)
			}
		case "left":
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to leave channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			chanpath := ch.Path_read()
			usr := server.User_find_id(server.Uid_read())
			if usr == nil {
				// server.Log_write("Error: unable to find the bot user for channel leaving, id " + strconv.Itoa(server.Uid_read()) + ". User doesn't exist.", true)
				continue loop
			}
			res := ch.User_remove(usr)
			if res {
				server.Log_console("Left "+chanpath, false)
				lnickname := usr.NickName_log()
				server.Log_username_set(usr.UserName_read())
				server.Log_write_files(lnickname + " has left " + chanpath)
			}
		case "removeuser":
			uid, _ := teamtalk_param_int(params, "userid")
			usr := server.User_find_id(uid)
			if usr == nil {
				// server.Log_write("Error: failed to remove user from channel. User ID " + strconv.Itoa(uid) + " doesn't exist.", true)
				continue loop
			}
			username := usr.UserName_read()
			server.Log_username_set(username)
			lnickname := usr.NickName_log()
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to remove "+lnickname+" from channel, "+strconv.Itoa(cid)+". The channel doesn't exist.", true)
				continue loop
			}
			chanpath := ch.Path_read()
			res := ch.User_remove(usr)
			if res {
				server.Log(lnickname + " has left " + chanpath)
			}
		case "messagedeliver":
			// Completely rewrite this to properly log all messages.
			msg_type, _ := teamtalk_param_int(params, "type")
			msg_content := teamtalk_param_str(params, "content")
			uid_src, _ := teamtalk_param_int(params, "srcuserid")
			usr_src := server.User_find_id(uid_src)
			uid_dest, _ := teamtalk_param_int(params, "destuserid")
			usr_dest := server.User_find_id(uid_dest)
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			server.Message_info(msg_type, usr_src, usr_dest, ch, msg_content)
		case "updatechannel":
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Error: failed to update channel "+strconv.Itoa(cid)+". Channel doesn't exist.", true)
				continue loop
			}
			msg := ""
			cname_old := ch.Name_read()
			cpath_old := ch.Path_read()
			cname := teamtalk_param_str(params, "name")
			if cname != cname_old {
				ch.Name_set(cname)
				msg += "New name: " + cname + "\r\nPath: " + ch.Path_read()
			}
			coptions_old := ch.Options_read()
			coptions, _ := teamtalk_param_int(params, "type")
			if coptions_old != coptions {
				ch.Options_set(coptions)
				msg += "New options: " + ch.Options_read_str() + "\r\n"
			}
			cprotected_old := ch.Protected_read()
			cprotected, _ := teamtalk_param_int(params, "protected")
			if cprotected != cprotected_old {
				ch.Protected_set(cprotected)
				if cprotected_old != 0 && cprotected == 0 {
					msg += "Channel no longer password protected."
				} else {
					msg += "Channel password protected."
				}
				msg += "\r\n"
			}
			cpassword_old := ch.Password_read()
			cpassword := teamtalk_param_str(params, "password")
			if cpassword != cpassword_old {
				ch.Password_set(cpassword)
				msg += "New password: " + cpassword + "\r\n"
			}
			coppassword_old := ch.Oppassword_read()
			coppassword := teamtalk_param_str(params, "oppassword")
			if coppassword != coppassword_old {
				ch.Oppassword_set(coppassword)
				msg += "New operator password: " + coppassword + "\r\n"
			}
			ctopic_old := ch.Topic_read()
			ctopic := teamtalk_param_str(params, "topic")
			if ctopic_old != ctopic {
				ch.Topic_set(ctopic)
				msg += "New topic: " + ctopic + "\r\n"
			}
			coperators_old := ch.Operators_read()
			coperators := teamtalk_param_list(params, "operators")
			ch.Operators_set(coperators)
			if len(coperators_old) != len(coperators) {
				msg += "New operators: " + ch.Operators_read_str() + "\r\n"
			}
			cmaxusers_old := ch.Maxusers_read()
			cmaxusers, _ := teamtalk_param_int(params, "maxusers")
			if cmaxusers_old != cmaxusers {
				ch.Maxusers_set(cmaxusers)
				msg += "New maximum users: " + strconv.Itoa(cmaxusers) + "\r\n"
			}
			cquota_old := ch.Quota_read()
			cquota, _ := teamtalk_param_int(params, "diskquota")
			if cquota_old != cquota {
				ch.Quota_set(cquota)
				msg += "New disk quota: " + ch.Quota_read_str() + "\r\n"
			}
			if msg != "" {
				server.Log("Channel " + cpath_old + " updated.\r\n" + msg)
			}
		case "error":
			msg := teamtalk_param_str(params, "message")
			param := teamtalk_param_str(params, "param")
			if param != "" {
				msg += " Missing parameter: " + param
			}
			if msg != "" {
				server.Cmderror_set_str(msg)
			}
		case "begin":
			id, _ := teamtalk_param_int(params, "id")
			if id != 0 {
				server.Cmdid_set(id)
			}
		case "ok":
			server.Cmderror_clear()
		case "end":
			id, _ := teamtalk_param_int(params, "id")
			server.Cmdid_set(0)
			if server.Cmd_sent_read() {
				server.cmdfinish <- id
			}
			if id == TT_CMD_LOGIN && server.Cmderror_read() == nil {
				// Login command has finished successfully.
				server.Login_info()
			}
			break
		case "pong":
			continue loop
		case "kicked":
			uid, _ := teamtalk_param_int(params, "kickerid")
			usr := server.User_find_id(uid)
			if usr == nil {
				// server.Log_write("Error: failed to identify the user who kicked this client. User ID " + strconv.Itoa(uid) + " doesn't exist.", true)
				continue loop
			}
			username := usr.UserName_read()
			server.Log_username_set(username)
			lnickname := usr.NickName_log()
			cid, _ := teamtalk_param_int(params, "chanid")
			ch := server.Channel_find_id(cid)
			if ch == nil {
				server.Log_write("Kicked from server by "+lnickname+".", true)
				server.Kicked_set(true)
			} else {
				cpath := ch.Path_read()
				server.Log_write("Kicked from "+cpath+" by "+lnickname+".", false)
			}
		case "loggedout":
			if len(params) == 0 {
				// This client has been kicked, or otherwise logged out.
				server.Log_write("Logged out.", true)
				server.init_vars()
				server.disconnect()
				if server.Kicked_read() && server.AutoConnectOnKick_read() {
					server.connect()
				}
				continue loop
			}
			uid, _ := teamtalk_param_int(params, "userid")
			usr := server.User_find_id(uid)
			if usr == nil {
				// server.Log_write("Error: failed to log out user. User ID " + strconv.Itoa(uid) + " doesn't exist.", true)
				continue loop
			}
			username := usr.UserName_read()
			server.Log_username_set(username)
			lnickname := usr.NickName_log()
			disconmsg := lnickname + " has disconnected"
			if usr.Conntime_isSet() {
				disconmsg += ", and was connected for " + usr.Conntime_read_str()
			} else {
				disconmsg += ". Connection time unknown"
			}
			server.Log(disconmsg + ".")
			server.User_remove(uid)
		case "useraccount":
			username := teamtalk_param_str(params, "username")
			password := teamtalk_param_str(params, "password")
			usertype, _ := teamtalk_param_int(params, "usertype")
			userrights, _ := teamtalk_param_int(params, "userrights")
			server.Lock()
			if server.accounts_cached == nil {
				server.accounts_cached = make(map[string]map[string]string)
			}
			if _, exists := server.accounts_cached[username]; !exists {
				server.accounts_cached[username] = make(map[string]string)
			}
			server.accounts_cached[username]["password"] = password
			server.accounts_cached[username]["usertype"] = teamtalk_flags_usertype_str(usertype)
			server.accounts_cached[username]["rights"] = teamtalk_flags_userrights_str(userrights)
			server.Unlock()
		case "userbanned":
			ip := teamtalk_param_str(params, "ipaddr")
			delete(params, "ipaddr")
			server.Lock()
			if server.bans_cached == nil {
				server.bans_cached = make(map[string]map[string]string)
			}
			if _, exists := server.bans_cached[ip]; !exists {
				server.bans_cached[ip] = make(map[string]string)
			}
			server.bans_cached[ip] = params
			server.Unlock()
		default:
			server.Log_write("Error: unrecognized command received.\r\nCommand:\r\n"+cmdline, true)
		}
		server.Log_send()
	}
}

func (server *tt_server) Account_add_prompt(username, password, usertype string) bool {
	aborted := false
	username, aborted = server.Account_username_prompt(username)
	if aborted {
		return true
	}
	password, aborted = server.Account_password_prompt(username, password)
	if aborted {
		return true
	}
	utype := 0
	utype, aborted = server.Account_usertype_prompt(username, password, usertype)
	if aborted {
		return true
	}
	urights := 0
	if utype != TT_USERTYPE_ADMIN {
		urights, aborted = server.Account_userrights_prompt(0)
		if aborted {
			return true
		}
	}
	res := server.cmd_new_account(username, password, utype, urights)
	if !res {
		console_write("Failed to add account.")
		return true
	}
	console_write("Successfully added account.")
	return false
}

func (server *tt_server) Account_username_prompt(username string) (string, bool) {
	var err error
	if username == "" {
		for {
			username, err = console_read_prompt("Enter the account username.")
			if err != nil {
				return "", true
			}
			if username == "" {
				answer, aborted := console_read_confirm("You are adding an anonymous account with no username. Are you sure this is what you want to do?\r\n")
				if aborted {
					return "", true
				}
				if !answer {
					console_write("Aborted.")
					return "", true
				}
				break
			} else {
				answer, aborted := console_read_confirm("Would you like the account username to be " + username + "?\r\n")
				if aborted {
					return "", true
				}
				if !answer {
					continue
				}
				break
			}
		}
	}
	return username, false
}

func (server *tt_server) Account_password_prompt(username, password string) (string, bool) {
	var err error
	if password == "" {
		for {
			password, err = console_read_prompt("Enter the account password.")
			if err != nil {
				return "", true
			}
			if password == "" {
				if username == "" {
					answer, aborted := console_read_confirm("You are adding a user account with no username and no password, which will allow anyone to connect to the TeamTalk server. Are you sure this is what you want to do?\r\n")
					if aborted {
						return "", true
					}
					if !answer {
						continue
					}
					break
				}
				answer, aborted := console_read_confirm("You are adding an account named " + username + " with no password. Anyone that knows or can guess the username will be able to log in to the server. Are you sure this is what you want to do?\r\n")
				if aborted {
					return "", true
				}
				if !answer {
					continue
				}
				break
			} else {
				if username == "" {
					answer, aborted := console_read_confirm("You are giving the anonymous account the password " + password + ". Is this correct?\r\n")
					if aborted {
						return "", true
					}
					if !answer {
						continue
					}
					break
				}
				if username == password {
					answer, aborted := console_read_confirm("The username and password of this account are both " + username + ". Are you sure this is what you want to do?\r\n")
					if aborted {
						return "", true
					}
					if !answer {
						continue
					}
					break
				}
				answer, aborted := console_read_confirm("You have entered the account password " + password + ". Is this correct?\r\n")
				if aborted {
					return "", true
				}
				if !answer {
					continue
				}
				break
			}
		}
	}
	return password, false
}

func (server *tt_server) Account_usertype_prompt(username, password, usertype string) (int, bool) {
	aborted := false
	answer := false
	utype := 0
	if usertype == "" {
		for {
			utype, aborted = teamtalk_flags_usertype_menu()
			if aborted {
				return 0, true
			}
			answer, aborted = console_read_confirm("You have selected the user type " + teamtalk_flags_usertype_str(utype) + ". Is this correct?\r\n")
			if aborted {
				return 0, true
			}
			if answer {
				break
			}
		}
	}
	switch strings.ToLower(usertype) {
	case "":
		break
	case TT_USERTYPE_DEFAULT_STR:
		utype = TT_USERTYPE_DEFAULT
	case TT_USERTYPE_ADMIN_STR:
		utype = TT_USERTYPE_ADMIN
	default:
		console_write("User type " + usertype + " unrecognized.")
		return server.Account_usertype_prompt(username, password, "")
	}
	if utype != TT_USERTYPE_ADMIN {
		return utype, false
	}
	if username != "" && password != "" {
		return utype, false
	}
	if utype == TT_USERTYPE_ADMIN {
		answer, aborted := console_read_confirm("You are adding an anonymous account with no password, and giving such an account administrator rights. It is an extreme security risk to have an account with no username or password, which possesses administrator rights. Are you sure this is what you want to do?\r\n")
		if aborted {
			return 0, true
		}
		if !answer {
			console_write("Aborted.")
			return 0, true
		}
	}
	return utype, false
}

func (server *tt_server) Account_userrights_prompt(rights int) (int, bool) {
	// Modify this to ask if the user wants to use the currently set user rights.
	// Create an option for the user rights in the config file.
	if rights != 0 {
		answer, aborted := console_read_confirm("Currently set user rights: " + teamtalk_flags_userrights_str(rights) + ".\r\nWould you like to configure the user rights now, or use the rights that are currently set?\r\n")
		if aborted {
			return rights, true
		}
		if !answer {
			return rights, false
		}
	}
	return teamtalk_flags_userrights_menu(rights)
}

func (server *tt_server) Message_info(msg_type int, usr_src, usr_dest *tt_user, ch *tt_channel, msg_content string) {
	if server.Cmdid_read() != 0 {
		return
	}
	msg_type_str := teamtalk_flags_message_type_str(msg_type)
	log_from := ""
	log_to := ""
	log_intercept := ""
	switch msg_type {
	case TT_MSGTYPE_BROADCAST:
		log_to = msg_type_str + " message sent.\r\n" + msg_content
		log_nick_src := usr_src.NickName_log()
		if server.Uid_read() == usr_src.Uid_read() {
			server.Log_username_set(usr_src.UserName_read())
			server.Log_write_files(log_nick_src + ": " + log_to)
			server.Log_console(log_to, true)
		} else {
			log_from = msg_type_str + " message received from " + log_nick_src + ":\r\n" + msg_content
			server.Log_write(log_from, true)
			server.Log_username_set(usr_src.UserName_read())
			server.Log_write_account(log_nick_src + ": " + log_to)
		}
		return
	case TT_MSGTYPE_CHANNEL:
		log_nick_src := usr_src.NickName_log()
		cpath := ch.Path_read()
		bot_usr := server.User_find_id(server.Uid_read())
		if server.Uid_read() == usr_src.Uid_read() {
			log_to = msg_type_str + " message sent"
			if ch == bot_usr.Channel_read() {
				log_to += ".\r\n" + msg_content
			} else {
				log_to += " to " + cpath + ":\r\n" + msg_content
			}
			server.Log_username_set(usr_src.UserName_read())
			server.Log_write_files(log_nick_src + ": " + log_to)
			server.Log_console(log_to, false)
		} else {
			if ch == bot_usr.Channel_read() {
				log_from = msg_type_str + " message received from " + log_nick_src + ":\r\n" + msg_content
			} else {
				log_from = msg_type_str + " message received from " + log_nick_src + " to " + cpath + ":\r\n" + msg_content
			}
			server.Log_write(log_from, false)
			server.Log_username_set(usr_src.UserName_read())
			server.Log_write_account(log_nick_src + ": " + msg_type_str + " message sent to " + cpath + ":\r\n" + msg_content)
		}
		return
	case TT_MSGTYPE_USER, TT_MSGTYPE_CUSTOM:
		log_nick_src := usr_src.NickName_log()
		log_nick_dest := usr_dest.NickName_log()
		log_to = msg_type_str + " message sent to " + log_nick_dest + ":\r\n" + msg_content
		log_from = msg_type_str + " message received from " + log_nick_src + ":\r\n" + msg_content
		log_intercept = msg_type_str + " message from " + log_nick_src + " to " + log_nick_dest + ":\r\n" + msg_content
		if usr_dest.Uid_read() == server.Uid_read() {
			// The bot received a message.
			server.Log_console(log_from, true)
			server.Log_username_set(usr_src.UserName_read())
			server.Log_write_files(log_nick_src + ": " + log_to)
			if usr_src.UserName_read() != usr_dest.UserName_read() {
				server.Log_username_set(usr_dest.UserName_read())
				server.Log_write_account(log_nick_dest + ": " + log_from)
			}
			// End bot receiving a message.
		} else if usr_src.Uid_read() == server.Uid_read() {
			// Bot sent a private message.
			server.Log_console(log_to, true)
			server.Log_username_set(usr_src.UserName_read())
			server.Log_write_files(log_nick_src + ": " + log_to)
			if usr_src.UserName_read() != usr_dest.UserName_read() {
				server.Log_username_set(usr_dest.UserName_read())
				server.Log_write_account(log_nick_dest + ": " + log_from)
			}
			// End of bot sending private message.
		} else {
			// Bot intercepted a private message.
			username_src := usr_src.UserName_read()
			username_dest := usr_dest.UserName_read()
			server.Log_username_set(username_src)
			if username_src == username_dest {
				server.Log_write_files(log_nick_src + ": " + log_to)
				server.Log_console(log_intercept, false)
			} else {
				server.Log_write_account(log_nick_src + ": " + log_to)
				server.Log_username_set(username_dest)
				server.Log_write_account(log_nick_dest + ": " + log_from)
				server.Log_write(log_intercept, false)
			}
			// End of bot intercepting a private message.
		}
		return
	}
}

func (server *tt_server) Uid_read() int {
	defer server.Unlock()
	server.Lock()
	return server.uid
}

func (server *tt_server) Uid_set(uid int) {
	defer server.Unlock()
	server.Lock()
	server.uid = uid
}

func (server *tt_server) User_rights_read() int {
	defer server.Unlock()
	server.Lock()
	return server.user_rights
}

func (server *tt_server) User_rights_set(user_rights int) {
	defer server.Unlock()
	server.Lock()
	server.user_rights = user_rights
}

func (server *tt_server) User_rights_check(rights int) bool {
	if server.User_type_read() == TT_USERTYPE_ADMIN || teamtalk_flags_read(server.User_rights_read(), rights) {
		return true
	}
	return false
}

func (server *tt_server) User_type_read() int {
	defer server.Unlock()
	server.Lock()
	return server.user_type
}

func (server *tt_server) User_type_set(user_type int) {
	defer server.Unlock()
	server.Lock()
	server.user_type = user_type
}

func (server *tt_server) Protocol_read() string {
	defer server.Unlock()
	server.Lock()
	return server.protocol
}

func (server *tt_server) Protocol_set(protocol string) {
	defer server.Unlock()
	server.Lock()
	server.protocol = protocol
}

func (server *tt_server) UserTimeout_read() int {
	defer server.Unlock()
	server.Lock()
	return server.usertimeout
}

func (server *tt_server) UserTimeout_set(usertimeout int) {
	defer server.Unlock()
	server.Lock()
	server.usertimeout = usertimeout
}

func (server *tt_server) Motd_read() string {
	defer server.Unlock()
	server.Lock()
	return server.motd
}

func (server *tt_server) Motd_set(motd string) {
	defer server.Unlock()
	server.Lock()
	server.motd = motd
}

func (server *tt_server) Version_read() string {
	defer server.Unlock()
	server.Lock()
	return server.version
}

func (server *tt_server) Version_set(version string) {
	defer server.Unlock()
	server.Lock()
	server.version = version
}

func (server *tt_server) AccountName_read() string {
	defer server.Unlock()
	server.Lock()
	return server.AccountName
}

func (server *tt_server) AccountName_set(name string) {
	defer server.Unlock()
	server.Lock()
	server.AccountName = name
}

func (server *tt_server) AccountPassword_read() string {
	defer server.Unlock()
	server.Lock()
	return server.AccountPassword
}

func (server *tt_server) AccountPassword_set(password string) {
	defer server.Unlock()
	server.Lock()
	server.AccountPassword = password
}

func (server *tt_server) NickName_read() string {
	defer server.Unlock()
	server.Lock()
	return server.NickName
}

func (server *tt_server) NickName_set(name string) {
	defer server.Unlock()
	server.Lock()
	server.NickName = name
}

func (server *tt_server) UseGlobalNickName_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.UseGlobalNickName
}

func (server *tt_server) UseGlobalNickName_set(UseGlobalNickName bool) {
	defer server.Unlock()
	server.Lock()
	server.UseGlobalNickName = UseGlobalNickName
}

func (server *tt_server) Name_read() string {
	defer server.Unlock()
	server.Lock()
	return server.name
}

func (server *tt_server) Name_set(name string) {
	defer server.Unlock()
	server.Lock()
	server.name = name
}

func (server *tt_server) MaxUsers_read() int {
	defer server.Unlock()
	server.Lock()
	return server.maxusers
}

func (server *tt_server) MaxUsers_set(max int) {
	defer server.Unlock()
	server.Lock()
	server.maxusers = max
}

func (server *tt_server) AutoConnectOnStart_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.AutoConnectOnStart
}

func (server *tt_server) AutoConnectOnStart_set(autoConnect bool) {
	defer server.Unlock()
	server.Lock()
	server.AutoConnectOnStart = autoConnect
}

func (server *tt_server) AutoConnectOnDisconnect_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.AutoConnectOnDisconnect
}

func (server *tt_server) AutoConnectOnDisconnect_set(autoConnect bool) {
	defer server.Unlock()
	server.Lock()
	server.AutoConnectOnDisconnect = autoConnect
}

func (server *tt_server) AutoConnectOnKick_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.AutoConnectOnKick
}

func (server *tt_server) AutoConnectOnKick_set(autoConnect bool) {
	defer server.Unlock()
	server.Lock()
	server.AutoConnectOnKick = autoConnect
}

func (server *tt_server) Kicked_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.kicked
}

func (server *tt_server) Kicked_set(kicked bool) {
	defer server.Unlock()
	server.Lock()
	server.kicked = kicked
}

func (server *tt_server) AutoSubscriptions_read() int {
	defer server.Unlock()
	server.Lock()
	return server.AutoSubscriptions
}

func (server *tt_server) AutoSubscriptions_read_str() string {
	return teamtalk_flags_subscriptions_str(server.AutoSubscriptions_read())
}

func (server *tt_server) AutoSubscriptions_set(subs int) {
	defer server.Unlock()
	server.Lock()
	server.AutoSubscriptions = subs
}

func (server *tt_server) AutoMoveFrom_read() int {
	defer server.Unlock()
	server.Lock()
	return server.autoMoveFrom
}

func (server *tt_server) AutoMoveFrom_set(cid int) {
	defer server.Unlock()
	server.Lock()
	server.autoMoveFrom = cid
}

func (server *tt_server) AutoMoveFrom_config_read() int {
	defer server.Unlock()
	server.Lock()
	return server.AutoMoveFrom
}

func (server *tt_server) AutoMoveFrom_config_set(cid int) {
	defer server.Unlock()
	server.Lock()
	server.AutoMoveFrom = cid
}

func (server *tt_server) AutoMoveTo_read() int {
	defer server.Unlock()
	server.Lock()
	return server.autoMoveTo
}

func (server *tt_server) AutoMoveTo_set(cid int) {
	defer server.Unlock()
	server.Lock()
	server.autoMoveTo = cid
}

func (server *tt_server) AutoMoveTo_config_read() int {
	defer server.Unlock()
	server.Lock()
	return server.AutoMoveTo
}

func (server *tt_server) AutoMoveTo_config_set(cid int) {
	defer server.Unlock()
	server.Lock()
	server.AutoMoveTo = cid
}

func (server *tt_server) AutoMove_config() {
	if server.AutoMoveFrom_read() == 0 && server.AutoMoveFrom_config_read() != 0 {
		server.AutoMoveFrom_set(server.AutoMoveFrom_config_read())
	}
	if server.AutoMoveTo_read() == 0 && server.AutoMoveTo_config_read() != 0 {
		server.AutoMoveTo_set(server.AutoMoveTo_config_read())
	}
}

func (server *tt_server) AutoMove_enabled() bool {
	server.AutoMove_config()
	autoMoveFrom := server.AutoMoveFrom_read()
	autoMoveTo := server.AutoMoveTo_read()
	if autoMoveFrom != 0 && autoMoveTo != 0 {
		return true
	} else if autoMoveFrom == 0 && autoMoveTo != 0 {
		return true
	} else if autoMoveFrom != 0 && autoMoveTo == 0 {
		server.Log_write("Incorrect automove settings found. Disabling automatic user moving.", true)
		server.AutoMove_clear()
		return false
	}
	return false
}

func (server *tt_server) AutoMove_clear() {
	server.AutoMoveFrom_set(0)
	server.AutoMoveTo_set(0)
	server.AutoMoveFrom_config_set(0)
	server.AutoMoveTo_config_set(0)
}

func (server *tt_server) automove(usr *tt_user) {
	if usr == nil {
		return
	}
	if usr == server.User_find_id(server.Uid_read()) {
		return
	}
	autoMoveFrom := server.AutoMoveFrom_read()
	autoMoveTo := server.AutoMoveTo_read()
	if autoMoveFrom == 0 && autoMoveTo == 0 {
		return
	}
	if autoMoveFrom != 0 && autoMoveTo == 0 {
		server.Log_write("Incorrect automove settings found. Disabling automatic user moving.", true)
		server.AutoMove_clear()
		return
	}
	ch_dest := server.Channel_find_id(autoMoveTo)
	lnickname := usr.NickName_log()
	ch := usr.Channel_read()
	if autoMoveFrom != 0 || autoMoveTo != 0 {
		if !server.User_rights_check(TT_USERRIGHT_MOVE_USERS) {
			server.Log_write("Insufficient user rights. Disabling automatic user moving.", true)
			server.AutoMove_clear()
			return
		}
	}
	if autoMoveFrom == 0 && autoMoveTo != 0 {
		if ch_dest == nil {
			server.Log_write("Unable to find destination channel. Disabling automatic moving.", true)
			server.AutoMove_clear()
			return
		}
		if ch != nil && ch == ch_dest {
			return
		}
		if ch != nil {
			return
		}
		time.Sleep(time.Millisecond * 500)
		ch = usr.Channel_read()
		if ch == ch_dest {
			return
		}
		res := server.cmd_move_user(usr.Uid_read(), ch_dest.Id_read())
		if !res {
			server.Log_console("Automatic user move for "+lnickname+" failed.", true)
			return
		}
		server.Log_console(lnickname+" automatically moved to "+ch_dest.Path_read(), false)
		return
	}
	if autoMoveFrom != 0 && autoMoveTo != 0 {
		ch_src := server.Channel_find_id(autoMoveFrom)
		if ch_src == nil {
			server.Log_write("Unable to find source channel. Disabling automatic moving.", true)
			server.AutoMove_clear()
			return
		}
		if ch_dest == nil {
			server.Log_write("Unable to find destination channel. Disabling automatic moving.", true)
			server.AutoMove_clear()
			return
		}
		if ch_src == ch_dest {
			server.Log_write("Source and destination channels are the same. Disabling automatic moving.", true)
			server.AutoMove_clear()
			return
		}
		if ch == ch_dest {
			return
		}
		if ch != ch_src {
			return
		}
		res := server.cmd_move_user(usr.Uid_read(), ch_dest.Id_read())
		if !res {
			server.Log_console("Automatic user move for "+lnickname+" failed.", true)
			return
		}
		server.Log_console(lnickname+" automatically moved from "+ch_src.Path_read()+" to "+ch_dest.Path_read(), false)
	}
}

func (server *tt_server) DisplayExtendedConnInfo_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.DisplayExtendedConnInfo
}

func (server *tt_server) DisplayExtendedConnInfo_set(info bool) {
	defer server.Unlock()
	server.Lock()
	server.DisplayExtendedConnInfo = info
}

func (server *tt_server) DisplayStatusUpdates_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.DisplayStatusUpdates
}

func (server *tt_server) DisplayStatusUpdates_set(status bool) {
	defer server.Unlock()
	server.Lock()
	server.DisplayStatusUpdates = status
}

func (server *tt_server) DisplaySubscriptionUpdates_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.DisplaySubscriptionUpdates
}

func (server *tt_server) DisplaySubscriptionUpdates_set(substatus bool) {
	defer server.Unlock()
	server.Lock()
	server.DisplaySubscriptionUpdates = substatus
}

func (server *tt_server) Cmd_sent_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.cmd_sent
}

func (server *tt_server) Cmd_sent_set(cmd_sent bool) {
	defer server.Unlock()
	server.Lock()
	server.cmd_sent = cmd_sent
}

func (server *tt_server) Logged_in_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.logged_in
}

func (server *tt_server) Logged_in_set(logged_in bool) {
	defer server.Unlock()
	server.Lock()
	server.logged_in = logged_in
}

func (server *tt_server) BeepOnCriticalEvents_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.BeepOnCriticalEvents
}

func (server *tt_server) BeepOnCriticalEvents_set(beepOnCriticalEvents bool) {
	defer server.Unlock()
	server.Lock()
	server.BeepOnCriticalEvents = beepOnCriticalEvents
}

func (server *tt_server) Debug_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.Debug
}

func (server *tt_server) Debug_set(debug bool) {
	defer server.Unlock()
	server.Lock()
	server.Debug = debug
}

func (server *tt_server) DisplayEvents_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.DisplayEvents
}

func (server *tt_server) DisplayEvents_set(displayEvents bool) {
	defer server.Unlock()
	server.Lock()
	server.DisplayEvents = displayEvents
}

func (server *tt_server) LogEvents_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.LogEvents
}

func (server *tt_server) LogEvents_set(logEvents bool) {
	defer server.Unlock()
	server.Lock()
	server.LogEvents = logEvents
}

func (server *tt_server) LogEventsAccount_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.LogEventsAccount
}

func (server *tt_server) LogEventsAccount_set(logEventsAccount bool) {
	defer server.Unlock()
	server.Lock()
	server.LogEventsAccount = logEventsAccount
}

func (server *tt_server) Cmderror_read() error {
	defer server.Unlock()
	server.Lock()
	return server.cmderror
}

func (server *tt_server) Cmderror_set(err error) {
	defer server.Unlock()
	server.Lock()
	server.cmderror = err
}

func (server *tt_server) Cmderror_set_str(err string) {
	defer server.Unlock()
	server.Lock()
	server.cmderror = errors.New(err)
}

func (server *tt_server) Cmderror_clear() {
	defer server.Unlock()
	server.Lock()
	server.cmderror = nil
}

func (server *tt_server) Cmdid_read() int {
	defer server.Unlock()
	server.Lock()
	return server.cmdid
}

func (server *tt_server) Cmdid_set(id int) {
	defer server.Unlock()
	server.Lock()
	server.cmdid = id
}

func (server *tt_server) Cmdid_add() int {
	defer server.Unlock()
	server.Lock()
	server.cmdid_add++
	return server.cmdid_add
}

func (server *tt_server) Send(cmd string, genid bool) (bool, error) {
	if genid {
		cmd += " id=" + strconv.Itoa(server.Cmdid_add())
	}
	// Ensures we can't send a command until the other has finished first.
	server.cl.Lock()
	defer server.cl.Unlock()
	server.Cmd_sent_set(true)
	err := server.Write(cmd + "\r\n")
	if err != nil {
		server.Cmd_sent_set(false)
		return false, err
	}
	<-server.cmdfinish
	server.Cmd_sent_set(false)
	err = server.Cmderror_read()
	res := true
	if err != nil {
		res = false
	}
	return res, err
}

func (server *tt_server) KeepAlive(secs int) {
	server.UserTimeout_set(secs)
	ms := secs * 1000
	if secs == 0 {
		ms = 400
	}
	ms /= 2
	server.Lock()
	if server.keepaliveon {
		server.keepaliveon = false
		server.keepalive.Stop()
		server.keepalivedone <- true
	}
	server.keepalive = time.NewTicker(time.Duration(ms) * time.Millisecond)
	server.keepaliveon = true
	server.Unlock()
	for {
		select {
		case <-server.keepalive.C:
			server.cmd_ping()
		case <-server.keepalivedone:
			return
		}
	}
}

func (server *tt_server) CheckEvents(secs int) {
	ms := secs * 1000
	if secs == 0 {
		ms = 400
	}
	ms /= 2
	server.Lock()
	if server.checkeventson {
		server.checkeventson = false
		server.checkevents.Stop()
		server.checkeventsdone <- true
	}
	server.checkevents = time.NewTicker(time.Duration(ms) * time.Millisecond)
	server.checkeventson = true
	server.Unlock()
	for {
		select {
		case <-server.checkevents.C:
			if server.User_type_read() == TT_USERTYPE_ADMIN {
				server.cmd_list_accounts()
			}
			if server.User_rights_check(TT_USERRIGHT_BAN_USERS) {
				server.cmd_list_bans()
			}
		case <-server.checkeventsdone:
			return
		}
	}
}

func (server *tt_server) disconnect_silent() bool {
	if !server.connected() {
		return false
	}
	server.Lock()
	server.conn.Close()
	server.conn = nil
	if server.keepaliveon {
		server.keepaliveon = false
		server.keepalive.Stop()
		server.keepalivedone <- true
	}
	if server.checkeventson {
		server.checkeventson = false
		server.checkevents.Stop()
		server.checkeventsdone <- true
	}
	server.Unlock()
	return true
}

func (server *tt_server) disconnect() bool {
	if !server.disconnect_silent() {
		return false
	}
	server.Log_reset()
	server.Log_write("Disconnected.", true)
	return true
}

func (server *tt_server) Write(str string) error {
	if str == "" {
		return errors.New("Sending empty string not allowed.")
	}
	server.Lock()
	server.writer.WriteString(str)
	err := server.writer.Flush()
	server.Unlock()
	if err != nil {
		server.Log_debug("Error sending data:\r\n" + err.Error())
		server.disconnect()
	} else {
		server.Log_debug("Sent data:\r\n" + str)
	}
	return err
}

func (server *tt_server) User_exists(id int) bool {
	defer server.Unlock()
	server.Lock()
	_, exists := server.users[id]
	return exists
}

func (server *tt_server) User_add(id int) *tt_user {
	if server.User_exists(id) {
		return nil
	}
	defer server.Unlock()
	server.Lock()
	server.users[id] = NewUser(id, server)
	return server.users[id]
}

func (server *tt_server) User_remove(id int) bool {
	if !server.User_exists(id) {
		return false
	}
	server.Lock()
	usr := server.users[id]
	delete(server.users, id)
	server.Unlock()
	if ch := usr.Channel_read(); ch != nil {
		ch.User_remove(usr)
	}
	return true
}

func (server *tt_server) User_find_id(id int) *tt_user {
	if !server.User_exists(id) {
		return nil
	}
	defer server.Unlock()
	server.Lock()
	return server.users[id]
}

func (server *tt_server) Users_sort(uid int) []*tt_user {
	ids := []int{}
	defer server.Unlock()
	server.Lock()
	for id := range server.users {
		if id != uid {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	users := []*tt_user{}
	for _, id := range ids {
		users = append(users, server.users[id])
	}
	return users
}

func (server *tt_server) User_find_nickname(name string) []*tt_user {
	users := []*tt_user{}
	for _, user := range server.Users_sort(server.Uid_read()) {
		if strings.ToLower(user.NickName_read()) == strings.ToLower(name) || strings.Contains(strings.ToLower(user.NickName_read()), strings.ToLower(name)) {
			users = append(users, user)
		}
	}
	return users
}

func (server *tt_server) User_find_username(name string) []*tt_user {
	users := []*tt_user{}
	for _, user := range server.Users_sort(server.Uid_read()) {
		if strings.ToLower(user.UserName_read()) == strings.ToLower(name) || strings.Contains(strings.ToLower(user.UserName_read()), strings.ToLower(name)) {
			users = append(users, user)
		}
	}
	return users
}

func (server *tt_server) User_find_all(name string) []*tt_user {
	users := []*tt_user{}
	for _, user := range server.Users_sort(server.Uid_read()) {
		if name == "" {
			users = append(users, user)
			continue
		}
		if strings.ToLower(user.NickName_read()) == strings.ToLower(name) || strings.Contains(strings.ToLower(user.NickName_read()), strings.ToLower(name)) {
			users = append(users, user)
			continue
		}
		if strings.ToLower(user.UserName_read()) == strings.ToLower(name) || strings.Contains(strings.ToLower(user.UserName_read()), strings.ToLower(name)) {
			users = append(users, user)
			continue
		}
	}
	return users
}

func (server *tt_server) Channels_read() map[int]*tt_channel {
	defer server.Unlock()
	server.Lock()
	return server.channels
}

func (server *tt_server) Users_read() map[int]*tt_user {
	defer server.Unlock()
	server.Lock()
	return server.users
}

func (server *tt_server) Channel_exists(id int) bool {
	defer server.Unlock()
	server.Lock()
	_, exists := server.channels[id]
	return exists
}

func (server *tt_server) Channel_add(id, idparent int) *tt_channel {
	if server.Channel_exists(id) {
		return nil
	}
	defer server.Unlock()
	server.Lock()
	server.channels[id] = NewChannel(id, idparent, server)
	return server.channels[id]
}

func (server *tt_server) Channel_remove(id int) bool {
	if !server.Channel_exists(id) {
		return false
	}
	server.Lock()
	delete(server.channels, id)
	server.Unlock()
	return true
}

func (server *tt_server) Channel_find_id(id int) *tt_channel {
	if !server.Channel_exists(id) {
		return nil
	}
	defer server.Unlock()
	server.Lock()
	return server.channels[id]
}

func (server *tt_server) Channels_sort() []*tt_channel {
	ids := []int{}
	defer server.Unlock()
	server.Lock()
	for id := range server.channels {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	channels := []*tt_channel{}
	for _, id := range ids {
		channels = append(channels, server.channels[id])
	}
	return channels
}

func (server *tt_server) Channel_find_name(name string) []*tt_channel {
	channels := []*tt_channel{}
	for _, channel := range server.Channels_sort() {
		if name == "" {
			channels = append(channels, channel)
			continue
		}
		if strings.ToLower(channel.Name_read()) == strings.ToLower(name) || strings.Contains(strings.ToLower(channel.Name_read()), strings.ToLower(name)) {
			channels = append(channels, channel)
		}
	}
	return channels
}

func (server *tt_server) Channel_find_path(path string) []*tt_channel {
	channels := []*tt_channel{}
	for _, channel := range server.Channels_sort() {
		if path == "" {
			channels = append(channels, channel)
			continue
		}
		if path == "/" && channel.Path_read() == path {
			channels = append(channels, channel)
			break
		}
		if strings.ToLower(channel.Path_read()) == strings.ToLower(path) || strings.Contains(strings.ToLower(channel.Path_read()), strings.ToLower(path)) {
			channels = append(channels, channel)
		}
	}
	return channels
}

func (server *tt_server) Shutdown() {
	if server.Shutdown_read() {
		return
	}
	server.Lock()
	server.shutdown = true
	server.Unlock()
	if server.connected() {
		server.cl.Lock()
		server.Write("quit\r\n")
		server.Cmd_sent_set(true)
		server.disconnect()
		server.Cmd_sent_set(false)
		server.cl.Unlock()
	}
}

func (server *tt_server) connected() bool {
	defer server.Unlock()
	server.Lock()
	if server.conn != nil {
		return true
	}
	return false
}

func (server *tt_server) DisplayName_read() string {
	defer server.Unlock()
	server.Lock()
	return server.DisplayName
}

func (server *tt_server) DisplayName_set(name string) {
	if name == "" {
		return
	}
	server.Lock()
	if strings.ToLower(name) == strings.ToLower(server.DisplayName) {
		server.Unlock()
		return
	}
	server.DisplayName = name
	server.Unlock()
	if server.Config().Server_active_read() == server {
		server.Config().Server_active_set(server)
	}
	return
}

func (server *tt_server) Host_read() string {
	defer server.Unlock()
	server.Lock()
	return server.Host
}

func (server *tt_server) Host_set(host string) {
	defer server.Unlock()
	server.Lock()
	server.Host = host
}

func (server *tt_server) Tcpport_read() string {
	defer server.Unlock()
	server.Lock()
	return server.Tcpport
}

func (server *tt_server) Tcpport_set(port string) {
	defer server.Unlock()
	server.Lock()
	server.Tcpport = port
}

func (server *tt_server) Ip_read() string {
	defer server.Unlock()
	server.Lock()
	return strings.Trim(server.ip, "[]")
}

func (server *tt_server) Shutdown_read() bool {
	defer server.Unlock()
	server.Lock()
	return server.shutdown
}

func (server *tt_server) Log_username_read() string {
	defer server.Unlock()
	server.Lock()
	return server.log_username
}

func (server *tt_server) Log_username_set(name string) {
	defer server.Unlock()
	server.Lock()
	server.log_username = name
}

func (server *tt_server) Log_write(data string, critical bool) {
	data = strings.Trim(data, "\r\n") + "\r\n"
	if data == "" {
		return
	}
	server.Log_write_files(data)
	server.Log_console(data, critical)
}

func (server *tt_server) Log_write_files(data string) {
	data = strings.Trim(data, "\r\n")
	if data == "" {
		return
	}
	server.Log_write_main(data)
	server.Log_write_account(data)
}

func (server *tt_server) Log_write_main(data string) bool {
	data = strings.Trim(data, "\r\n")
	if data == "" {
		return false
	}
	date := server.Config().Log_timestamp_init() + "\r\n"
	server.Lock()
	if date != server.log_timestamp {
		server.log_timestamp = date
	} else {
		date = ""
	}
	server.Unlock()
	if !server.LogEvents_read() {
		return false
	}
	path := server.Log_path_bass()
	if path == "" {
		return false
	}
	if err := file_write(path+"server.log", date+data+"\r\n\r\n"); err != nil {
		server.Log_console("Error writing to server log.\r\n"+err.Error(), true)
		return false
	}
	return true
}

func (server *tt_server) Log_write_account(data string) bool {
	data = strings.Trim(data, "\r\n")
	if data == "" {
		return false
	}
	if !server.LogEvents_read() {
		return false
	}
	if !server.LogEventsAccount_read() {
		return false
	}
	path := server.Log_path_account()
	if path == "" {
		return false
	}
	username := server.Log_username_read()
	if username == "" {
		return false
	}
	server.Log_username_set("")
	date := server.Config().Log_timestamp_init() + "\r\n"
	_, timestampexists := server.log_timestamp_account[username]
	server.Lock()
	if !timestampexists || date != server.log_timestamp_account[username] {
		server.log_timestamp_account[username] = date
	} else {
		date = ""
	}
	server.Unlock()
	if err := file_write(path+username+".log", date+data+"\r\n\r\n"); err != nil {
		server.Log_console("Error writing to account log.\r\n"+err.Error(), true)
		return false
	}
	return true
}

func (server *tt_server) Log_path_bass() string {
	ps := string(filepath.Separator)
	path, err := filepath.Abs(wd + ps + "logs" + ps + server.DisplayName_read())
	if err != nil {
		return ""
	}
	return path + ps
}

func (server *tt_server) Log_path_account() string {
	path := server.Log_path_bass()
	if path == "" {
		return ""
	}
	return path + "account_logs" + string(filepath.Separator)
}

func (server *tt_server) Log_console_timestamp() string {
	if !server.Config().DisplayTimestamp_read() {
		return ""
	}
	date := server.Config().Log_timestamp_init() + "\r\n"
	if date != server.Config().Timestamp_console_read() {
		server.Config().Timestamp_console_set(date)
		return date
	}
	return ""
}

func (server *tt_server) Log_console(data string, critical bool) {
	header := ""
	if critical && server.BeepOnCriticalEvents_read() {
		header = TT_BEEP
	}
	name := server.DisplayName_read()
	if server.Config().Server_active_read() != server || critical || server.Config().Logged_console_read() != name {
		if server.Config().Logged_console_read() != name {
			header += "[" + name + "]: "
		}
	}
	date := server.Log_console_timestamp()
	if critical || server.DisplayEvents_read() || server == server.Config().Server_active_read() {
		if critical || server.Cmdid_read() == TT_CMD_NONE {
			server.Config().Logged_console_set(name)
			console_write(date + header + data)
		}
	}
	server.Log_history_store(date + data)
}

func (server *tt_server) Log_history_store(data string) {
	data = strings.Trim(data, "\r\n")
	if data == "" {
		return
	}
	server.Lock()
	defer server.Unlock()
	if server.log_history == nil {
		server.log_history = []string{}
		server.log_history_number = 10
	}
	log_history := server.log_history
	log_history_number := server.log_history_number
	if len(log_history) == log_history_number {
		server.log_history = log_history[1:]
	}
	server.log_history = append(server.log_history, data)
}

func (server *tt_server) Log_history_read() []string {
	server.Lock()
	defer server.Unlock()
	return server.log_history
}

func (server *tt_server) Log_debug(data string) {
	if !server.Debug_read() || data == "" {
		return
	}
	server.Log_write(data, true)
}

func (server *tt_server) Log(data string) {
	if data == "" {
		return
	}
	defer server.Unlock()
	server.Lock()
	server.log_buffer += data + "\r\n"
}

func (server *tt_server) Log_send() {
	server.Lock()
	data := strings.TrimSuffix(server.log_buffer, "\r\n")
	server.Unlock()
	if data == "" {
		return
	}
	server.Log_write(data, false)
}

func (server *tt_server) Log_reset() {
	server.Lock()
	if server.log_buffer != "" {
		server.log_buffer = ""
	}
	server.Unlock()
	server.Log_username_set("")
}

func (server *tt_server) Info_str() string {
	str := "Name: " + server.DisplayName_read() + "\r\n"
	str += "Host: " + server.Host_read() + "\r\n"
	str += "TCP port: " + server.Tcpport_read() + "\r\n"
	NickName := server.NickName_read()
	if NickName == "" && server.UseGlobalNickName_read() && server.Config().NickName_read() != "" {
		NickName = server.Config().NickName_read()
	}
	if NickName != "" {
		str += "Nickname: " + NickName + "\r\n"
	}
	if AccountName := server.AccountName_read(); AccountName != "" {
		str += "Username: " + AccountName + "\r\n"
	}
	if AccountPassword := server.AccountPassword_read(); AccountPassword != "" {
		str += "Password: " + AccountPassword + "\r\n"
	}
	str += "Automatically connect on start: " + str_yes_no(server.AutoConnectOnStart_read()) + "\r\n"
	str += "Automatically reconnect on disconnect: " + str_yes_no(server.AutoConnectOnDisconnect_read()) + "\r\n"
	str += "Automatically reconnect when kicked: " + str_yes_no(server.AutoConnectOnKick_read()) + "\r\n"
	if sub_str := server.AutoSubscriptions_read_str(); sub_str != "" {
		str += "Current automatic local subscriptions: " + sub_str + "\r\n"
	}
	str += "Display extended connection info: " + str_yes_no(server.DisplayExtendedConnInfo_read()) + "\r\n"

	str += "Display status updates: " + str_yes_no(server.DisplayStatusUpdates_read()) + "\r\n"
	str += "Display subscription updates: " + str_yes_no(server.DisplaySubscriptionUpdates_read()) + "\r\n"
	str += "Display server events if inactive: " + str_yes_no(server.DisplayEvents_read()) + "\r\n"
	str += "Beep on critical server events: " + str_yes_no(server.BeepOnCriticalEvents_read()) + "\r\n"
	log_server_events := server.LogEvents_read()
	str += "Log server events: " + str_yes_no(log_server_events) + "\r\n"
	if log_server_events {
		str += "Log events per user account: " + str_yes_no(server.LogEventsAccount_read()) + "\r\n"
	}
	return str
}

func (server *tt_server) Config() *config {
	defer server.Unlock()
	server.Lock()
	if server.config == nil {
		server.config = c
	}
	return server.config
}

// Section for server commands.

func (server *tt_server) cmd_can_send(failed string) bool {
	not_connected := "Not connected to server."
	not_logged_in := "Not logged in to server."
	not_sent := "Failed to send command."
	user_not_found := "Unable to find current user."
	if !server.connected() {
		if failed != "" {
			server.Log_write(failed+" "+not_connected, true)
		} else {
			server.Log_write(not_sent+" "+not_connected, true)
		}
		return false
	}
	if !server.Logged_in_read() {
		if failed != "" {
			server.Log_write(failed+" "+not_logged_in, true)
		} else {
			server.Log_write(not_sent+" "+not_connected, true)
		}
		return false
	}
	usr := server.User_find_id(server.Uid_read())
	if usr == nil {
		if failed != "" {
			server.Log_write(failed+" "+user_not_found, true)
		} else {
			server.Log_write(not_sent+" "+user_not_found, true)
		}
		return false
	}
	return true
}

func (server *tt_server) Login() bool {
	if !server.connected() {
		server.Log_write("Failed to log in. Not connected to server.", true)
		return false
	}
	if server.Logged_in_read() {
		server.Log_write("Failed to log in to server. Already logged in.", true)
		return false
	}
	accountname := server.AccountName_read()
	accountpassword := server.AccountPassword_read()
	nickname := server.NickName_read()
	if nickname == "" && server.UseGlobalNickName_read() {
		nickname = server.Config().NickName_read()
	}
	scmd := teamtalk_format_cmd("login", "username", accountname, "password", accountpassword, "nickname", nickname, "clientname", bot_name, "protocol", protocol_version, "version", Version, "id", strconv.Itoa(TT_CMD_LOGIN))
	res, err := server.Send(scmd, false)
	if err != nil {
		server.Log_console("Login error: "+err.Error(), true)
		server.Shutdown()
		return false
	}
	return res
}

func (server *tt_server) Login_info() {
	if !server.connected() {
		return
	}
	msg := "Server version: " + server.Version_read() + "\r\n"
	users := server.Users_read()
	channels := server.Channels_read()
	numusers := len(users)
	numchannels := len(channels)
	msg += "There "
	switch numusers {
	case 1:
		msg += "is 1 user"
	default:
		msg += "are " + strconv.Itoa(numusers) + " users"
	}
	msg += " currently connected, and " + strconv.Itoa(numchannels) + " channel"
	if numchannels != 1 {
		msg += "s"
	}
	msg += " on this server.\r\n"
	server.Log_console(strings.TrimSuffix(msg, "\r\n"), false)
}

func (server *tt_server) Logout() (bool, error) {
	if !server.connected() {
		return false, errors.New("Not connected.")
	}
	res := true
	err := server.Write("logout\r\n")
	if err != nil {
		res = false
	}
	return res, err
}

func (server *tt_server) autosubscribe(usr *tt_user) bool {
	if usr == nil {
		return false
	}
	autosubs := server.AutoSubscriptions_read()
	if autosubs == 0 {
		return false
	}
	old_subscriptions_local := usr.Subscriptions_local_read()
	if old_subscriptions_local == autosubs {
		return false
	}
	res := server.cmd_changesubscriptions(usr.Uid_read(), autosubs)
	if !res {
		return false
	}
	if old_subscriptions_local != usr.Subscriptions_local_read() && server.DisplaySubscriptionUpdates_read() && usr != server.User_find_id(server.Uid_read()) {
		sub_local_msg := ""

		subs_local_added_str := usr.Subscriptions_local_added_str(old_subscriptions_local)

		subs_local_removed_str := usr.Subscriptions_local_removed_str(old_subscriptions_local)

		if subs_local_added_str != "" {
			sub_local_msg += "Subscriptions added: " + subs_local_added_str + "\r\n"
		}
		if subs_local_removed_str != "" {
			sub_local_msg += "Subscriptions removed: " + subs_local_removed_str + "\r\n"
		}
		if sub_local_msg != "" {
			server.Log_console(usr.NickName_log()+": local subscription change.\r\n"+sub_local_msg, false)
		}
	}
	return true
}

func (server *tt_server) cmd_changenick(nickname string) bool {
	if !server.cmd_can_send("Unable to change nickname.") {
		return false
	}
	oldnick := server.User_find_id(server.Uid_read()).NickName_read()
	if oldnick == nickname {
		server.Log_write("Failed to change nickname: nicknames identical.", true)
		return false
	}
	scmd := teamtalk_format_cmd("changenick", "nickname", nickname)
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to change nickname: "+err.Error(), true)
	}
	return res
}

func (server *tt_server) cmd_changestatus(mode int, msg string) bool {
	if !server.cmd_can_send("Unable to change status.") {
		return false
	}
	usr := server.User_find_id(server.Uid_read())
	if mode == usr.StatusMode_read() && msg == usr.StatusMsg_read() {
		server.Log_write("Failed to change status: status identical.", true)
		return false
	}
	scmd := teamtalk_format_cmd("changestatus", "statusmode", strconv.Itoa(mode), "statusmsg", msg)
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to change status: "+err.Error(), true)
	}
	return res
}

func (server *tt_server) cmd_message_user(uid int, message string) bool {
	if !server.cmd_can_send("Unable to send message.") {
		return false
	}
	if uid == server.Uid_read() {
		server.Log_write("Failed to send message: user ID is the bot.", true)
		return false
	}
	if message == "" {
		server.Log_write("Failed to send message: message empty.", true)
		return false
	}
	msg_type := TT_MSGTYPE_USER
	scmd := teamtalk_format_cmd("message", "type", strconv.Itoa(msg_type), "destuserid", strconv.Itoa(uid), "content", message)
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to send message: "+err.Error(), true)
	}
	if res {
		server.Message_info(msg_type, server.User_find_id(server.Uid_read()), server.User_find_id(uid), nil, message)
	}
	return res
}

func (server *tt_server) cmd_message_channel(cid int, message string) bool {
	if !server.cmd_can_send("Unable to send message.") {
		return false
	}
	if server.Channel_find_id(cid) == nil {
		server.Log_write("Failed to send message: invalid channel.", true)
		return false
	}
	if message == "" {
		server.Log_write("Failed to send message: message empty.", true)
		return false
	}
	msg_type := TT_MSGTYPE_CHANNEL
	scmd := teamtalk_format_cmd("message", "type", strconv.Itoa(msg_type), "chanid", strconv.Itoa(cid), "content", message)
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to send message: "+err.Error(), true)
	}
	if res {
		server.Message_info(msg_type, server.User_find_id(server.Uid_read()), nil, server.Channel_find_id(cid), message)
	}
	return res
}

func (server *tt_server) cmd_message_broadcast(message string) bool {
	if !server.cmd_can_send("Unable to send message.") {
		return false
	}
	if message == "" {
		server.Log_write("Failed to send message: message empty.", true)
		return false
	}
	msg_type := TT_MSGTYPE_BROADCAST
	scmd := teamtalk_format_cmd("message", "type", strconv.Itoa(msg_type), "content", message)
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to send message: "+err.Error(), true)
	}
	if res {
		server.Message_info(msg_type, server.User_find_id(server.Uid_read()), nil, nil, message)
	}
	return res
}

func (server *tt_server) cmd_join(cid int, password string) bool {
	if !server.cmd_can_send("Unable to join channel.") {
		return false
	}
	if server.Channel_find_id(cid) == nil {
		server.Log_write("Failed to join channel: invalid channel id.", true)
		return false
	}
	usr := server.User_find_id(server.Uid_read())
	if usr.Channel_read() != nil && usr.Channel_read().Id_read() == cid {
		server.Log_write("Failed to join channel: already in channel "+strconv.Itoa(cid)+".", true)
		return false
	}
	scmd := teamtalk_format_cmd("join", "chanid", strconv.Itoa(cid), "password", password)
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to join channel: "+err.Error(), true)
	}
	return res
}

func (server *tt_server) cmd_leave() bool {
	if !server.cmd_can_send("Unable to leave channel.") {
		return false
	}
	usr := server.User_find_id(server.Uid_read())
	if usr.Channel_read() == nil {
		server.Log_write("Failed to leave channel: not in a channel.", true)
		return false
	}
	res, err := server.Send("leave", true)
	if err != nil {
		server.Log_write("Failed to leave channel: "+err.Error(), true)
	}
	return res
}

func (server *tt_server) cmd_move_user(uid, cid int) bool {
	if !server.cmd_can_send("Unable to move user.") {
		return false
	}
	usr := server.User_find_id(uid)
	if usr == nil {
		server.Log_write("Failed to move user: invalid user ID.", true)
		return false
	}
	ch := server.Channel_find_id(cid)
	if ch == nil {
		server.Log_write("Failed to move user: invalid channel ID.", true)
		return false
	}
	scmd := teamtalk_format_cmd("moveuser", "userid", strconv.Itoa(uid), "chanid", strconv.Itoa(cid))
	res, err := server.Send(scmd, true)
	if err != nil {
		server.Log_write("Failed to move user: "+err.Error(), true)
	}
	return res
}

func (server *tt_server) cmd_ping() bool {
	if !server.cmd_can_send("Unable to ping server.") {
		return false
	}
	res, err := server.Send("ping", true)
	if err != nil {
		server.Log_write("Failed to ping server: "+err.Error(), true)
	}
	return res
}

func (server *tt_server) cmd_changesubscriptions(uid, subs int) bool {
	if !server.cmd_can_send("Unable to change subscriptions.") {
		return false
	}
	usr := server.User_find_id(uid)
	if usr == nil {
		server.Log_write("Failed to change user subscriptions: invalid user ID.", true)
		return false
	}
	if usr.Subscriptions_local_read() == subs {
		server.Log_write("Failed to change user subscriptions: local subscriptions already match given subscriptions.", true)
		return false
	}
	subscribe := usr.Subscriptions_local_removed(subs)
	unsubscribe := usr.Subscriptions_local_added(subs)
	sub_cmd := ""
	unsub_cmd := ""
	var err error
	if subscribe != 0 {
		sub_cmd = teamtalk_format_cmd("subscribe", "userid", strconv.Itoa(uid), "sublocal", strconv.Itoa(subscribe))
	}

	if unsubscribe != 0 {
		unsub_cmd = teamtalk_format_cmd("unsubscribe", "userid", strconv.Itoa(uid), "sublocal", strconv.Itoa(unsubscribe))
	}
	sub_res := true
	unsub_res := true
	if sub_cmd != "" {
		sub_res, err = server.Send(sub_cmd, true)
		if err != nil {
			server.Log_write("Subscription error: "+err.Error(), true)
		}
	}
	if unsub_cmd != "" {
		unsub_res, err = server.Send(unsub_cmd, true)
		if err != nil {
			server.Log_write("Unsubscription error: "+err.Error(), true)
		}
	}
	if !sub_res && !unsub_res {
		return false
	}
	return true
}

func (server *tt_server) cmd_list_accounts() bool {
	if !server.cmd_can_send("Unable to list user accounts.") {
		return false
	}
	if server.User_type_read() != TT_USERTYPE_ADMIN {
		server.Log_write("Unable to list user accounts. Insufficient permission.", true)
		return false
	}
	res, err := server.Send(teamtalk_format_cmd("listaccounts", "index", "0", "count", "100000", "id", strconv.Itoa(TT_CMD_LIST_ACCOUNTS)), false)
	if !res {
		if err != nil {
			server.Log_write("Failed to list user accounts: "+err.Error(), true)
		}
		return false
	}
	server.Lock()
	if server.accounts == nil {
		server.accounts = server.accounts_cached
		server.accounts_cached = nil
		server.Unlock()
		return true
	}
	defer func() {
		server.Lock()
		server.accounts = server.accounts_cached
		server.accounts_cached = nil
		server.Unlock()
	}()
	added, changed, removed := mapCompare(server.accounts_cached, server.accounts)
	accounts := server.accounts
	server.Unlock()
	if len(added) == 0 && len(changed) == 0 && len(removed) == 0 {
		return true
	}
	msg_added := ""
	msg_changed := ""
	msg_removed := ""
	if len(added) != 0 {
		msg_added = "The following user account"
		if len(added) != 1 {
			msg_added += "s have"
		} else {
			msg_added += " has"
		}
		msg_added += " been added:\r\n"
		names := ""
		for name := range added {
			names += name + ", "
		}
		msg_added += strings.TrimSuffix(names, ", ") + "\r\n"
	}
	if len(changed) != 0 {
		msg_changed = "The following account"
		if len(changed) != 1 {
			msg_changed += "s have"
		} else {
			msg_changed += " has"
		}
		msg_changed += " been changed:\r\n"
		for name := range changed {
			msg_changed += name + "\r\n"
			if password, exists := changed[name]["password"]; exists {
				if old_password := accounts[name]["password"]; old_password != "" {
					msg_changed += "Old password: " + old_password + "\r\nNew password: "
				} else {
					msg_changed += "Password added: "
				}
				msg_changed += password
			}
			if usertype, exists := changed[name]["usertype"]; exists {
				msg_changed += "Old user type: " + accounts[name]["usertype"] + "\r\nNew usertype: " + usertype
			}
			if userrights, exists := changed[name]["userrights"]; exists {
				if old_userrights := accounts[name]["userrights"]; old_userrights != "" {
					if userrights != "" {
						msg_changed += "Old user rights: " + old_userrights + "\r\nNew user rights: " + userrights
					}
				} else {
					if userrights != "" {
						msg_changed += "User rights: " + userrights
					}
				}
			}

		}

	}
	if len(removed) != 0 {
		msg_removed = "The following user account"
		if len(removed) != 1 {
			msg_removed += "s have"
		} else {
			msg_removed += " has"
		}
		msg_removed += " been removed:\r\n"
		names := ""
		for name := range removed {
			names += name + ", "
		}
		msg_removed += strings.TrimSuffix(names, ", ") + "\r\n"
	}
	msg := ""
	if msg_added != "" || msg_changed != "" || msg_removed != "" {
		msg = "User account changes.\r\n" + msg_added + msg_changed + msg_removed
	}
	if msg != "" {
		server.Log_write(msg, true)
	}
	return true
}

func (server *tt_server) cmd_list_bans() bool {
	if !server.cmd_can_send("Unable to list user bans.") {
		return false
	}
	if !server.User_rights_check(TT_USERRIGHT_BAN_USERS) {
		server.Log_write("Unable to list user bans. Insufficient permission.", true)
		return false
	}
	res, err := server.Send(teamtalk_format_cmd("listbans", "index", "0", "count", "1000000", "id", strconv.Itoa(TT_CMD_LIST_BANS)), false)
	if !res {
		if err != nil {
			server.Log_write("Failed to list user bans: "+err.Error(), true)
		}
		return false
	}
	server.Lock()
	if server.bans == nil {
		server.bans = server.bans_cached
		server.bans_cached = nil
		server.Unlock()
		return true
	}
	defer func() {
		server.Lock()
		server.bans = server.bans_cached
		server.bans_cached = nil
		server.Unlock()
	}()
	added, _, removed := mapCompare(server.bans_cached, server.bans)
	server.Unlock()
	if len(added) == 0 && len(removed) == 0 {
		return true
	}
	msg_added := ""
	msg_removed := ""
	if len(added) != 0 {
		msg_added = "The following IP address"
		if len(added) != 1 {
			msg_added += "es are"
		} else {
			msg_added += " is"
		}
		msg_added += " ban:\r\n"
		addrs := ""
		for addr := range added {
			addrs += addr + "\r\n"
		}
		msg_added += addrs
	}
	if len(removed) != 0 {
		msg_removed = "The following IP address"
		if len(removed) != 1 {
			msg_removed += "es are"
		} else {
			msg_removed += " is"
		}
		msg_removed += " no longer ban:\r\n"
		addrs := ""
		for addr := range removed {
			addrs += addr + "\r\n"
		}
		msg_removed += addrs
	}
	msg := ""
	if msg_added != "" || msg_removed != "" {
		msg = "Ban changes.\r\n" + msg_added + msg_removed
	}
	if msg != "" {
		server.Log_write(msg, true)
	}
	return true
}

func (server *tt_server) cmd_new_account(username, password string, utype, urights int) bool {
	if !server.cmd_can_send("Unable to add user account.") {
		return false
	}
	if server.User_type_read() != TT_USERTYPE_ADMIN {
		server.Log_write("Unable to add user account. Insufficient permission.", true)
		return false
	}
	res, err := server.Send(teamtalk_format_cmd("newaccount",
		"username", username,
		"password", password,
		"usertype", strconv.Itoa(utype),
		"userrights", strconv.Itoa(urights)),
		true)
	if !res {
		if err != nil {
			server.Log_write("Failed to list user bans: "+err.Error(), true)
		}
		return false
	}
	server.cmd_list_accounts()
	return true
}
