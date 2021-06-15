package errors

import (
	"errors"
	"strconv"
)

type ErrCode int

const (
	Succ            ErrCode = 0
	Requesting      ErrCode = 1
	ReqinfoInvalid  ErrCode = 2
	AddressInvalid  ErrCode = 3
	AddressingFail  ErrCode = 4
	NetworkInvalid  ErrCode = 5
	ResolveFail     ErrCode = 6
	DialFail        ErrCode = 7
	MarshalFail     ErrCode = 8
	SendFail        ErrCode = 9
	SendTimeout     ErrCode = 10
	RecvFail        ErrCode = 11
	RspPkgTooBig    ErrCode = 12
	RecvTimeout     ErrCode = 13
	CheckFail       ErrCode = 14
	UnmarshalFail   ErrCode = 15
	RequestPanic    ErrCode = 16
	ContextCanceled ErrCode = 17
	ContextTimeout  ErrCode = 18
	Unknown         ErrCode = 19
	// 生成mqclient冲突
	LuaErr ErrCode = 110001
	// 这里扩展错误码 并在下面对应补充中文释义
)

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

func (ec ErrCode) StringCode() string {
	return strconv.Itoa(int(ec))
}

func (ec ErrCode) Int32() int32 {
	return int32(ec)
}

func (ec ErrCode) Int() int {
	return int(ec)
}

func (ec ErrCode) Error() error {
	if ec == Succ {
		return nil
	}
	return errors.New(ec.String())
}
