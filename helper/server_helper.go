/**
 * Created by GoLand.
 * User: link1st
 * Date: 2019-07-25
 * Time: 17:27
 */

package helper

import (
	"net"
	"os"
)

// 获取服务器Ip
func GetServerIp() (ip string) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && IsPublicIP(ipNet.IP) {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
			}
		}
	}

	return
}

//判断是否公网IP
//tcp/ip协议中，专门保留了三个IP地址区域作为私有地址，其地址范围如下：
//10.0.0.0/8：10.0.0.0～10.255.255.255
//172.16.0.0/12：172.16.0.0～172.31.255.255
//192.168.0.0/16：192.168.0.0～192.168.255.255

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func PathGuarantee(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
