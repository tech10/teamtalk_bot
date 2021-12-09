package main

import (
	"strings"
)

// Channel flags.
const TT_CHANNEL_DEFAULT = 0x0000

const (
	TT_CHANNEL_DEFAULT_STR             = "default"
	TT_CHANNEL_PERMANENT               = 0x0001
	TT_CHANNEL_PERMANENT_STR           = "permanent"
	TT_CHANNEL_SOLO_TRANSMIT           = 0x0002
	TT_CHANNEL_SOLO_TRANSMIT_STR       = "solo transmit"
	TT_CHANNEL_CLASSROOM               = 0x0004
	TT_CHANNEL_CLASSROOM_STR           = "classroom"
	TT_CHANNEL_OPERATOR_RECV_ONLY      = 0x0008
	TT_CHANNEL_OPERATOR_RECV_ONLY_STR  = "operator receive only"
	TT_CHANNEL_NO_VOICE_ACTIVATION     = 0x0010
	TT_CHANNEL_NO_VOICE_ACTIVATION_STR = "no voice activation"
	TT_CHANNEL_NO_RECORDING            = 0x0020
	TT_CHANNEL_NO_RECORDING_STR        = "no recording"
)

// User right flags.
const TT_USERRIGHT_NONE = 0x00000000

const (
	TT_USERRIGHT_NONE_STR                      = "none"
	TT_USERRIGHT_MULTI_LOGIN                   = 0x00000001
	TT_USERRIGHT_MULTI_LOGIN_STR               = "login multiple times"
	TT_USERRIGHT_VIEW_ALL_USERS                = 0x00000002
	TT_USERRIGHT_VIEW_ALL_USERS_STR            = "view all users"
	TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL      = 0x00000004
	TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL_STR  = "create temporary channels"
	TT_USERRIGHT_MODIFY_CHANNELS               = 0x00000008
	TT_USERRIGHT_MODIFY_CHANNELS_STR           = "modify channels"
	TT_USERRIGHT_TEXT_MESSAGE_BROADCAST        = 0x00000010
	TT_USERRIGHT_TEXT_MESSAGE_BROADCAST_STR    = "send broadcast messages"
	TT_USERRIGHT_KICK_USERS                    = 0x00000020
	TT_USERRIGHT_KICK_USERS_STR                = "kick users"
	TT_USERRIGHT_BAN_USERS                     = 0x00000040
	TT_USERRIGHT_BAN_USERS_STR                 = "ban users"
	TT_USERRIGHT_MOVE_USERS                    = 0x00000080
	TT_USERRIGHT_MOVE_USERS_STR                = "move users between channels"
	TT_USERRIGHT_OPERATOR_ENABLE               = 0x00000100
	TT_USERRIGHT_OPERATOR_ENABLE_STR           = "make other users channel operators"
	TT_USERRIGHT_UPLOAD_FILES                  = 0x00000200
	TT_USERRIGHT_UPLOAD_FILES_STR              = "upload files"
	TT_USERRIGHT_DOWNLOAD_FILES                = 0x00000400
	TT_USERRIGHT_DOWNLOAD_FILES_STR            = "download files"
	TT_USERRIGHT_UPDATE_SERVER_PROPERTIES      = 0x00000800
	TT_USERRIGHT_UPDATE_SERVER_PROPERTIES_STR  = "update server properties"
	TT_USERRIGHT_TRANSMIT_VOICE                = 0x00001000
	TT_USERRIGHT_TRANSMIT_VOICE_STR            = "transmit audio"
	TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE        = 0x00002000
	TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE_STR    = "transmit video"
	TT_USERRIGHT_TRANSMIT_DESKTOP              = 0x00004000
	TT_USERRIGHT_TRANSMIT_DESKTOP_STR          = "transmit desktop"
	TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT        = 0x00008000
	TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT_STR    = "transmit desktop input"
	TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO     = 0x00010000
	TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO_STR = "transmit audio media file"
	TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO     = 0x00020000
	TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO_STR = "transmit video media file"
)

// User types
const TT_USERTYPE_NONE = 0x0

const (
	TT_USERTYPE_NONE_STR    = "unauthorized"
	TT_USERTYPE_DEFAULT     = 0x01
	TT_USERTYPE_DEFAULT_STR = "default"
	TT_USERTYPE_ADMIN       = 0x02
	TT_USERTYPE_ADMIN_STR   = "admin"
)

// Subscription types.
const TT_SUBSCRIBE_NONE = 0x00000000

const (
	TT_SUBSCRIBE_NONE_STR                  = "none"
	TT_SUBSCRIBE_USER_MSG                  = 0x00000001
	TT_SUBSCRIBE_USER_MSG_STR              = "private messages"
	TT_SUBSCRIBE_CHANNEL_MSG               = 0x00000002
	TT_SUBSCRIBE_CHANNEL_MSG_STR           = "channel messages"
	TT_SUBSCRIBE_BROADCAST_MSG             = 0x00000004
	TT_SUBSCRIBE_BROADCAST_MSG_STR         = "broadcast messages"
	TT_SUBSCRIBE_CUSTOM_MSG                = 0x00000008
	TT_SUBSCRIBE_CUSTOM_MSG_STR            = "custom private messages"
	TT_SUBSCRIBE_VOICE                     = 0x00000010
	TT_SUBSCRIBE_VOICE_STR                 = "audio"
	TT_SUBSCRIBE_VIDEO_CAPTURE             = 0x00000020
	TT_SUBSCRIBE_VIDEO_CAPTURE_STR         = "video"
	TT_SUBSCRIBE_DESKTOP                   = 0x00000040
	TT_SUBSCRIBE_DESKTOP_STR               = "desktop"
	TT_SUBSCRIBE_DESKTOP_INPUT             = 0x00000080
	TT_SUBSCRIBE_DESKTOP_INPUT_STR         = "desktop input"
	TT_SUBSCRIBE_MEDIA_FILE                = 0x00000100
	TT_SUBSCRIBE_MEDIA_FILE_STR            = "media file stream"
	TT_SUBSCRIBE_INTERCEPT_USER_MSG        = 0x00010000
	TT_SUBSCRIBE_INTERCEPT_USER_MSG_STR    = "intercept private messages"
	TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG     = 0x00020000
	TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG_STR = "intercept channel messages"
)

// const TT_SUBSCRIBE_INTERCEPT_BROADCAST_MSG = 0x00040000
// const TT_SUBSCRIBE_INTERCEPT_BROADCAST_MSG_STR = "intercept broadcast messages"
const TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG = 0x00080000

const (
	TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG_STR    = "intercept custom private messages"
	TT_SUBSCRIBE_INTERCEPT_VOICE             = 0x00100000
	TT_SUBSCRIBE_INTERCEPT_VOICE_STR         = "intercept audio"
	TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE     = 0x00200000
	TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE_STR = "intercept video"
	TT_SUBSCRIBE_INTERCEPT_DESKTOP           = 0x00400000
	TT_SUBSCRIBE_INTERCEPT_DESKTOP_STR       = "intercept desktop"
)

// const TT_SUBSCRIBE_INTERCEPT_DESKTOP_INPUT = 0x00800000
// const TT_SUBSCRIBE_INTERCEPT_DESKTOP_INPUT_STR = "intercept desktop input"
const TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE = 0x01000000
const TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE_STR = "intercept media file stream"

// Message types.
const TT_MSGTYPE_USER = 1

const (
	TT_MSGTYPE_USER_STR      = "private"
	TT_MSGTYPE_CHANNEL       = 2
	TT_MSGTYPE_CHANNEL_STR   = "channel"
	TT_MSGTYPE_BROADCAST     = 3
	TT_MSGTYPE_BROADCAST_STR = "broadcast"
	TT_MSGTYPE_CUSTOM        = 4
	TT_MSGTYPE_CUSTOM_STR    = "custom private"
)

// TeamTalk user status flags

const (
	TT_USERSTATUS_NONE         = 0x00000000
	TT_USERSTATUS_NONE_STR     = "online"
	TT_USERSTATUS_AWAY         = 0x00000001
	TT_USERSTATUS_AWAY_STR     = "away"
	TT_USERSTATUS_QUESTION     = 0x00000002
	TT_USERSTATUS_QUESTION_STR = "question"
)

func teamtalk_flags_fmt_str(str string) string {
	return strings.TrimSuffix(str, ", ")
}

func teamtalk_flags_read(flags, flag int) bool {
	if flags == 0 && flag == 0 {
		return true
	}
	if flags&flag != 0 {
		return true
	}
	return false
}

func teamtalk_flags_set(flags, flag int) int {
	if !teamtalk_flags_read(flags, flag) {
		flags |= flag
	}
	return flags
}

func teamtalk_flags_unset(flags, flag int) int {
	flags &^= flag
	return flags
}

func teamtalk_flags_toggle(flags, flag int) int {
	if !teamtalk_flags_read(flags, flag) {
		return teamtalk_flags_set(flags, flag)
	}
	return teamtalk_flags_unset(flags, flag)
}

func teamtalk_flags_subscriptions_str(flags int) string {
	str := ""
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_USER_MSG) {
		str += TT_SUBSCRIBE_USER_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_CHANNEL_MSG) {
		str += TT_SUBSCRIBE_CHANNEL_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_BROADCAST_MSG) {
		str += TT_SUBSCRIBE_BROADCAST_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_CUSTOM_MSG) {
		str += TT_SUBSCRIBE_CUSTOM_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_VOICE) {
		str += TT_SUBSCRIBE_VOICE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_VIDEO_CAPTURE) {
		str += TT_SUBSCRIBE_VIDEO_CAPTURE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_DESKTOP) {
		str += TT_SUBSCRIBE_DESKTOP_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_DESKTOP_INPUT) {
		str += TT_SUBSCRIBE_DESKTOP_INPUT_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_MEDIA_FILE) {
		str += TT_SUBSCRIBE_MEDIA_FILE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_USER_MSG) {
		str += TT_SUBSCRIBE_INTERCEPT_USER_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG) {
		str += TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG) {
		str += TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_VOICE) {
		str += TT_SUBSCRIBE_INTERCEPT_VOICE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE) {
		str += TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_DESKTOP) {
		str += TT_SUBSCRIBE_INTERCEPT_DESKTOP_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE) {
		str += TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE_STR + ", "
	}
	return teamtalk_flags_fmt_str(str)
}

func teamtalk_flags_channel_options_str(flags int) string {
	str := ""
	if teamtalk_flags_read(flags, TT_CHANNEL_DEFAULT) {
		str += TT_CHANNEL_DEFAULT_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_CHANNEL_PERMANENT) {
		str += TT_CHANNEL_PERMANENT_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_CHANNEL_SOLO_TRANSMIT) {
		str += TT_CHANNEL_SOLO_TRANSMIT_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_CHANNEL_CLASSROOM) {
		str += TT_CHANNEL_CLASSROOM_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_CHANNEL_OPERATOR_RECV_ONLY) {
		str += TT_CHANNEL_OPERATOR_RECV_ONLY_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_CHANNEL_NO_VOICE_ACTIVATION) {
		str += TT_CHANNEL_NO_VOICE_ACTIVATION_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_CHANNEL_NO_RECORDING) {
		str += TT_CHANNEL_NO_RECORDING_STR + ", "
	}
	return teamtalk_flags_fmt_str(str)
}

func teamtalk_flags_userrights_str(flags int) string {
	str := ""
	if teamtalk_flags_read(flags, TT_USERRIGHT_NONE) {
		str += TT_USERRIGHT_NONE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_MULTI_LOGIN) {
		str += TT_USERRIGHT_MULTI_LOGIN_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_VIEW_ALL_USERS) {
		str += TT_USERRIGHT_VIEW_ALL_USERS_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL) {
		str += TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_MODIFY_CHANNELS) {
		str += TT_USERRIGHT_MODIFY_CHANNELS_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TEXT_MESSAGE_BROADCAST) {
		str += TT_USERRIGHT_TEXT_MESSAGE_BROADCAST_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_KICK_USERS) {
		str += TT_USERRIGHT_KICK_USERS_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_BAN_USERS) {
		str += TT_USERRIGHT_BAN_USERS_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_MOVE_USERS) {
		str += TT_USERRIGHT_MOVE_USERS_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_OPERATOR_ENABLE) {
		str += TT_USERRIGHT_OPERATOR_ENABLE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_UPLOAD_FILES) {
		str += TT_USERRIGHT_UPLOAD_FILES_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_DOWNLOAD_FILES) {
		str += TT_USERRIGHT_DOWNLOAD_FILES_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_UPDATE_SERVER_PROPERTIES) {
		str += TT_USERRIGHT_UPDATE_SERVER_PROPERTIES_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TRANSMIT_VOICE) {
		str += TT_USERRIGHT_TRANSMIT_VOICE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE) {
		str += TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TRANSMIT_DESKTOP) {
		str += TT_USERRIGHT_TRANSMIT_DESKTOP_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT) {
		str += TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO) {
		str += TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO) {
		str += TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO_STR + ", "
	}
	return teamtalk_flags_fmt_str(str)
}

func teamtalk_flags_usertype_str(utype int) string {
	switch utype {
	case TT_USERTYPE_NONE:
		return TT_USERTYPE_NONE_STR
	case TT_USERTYPE_DEFAULT:
		return TT_USERTYPE_DEFAULT_STR
	case TT_USERTYPE_ADMIN:
		return TT_USERTYPE_ADMIN_STR
	}
	return "unknown"
}

func teamtalk_flags_status_mode_str(flags int) string {
	str := ""
	if teamtalk_flags_read(flags, TT_USERSTATUS_NONE) {
		return TT_USERSTATUS_NONE_STR
	}
	if teamtalk_flags_read(flags, TT_USERSTATUS_AWAY) {
		str += TT_USERSTATUS_AWAY_STR + ", "
	}
	if teamtalk_flags_read(flags, TT_USERSTATUS_QUESTION) {
		str += TT_USERSTATUS_QUESTION_STR + ", "
	}
	if !teamtalk_flags_read(flags, TT_USERSTATUS_AWAY) && !teamtalk_flags_read(flags, TT_USERSTATUS_QUESTION) {
		return TT_USERSTATUS_NONE_STR
	}
	return teamtalk_flags_fmt_str(str)
}

func teamtalk_flags_message_type_str(flag int) string {
	str := ""
	switch flag {
	case TT_MSGTYPE_USER:
		str = TT_MSGTYPE_USER_STR
	case TT_MSGTYPE_CHANNEL:
		str = TT_MSGTYPE_CHANNEL_STR
	case TT_MSGTYPE_BROADCAST:
		str = TT_MSGTYPE_BROADCAST_STR
	case TT_MSGTYPE_CUSTOM:
		str = TT_MSGTYPE_CUSTOM_STR
	}
	return str
}

func teamtalk_flags_menu_item(flags, flag int, flag_str string) string {
	item := flag_str + " ("
	if teamtalk_flags_read(flags, flag) {
		item += "enabled"
	} else {
		item += "disabled"
	}
	return item + ")"
}

func teamtalk_flags_subscriptions_menu(flags int) (int, bool) {
	console_write("Anything disabled in the subscriptions menu, is or will be unsubscribed from, and anything enabled, is or will be subscribed to.")
	for {
		menu := []string{}
		menu_flags := []int{}

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_USER_MSG, TT_SUBSCRIBE_USER_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_USER_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_CHANNEL_MSG, TT_SUBSCRIBE_CHANNEL_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_CHANNEL_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_BROADCAST_MSG, TT_SUBSCRIBE_BROADCAST_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_BROADCAST_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_CUSTOM_MSG, TT_SUBSCRIBE_CUSTOM_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_CUSTOM_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_VOICE, TT_SUBSCRIBE_VOICE_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_VOICE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_VIDEO_CAPTURE, TT_SUBSCRIBE_VIDEO_CAPTURE_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_VIDEO_CAPTURE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_DESKTOP, TT_SUBSCRIBE_DESKTOP_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_DESKTOP)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_DESKTOP_INPUT, TT_SUBSCRIBE_DESKTOP_INPUT_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_DESKTOP_INPUT)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_MEDIA_FILE, TT_SUBSCRIBE_MEDIA_FILE_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_MEDIA_FILE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_USER_MSG, TT_SUBSCRIBE_INTERCEPT_USER_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_USER_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG, TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_CHANNEL_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG, TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_CUSTOM_MSG)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_VOICE, TT_SUBSCRIBE_INTERCEPT_VOICE_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_VOICE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE, TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_VIDEO_CAPTURE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_DESKTOP, TT_SUBSCRIBE_INTERCEPT_DESKTOP_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_DESKTOP)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE, TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE_STR))
		menu_flags = append(menu_flags, TT_SUBSCRIBE_INTERCEPT_MEDIA_FILE)

		menu = append(menu, "done")

		res, aborted := console_read_menu("Please select what you want to enable or disable.\r\n", menu)
		if aborted || res == -1 {
			return flags, true
		}
		if menu[res] == "done" {
			break
		}
		flags = teamtalk_flags_toggle(flags, menu_flags[res])
	}
	return flags, false
}

func teamtalk_flags_userrights_menu(flags int) (int, bool) {
	for {
		menu := []string{}
		menu_flags := []int{}

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_MULTI_LOGIN, TT_USERRIGHT_MULTI_LOGIN_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_MULTI_LOGIN)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_VIEW_ALL_USERS, TT_USERRIGHT_VIEW_ALL_USERS_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_VIEW_ALL_USERS)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL, TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_CREATE_TEMPORARY_CHANNEL)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_MODIFY_CHANNELS, TT_USERRIGHT_MODIFY_CHANNELS_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_MODIFY_CHANNELS)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TEXT_MESSAGE_BROADCAST, TT_USERRIGHT_TEXT_MESSAGE_BROADCAST_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TEXT_MESSAGE_BROADCAST)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_KICK_USERS, TT_USERRIGHT_KICK_USERS_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_KICK_USERS)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_BAN_USERS, TT_USERRIGHT_BAN_USERS_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_BAN_USERS)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_MOVE_USERS, TT_USERRIGHT_MOVE_USERS_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_MOVE_USERS)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_OPERATOR_ENABLE, TT_USERRIGHT_OPERATOR_ENABLE_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_OPERATOR_ENABLE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_UPLOAD_FILES, TT_USERRIGHT_UPLOAD_FILES_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_UPLOAD_FILES)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_DOWNLOAD_FILES, TT_USERRIGHT_DOWNLOAD_FILES_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_DOWNLOAD_FILES)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_UPDATE_SERVER_PROPERTIES, TT_USERRIGHT_UPDATE_SERVER_PROPERTIES_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_UPDATE_SERVER_PROPERTIES)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TRANSMIT_VOICE, TT_USERRIGHT_TRANSMIT_VOICE_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TRANSMIT_VOICE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE, TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TRANSMIT_VIDEO_CAPTURE)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TRANSMIT_DESKTOP, TT_USERRIGHT_TRANSMIT_DESKTOP_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TRANSMIT_DESKTOP)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT, TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TRANSMIT_DESKTOP_INPUT)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_AUDIO)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO_STR))
		menu_flags = append(menu_flags, TT_USERRIGHT_TRANSMIT_MEDIA_FILE_VIDEO)

		menu = append(menu, "done")

		res, aborted := console_read_menu("Please select what you want to enable or disable.\r\n", menu)
		if aborted || res == -1 {
			return flags, true
		}
		if menu[res] == "done" {
			break
		}
		flags = teamtalk_flags_toggle(flags, menu_flags[res])
	}
	return flags, false
}

func teamtalk_flags_channel_options_menu(flags int) (int, bool) {
	for {
		menu := []string{}
		menu_flags := []int{}

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_CHANNEL_PERMANENT, TT_CHANNEL_PERMANENT_STR))
		menu_flags = append(menu_flags, TT_CHANNEL_PERMANENT)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_CHANNEL_SOLO_TRANSMIT, TT_CHANNEL_SOLO_TRANSMIT_STR))
		menu_flags = append(menu_flags, TT_CHANNEL_SOLO_TRANSMIT)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_CHANNEL_CLASSROOM, TT_CHANNEL_CLASSROOM_STR))
		menu_flags = append(menu_flags, TT_CHANNEL_CLASSROOM)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_CHANNEL_OPERATOR_RECV_ONLY, TT_CHANNEL_OPERATOR_RECV_ONLY_STR))
		menu_flags = append(menu_flags, TT_CHANNEL_OPERATOR_RECV_ONLY)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_CHANNEL_NO_VOICE_ACTIVATION, TT_CHANNEL_NO_VOICE_ACTIVATION_STR))
		menu_flags = append(menu_flags, TT_CHANNEL_NO_VOICE_ACTIVATION)

		menu = append(menu, teamtalk_flags_menu_item(flags, TT_CHANNEL_NO_RECORDING, TT_CHANNEL_NO_RECORDING_STR))
		menu_flags = append(menu_flags, TT_CHANNEL_NO_RECORDING)

		menu = append(menu, "done")

		res, aborted := console_read_menu("Please select what you want to enable or disable.\r\n", menu)
		if aborted || res == -1 {
			return flags, true
		}
		if menu[res] == "done" {
			break
		}
		flags = teamtalk_flags_toggle(flags, menu_flags[res])
	}
	return flags, false
}

func teamtalk_flags_usertype_menu() (int, bool) {
	menu := []string{}
	menu_flags := []int{}

	menu = append(menu, TT_USERTYPE_DEFAULT_STR)
	menu_flags = append(menu_flags, TT_USERTYPE_DEFAULT)

	menu = append(menu, TT_USERTYPE_ADMIN_STR)
	menu_flags = append(menu_flags, TT_USERTYPE_ADMIN)

	res, aborted := console_read_menu("Please select a user type.\r\n", menu)
	if aborted || res == -1 {
		return -1, true
	}
	return menu_flags[res], false
}
