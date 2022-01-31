package internet

import (
	"net"
)

type Connection interface {
	net.Conn
}

type StatCouterConnection struct {
	Connection
}

func (c *StatCouterConnection) Read(b []byte) (int, error) {
	nBytes, err := c.Connection.Read(b)
	return nBytes, err
}

func (c *StatCouterConnection) Write(b []byte) (int, error) {
	nBytes, err := c.Connection.Write(b)
	return nBytes, err
}
