package xerr

func MapErrMsg(errCode uint32) string {
	if msg, ok := message[errCode]; ok {
		return msg
	} else {
		return "服务器开小差啦,稍后再来试一试"
	}
}

func IsCodeErr(errCode uint32) bool {
	if _, ok := message[errCode]; ok {
		return true
	} else {
		return false
	}
}
