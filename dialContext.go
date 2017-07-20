package moxxi

import (
	"context"
	"net"
)

func StaticDialContext(address net.IP, port int) (
	func(ctx context.Context, network, ignoredAddress string) (net.Conn, error),
	error) {

	tcpAddr := net.TCPAddr{
		IP:   address,
		Port: port,
	}

	return func(ctx context.Context, network, ignoredAddress string) (net.Conn, error) {
		// conn, err := net.Dial("tcp", net.JoinHostPort(address.String(), strconv.Itoa(port)))
		conn, err := net.DialTCP("tcp", nil, &tcpAddr)
		if err != nil {
			return nil, err
		}

		if d, ok := ctx.Deadline(); ok {
			if err := conn.SetDeadline(d); err != nil {
				return nil, err
			}
		}

		go func() {
			if d := ctx.Done(); d != nil {
				<-d
				conn.Close()
			}
		}()

		return conn, err
	}, nil
}
