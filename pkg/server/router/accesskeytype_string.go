// Code generated by "stringer -type=AccessKeyType"; DO NOT EDIT.

package router

import "strconv"

const _AccessKeyType_name = "NoAccessKeyClientAccessKeyMasterAccessKey"

var _AccessKeyType_index = [...]uint8{0, 11, 26, 41}

func (i AccessKeyType) String() string {
	if i < 0 || i >= AccessKeyType(len(_AccessKeyType_index)-1) {
		return "AccessKeyType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AccessKeyType_name[_AccessKeyType_index[i]:_AccessKeyType_index[i+1]]
}
