package errors

import (
	"errors"
	"strconv"
)

type ErrCode int

const (
	Succ ErrCode = 0
	// 生成mqclient冲突
	LuaErr ErrCode = 110001
	// 这里扩展错误码 并在下面对应补充中文释义
)

func (ec ErrCode) String() string {
	switch ec {
	case Succ:
		return "成功"
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

func (ec ErrCode) Error() error {
	if ec == Succ {
		return nil
	}
	return errors.New(ec.String())
}