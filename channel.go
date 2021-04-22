package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

type tt_channel struct {
	sync.Mutex
	id         int
	idparent   int
	name       string
	password   string
	oppassword string
	protected  int
	topic      string
	operators  []int
	quota      int
	maxusers   int
	options    int
	audiocodec []int
	audiocfg   []int
	files      map[int]*tt_file
	users      map[int]*tt_user
	server     *tt_server
}

type tt_file struct {
	name  string
	size  int
	owner string
	id    int
}

func NewChannel(id, idparent int, s *tt_server) *tt_channel {
	return &tt_channel{
		id:       id,
		idparent: idparent,
		server:   s,
		files:    make(map[int]*tt_file, 0),
		users:    make(map[int]*tt_user, 0),
	}
}

func (ch *tt_channel) Server_read() *tt_server {
	defer ch.Unlock()
	ch.Lock()
	return ch.server
}

func (ch *tt_channel) Server_set(s *tt_server) {
	defer ch.Unlock()
	ch.Lock()
	ch.server = s
}

func (ch *tt_channel) Id_read() int {
	defer ch.Unlock()
	ch.Lock()
	return ch.id
}

func (ch *tt_channel) Id_set(id int) {
	defer ch.Unlock()
	ch.Lock()
	ch.id = id
}

func (ch *tt_channel) Idparent_read() int {
	defer ch.Unlock()
	ch.Lock()
	return ch.idparent
}

func (ch *tt_channel) Idparent_set(id int) {
	defer ch.Unlock()
	ch.Lock()
	ch.idparent = id
}

func (ch *tt_channel) Path_read() string {
	ids := []int{}
	channel := ch
	server := ch.server
	for {
		ids = append(ids, channel.Id_read())
		channel = server.Channel_find_id(channel.Idparent_read())
		if channel == nil {
			break
		}
	}
	sort.Ints(ids)
	path := ""
	for _, cid := range ids {
		channel := server.Channel_find_id(cid)
		if channel == nil {
			return ""
		}
		path += channel.Name_read() + "/"
	}
	return path
}

func (ch *tt_channel) Name_read() string {
	defer ch.Unlock()
	ch.Lock()
	return ch.name
}

func (ch *tt_channel) Name_set(name string) {
	defer ch.Unlock()
	ch.Lock()
	ch.name = name
}

func (ch *tt_channel) Topic_read() string {
	defer ch.Unlock()
	ch.Lock()
	return ch.topic
}

func (ch *tt_channel) Topic_set(topic string) {
	defer ch.Unlock()
	ch.Lock()
	ch.topic = topic
}

func (ch *tt_channel) Password_read() string {
	defer ch.Unlock()
	ch.Lock()
	return ch.password
}

func (ch *tt_channel) Password_set(password string) {
	defer ch.Unlock()
	ch.Lock()
	ch.password = password
}

func (ch *tt_channel) Oppassword_read() string {
	defer ch.Unlock()
	ch.Lock()
	return ch.oppassword
}

func (ch *tt_channel) Oppassword_set(password string) {
	defer ch.Unlock()
	ch.Lock()
	ch.oppassword = password
}

func (ch *tt_channel) Protected_read() int {
	defer ch.Unlock()
	ch.Lock()
	return ch.protected
}

func (ch *tt_channel) Protected_set(protected int) {
	defer ch.Unlock()
	ch.Lock()
	ch.protected = protected
}

func (ch *tt_channel) Maxusers_read() int {
	defer ch.Unlock()
	ch.Lock()
	return ch.maxusers
}

func (ch *tt_channel) Maxusers_set(max int) {
	defer ch.Unlock()
	ch.Lock()
	ch.maxusers = max
}

func (ch *tt_channel) Options_read() int {
	defer ch.Unlock()
	ch.Lock()
	return ch.options
}

func (ch *tt_channel) Options_read_str() string {
	return teamtalk_flags_channel_options_str(ch.Options_read())
}

func (ch *tt_channel) Options_set(options int) {
	defer ch.Unlock()
	ch.Lock()
	ch.options = options
}

func (ch *tt_channel) Quota_read() int {
	defer ch.Unlock()
	ch.Lock()
	return ch.quota
}

func (ch *tt_channel) Quota_read_str() string {
	quota := ch.Quota_read()
	if quota == 0 {
		return "0 bytes"
	}
	return strconv.Itoa(quota) + " bytes"
}

func (ch *tt_channel) Quota_set(quota int) {
	defer ch.Unlock()
	ch.Lock()
	ch.quota = quota
}

func (ch *tt_channel) Operators_read() []int {
	defer ch.Unlock()
	ch.Lock()
	return ch.operators
}

func (ch *tt_channel) Operators_read_str() string {
	str := ""
	server := ch.server
	for _, uid := range ch.Operators_read() {
		usr := server.User_find_id(uid)
		if usr == nil {
			return ""
		}
		str += usr.NickName_log() + ", "
	}
	return strings.TrimSuffix(str, ", ")
}

func (ch *tt_channel) Operators_set(ops []int) {
	defer ch.Unlock()
	ch.Lock()
	ch.operators = ops
}

func (ch *tt_channel) Audiocodec_read() []int {
	defer ch.Unlock()
	ch.Lock()
	return ch.audiocodec
}

func (ch *tt_channel) Audiocodec_set(ac []int) {
	defer ch.Unlock()
	ch.Lock()
	ch.audiocodec = ac
}

func (ch *tt_channel) Audiocfg_read() []int {
	defer ch.Unlock()
	ch.Lock()
	return ch.audiocfg
}

func (ch *tt_channel) Audiocfg_set(ac []int) {
	defer ch.Unlock()
	ch.Lock()
	ch.audiocfg = ac
}

func (ch *tt_channel) File_exists(fid int) bool {
	defer ch.Unlock()
	ch.Lock()
	_, exists := ch.files[fid]
	return exists
}

func (ch *tt_channel) File_add(fid int, fname string, fsize int, fowner string) bool {
	if ch.File_exists(fid) {
		return false
	}
	defer ch.Unlock()
	ch.Lock()
	ch.files[fid] = &tt_file{
		id:    fid,
		name:  fname,
		owner: fowner,
		size:  fsize,
	}
	return true
}

func (ch *tt_channel) Files_read() []*tt_file {
	defer ch.Unlock()
	ch.Lock()
	files := []*tt_file{}
	ids := []int{}
	for id := range ch.files {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		files = append(files, ch.files[id])
	}
	return files
}

func (ch *tt_channel) File_find_name(fname string) *tt_file {
	defer ch.Unlock()
	ch.Lock()
	for _, f := range ch.files {
		if f.name == fname {
			return f
		}
	}
	return nil
}

func (ch *tt_channel) File_remove(fname string) bool {
	file := ch.File_find_name(fname)
	if file == nil {
		return false
	}
	defer ch.Unlock()
	ch.Lock()
	delete(ch.files, file.id)
	return true
}

func (ch *tt_channel) File_size(f *tt_file) int {
	defer ch.Unlock()
	ch.Lock()
	return f.size
}

func (ch *tt_channel) File_id(f *tt_file) int {
	defer ch.Unlock()
	ch.Lock()
	return f.id
}

func (ch *tt_channel) File_name(f *tt_file) string {
	defer ch.Unlock()
	ch.Lock()
	return f.name
}

func (ch *tt_channel) File_owner(f *tt_file) string {
	defer ch.Unlock()
	ch.Lock()
	return f.owner
}

func (ch *tt_channel) User_exists(uid int) bool {
	defer ch.Unlock()
	ch.Lock()
	_, exists := ch.users[uid]
	return exists
}

func (ch *tt_channel) User_add(usr *tt_user) bool {
	uid := usr.Uid_read()
	if ch.User_exists(uid) {
		return false
	}
	defer ch.Unlock()
	ch.Lock()
	ch.users[uid] = usr
	usr.Channel_set(ch)
	return true
}

func (ch *tt_channel) Users_read() []*tt_user {
	defer ch.Unlock()
	ch.Lock()
	users := []*tt_user{}
	ids := []int{}
	for id := range ch.users {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		users = append(users, ch.users[id])
	}
	return users
}

func (ch *tt_channel) User_remove(usr *tt_user) bool {
	uid := usr.Uid_read()
	if !ch.User_exists(uid) {
		return false
	}
	defer ch.Unlock()
	ch.Lock()
	delete(ch.users, uid)
	usr.Channel_clear()
	return true
}
