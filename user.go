package main

import (
	"strconv"
	"sync"
	"time"
)

type tt_user struct {
	sync.Mutex
	uid                  int
	ip                   string
	nickname             string
	username             string
	usertype             int
	subscriptions_local  int
	subscriptions_remote int
	clientname           string
	version              string
	statusmode           int
	statusmsg            string
	channel              *tt_channel
	server               *tt_server
	conntime             time.Time
	conntime_set         bool
}

func NewUser(id int, s *tt_server) *tt_user {
	return &tt_user{
		uid:    id,
		server: s,
	}
}

func (user *tt_user) Uid_read() int {
	defer user.Unlock()
	user.Lock()
	return user.uid
}

func (user *tt_user) Ip_read() string {
	defer user.Unlock()
	user.Lock()
	return user.ip
}

func (user *tt_user) Ip_set(ip string) {
	defer user.Unlock()
	user.Lock()
	user.ip = ip
}

func (user *tt_user) UserName_read() string {
	defer user.Unlock()
	user.Lock()
	return user.username
}

func (user *tt_user) UserName_set(username string) {
	defer user.Unlock()
	user.Lock()
	user.username = username
}

func (user *tt_user) UserType_read() int {
	defer user.Unlock()
	user.Lock()
	return user.usertype
}

func (user *tt_user) UserType_set(utype int) {
	defer user.Unlock()
	user.Lock()
	user.usertype = utype
}

func (user *tt_user) UserType_read_str() string {
	return teamtalk_flags_usertype_str(user.UserType_read())
}

func (user *tt_user) Subscriptions_local_read() int {
	defer user.Unlock()
	user.Lock()
	return user.subscriptions_local
}

func (user *tt_user) Subscriptions_local_set(usubs int) {
	defer user.Unlock()
	user.Lock()
	user.subscriptions_local = usubs
}

func (user *tt_user) Subscribed_local(subscription int) bool {
	return teamtalk_flags_read(user.Subscriptions_local_read(), subscription)
}

func (user *tt_user) Subscriptions_local_read_str() string {
	return teamtalk_flags_subscriptions_str(user.Subscriptions_local_read())
}

func (user *tt_user) Subscriptions_local_added(oldsubs int) int {
	return user.Subscriptions_local_read() &^ oldsubs
}

func (user *tt_user) Subscriptions_local_added_str(oldsubs int) string {
	return teamtalk_flags_subscriptions_str(user.Subscriptions_local_added(oldsubs))
}

func (user *tt_user) Subscriptions_local_removed(oldsubs int) int {
	return oldsubs &^ user.Subscriptions_local_read()
}

func (user *tt_user) Subscriptions_local_removed_str(oldsubs int) string {
	return teamtalk_flags_subscriptions_str(user.Subscriptions_local_removed(oldsubs))
}

func (user *tt_user) Subscriptions_remote_read() int {
	defer user.Unlock()
	user.Lock()
	return user.subscriptions_remote
}

func (user *tt_user) Subscriptions_remote_set(usubs int) {
	defer user.Unlock()
	user.Lock()
	user.subscriptions_remote = usubs
}

func (user *tt_user) Subscribed_remote(subscription int) bool {
	return teamtalk_flags_read(user.Subscriptions_remote_read(), subscription)
}

func (user *tt_user) Subscriptions_remote_read_str() string {
	return teamtalk_flags_subscriptions_str(user.Subscriptions_remote_read())
}

func (user *tt_user) Subscriptions_remote_added(oldsubs int) int {
	return user.Subscriptions_remote_read() &^ oldsubs
}

func (user *tt_user) Subscriptions_remote_added_str(oldsubs int) string {
	return teamtalk_flags_subscriptions_str(user.Subscriptions_remote_added(oldsubs))
}

func (user *tt_user) Subscriptions_remote_removed(oldsubs int) int {
	return oldsubs &^ user.Subscriptions_remote_read()
}

func (user *tt_user) Subscriptions_remote_removed_str(oldsubs int) string {
	return teamtalk_flags_subscriptions_str(user.Subscriptions_remote_removed(oldsubs))
}

func (user *tt_user) ClientName_read() string {
	defer user.Unlock()
	user.Lock()
	return user.clientname
}

func (user *tt_user) ClientName_set(cname string) {
	defer user.Unlock()
	user.Lock()
	user.clientname = cname
}

func (user *tt_user) Version_read() string {
	defer user.Unlock()
	user.Lock()
	return user.version
}

func (user *tt_user) Version_set(version string) {
	defer user.Unlock()
	user.Lock()
	user.version = version
}

func (user *tt_user) NickName_read() string {
	defer user.Unlock()
	user.Lock()
	return user.nickname
}

func (user *tt_user) NickName_set(name string) {
	defer user.Unlock()
	user.Lock()
	user.nickname = name
}

func (user *tt_user) NickName_log() string {
	nickname := user.NickName_read()
	if nickname == "" {
		username := user.UserName_read()
		id := strconv.Itoa(user.Uid_read())
		if username == "" {
			return "#" + id
		}
		return "#" + id + " " + username
	}
	return nickname
}

func (user *tt_user) StatusMode_read() int {
	defer user.Unlock()
	user.Lock()
	return user.statusmode
}

func (user *tt_user) StatusMode_read_str() string {
	return teamtalk_flags_status_mode_str(user.StatusMode_read())
}

func (user *tt_user) StatusMode_set(mode int) {
	defer user.Unlock()
	user.Lock()
	user.statusmode = mode
}

func (user *tt_user) StatusMsg_read() string {
	defer user.Unlock()
	user.Lock()
	return user.statusmsg
}

func (user *tt_user) StatusMsg_set(msg string) {
	defer user.Unlock()
	user.Lock()
	user.statusmsg = msg
}

func (user *tt_user) Channel_read() *tt_channel {
	defer user.Unlock()
	user.Lock()
	return user.channel
}

func (user *tt_user) Channel_set(ch *tt_channel) {
	defer user.Unlock()
	user.Lock()
	user.channel = ch
}

func (user *tt_user) Channel_clear() {
	defer user.Unlock()
	user.Lock()
	user.channel = nil
}

func (user *tt_user) Server_read() *tt_server {
	defer user.Unlock()
	user.Lock()
	return user.server
}

func (user *tt_user) Server_set(s *tt_server) {
	defer user.Unlock()
	user.Lock()
	user.server = s
}

func (user *tt_user) Conntime_isSet() bool {
	defer user.Unlock()
	user.Lock()
	return user.conntime_set
}

func (user *tt_user) Conntime_read() time.Time {
	defer user.Unlock()
	user.Lock()
	return user.conntime
}

func (user *tt_user) Conntime_read_str() string {
	user.Lock()
	duration := time.Now().Sub(user.conntime)
	user.Unlock()
	return time_duration_str(duration)
}

func (user *tt_user) Conntime_set() {
	if user.Conntime_isSet() {
		return
	}
	defer user.Unlock()
	user.Lock()
	user.conntime = time.Now()
	user.conntime_set = true
}
