package errors

import (
	"errors"
	"strconv"
)

// ErrCode 定义下code
type ErrCode int

const (
	// Succ TODO
	Succ ErrCode = 0
	// Requesting TODO
	Requesting ErrCode = 1
	// ReqinfoInvalid TODO
	ReqinfoInvalid ErrCode = 2
	// AddressInvalid TODO
	AddressInvalid ErrCode = 3
	// AddressingFail TODO
	AddressingFail ErrCode = 4
	// NetworkInvalid TODO
	NetworkInvalid ErrCode = 5
	// ResolveFail TODO
	ResolveFail ErrCode = 6
	// DialFail TODO
	DialFail ErrCode = 7
	// MarshalFail TODO
	MarshalFail ErrCode = 8
	// SendFail TODO
	SendFail ErrCode = 9
	// SendTimeout TODO
	SendTimeout ErrCode = 10
	// RecvFail TODO
	RecvFail ErrCode = 11
	// RspPkgTooBig TODO
	RspPkgTooBig ErrCode = 12
	// RecvTimeout TODO
	RecvTimeout ErrCode = 13
	// CheckFail TODO
	CheckFail ErrCode = 14
	// UnmarshalFail TODO
	UnmarshalFail ErrCode = 15
	// RequestPanic TODO
	RequestPanic ErrCode = 16
	// ContextCanceled TODO
	ContextCanceled ErrCode = 17
	// ContextTimeout TODO
	ContextTimeout ErrCode = 18
	// Unknown TODO
	Unknown ErrCode = 19
	// LuaErr TODO
	// 生成mqclient冲突
	LuaErr ErrCode = 110001
	// 这里扩展错误码 并在下面对应补充中文释义
)

// String TODO
func (ec ErrCode) String() string {
	switch ec {
	case Succ:
		return "成功"
	case Requesting:
		return "请求中"
	case ReqinfoInvalid:
		return "收方非法"
	case AddressInvalid:
		return "ip非法"
	case AddressingFail:
		return "ip错误"
	case NetworkInvalid:
		return "网络非法"
	case ResolveFail:
		return "重新发包失败"
	case DialFail:
		return "链接出错"
	case MarshalFail:
		return "MarshalFail"
	case SendFail:
		return "发包失败"
	case SendTimeout:
		return "发包超时"
	case RecvFail:
		return "收包失败"
	case RspPkgTooBig:
		return "收包过大"
	case RecvTimeout:
		return "收包超时"
	case CheckFail:
		return "参数检查错误"
	case UnmarshalFail:
		return "Unmarshal失败"
	case RequestPanic:
		return "RequestPanic"
	case ContextCanceled:
		return "ContextCanceled"
	case ContextTimeout:
		return "ContextTimeout"
	case Unknown:
		return "未知错误"
	default:
		return "未知错误"
	}
}

// StringCode TODO
func (ec ErrCode) StringCode() string {
	return strconv.Itoa(int(ec))
}

// Int32 TODO
func (ec ErrCode) Int32() int32 {
	return int32(ec)
}

// Int TODO
func (ec ErrCode) Int() int {
	return int(ec)
}

// Error TODO
func (ec ErrCode) Error() error {
	if ec == Succ {
		return nil
	}
	return errors.New(ec.String())
}
