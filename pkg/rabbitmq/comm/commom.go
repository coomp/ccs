package comm

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coomp/ccs/comm"
)

// GetLocalIpString 获取本地ip
func GetLocalIpString() string {
	b := comm.GetLocalIpByte()
	return net.IP(b).String()
}

// GenerateMQID 做一个id,用于心跳检测
func GenerateMQID(groupName string) string {
	timeStamp := time.Now().UTC()
	var buffer strings.Builder
	buffer.WriteString(groupName)
	buffer.WriteByte('_')
	buffer.WriteString(GetLocalIpString())
	buffer.WriteByte('-')
	buffer.WriteString(strconv.FormatInt(int64(os.Getpid()), 10))
	buffer.WriteByte('-')
	buffer.WriteString(strconv.FormatInt(timeStamp.Unix(), 10))
	buffer.WriteByte('-')
	buffer.WriteString(strconv.FormatInt(int64(timeStamp.Nanosecond()), 10))
	return buffer.String()
}
