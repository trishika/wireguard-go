// +build !linux

package wireguard

import (
	"net"
)

func setMark(conn *net.UDPConn, value uint32) error {
	return nil
}
