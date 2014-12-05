package sproxy

import (
	"net"
)

type ProxyListener struct {
	ln	net.Listener
}

type ProxyConn struct {
	ip		string		/* dest ip */
	port	string		/* dest port */
	conn	net.Conn
}

func initListenConn(ln net.Listener) *ProxyListener {
	pl := &ProxyListener{}
	pl.ln = ln

	return pl
}

func initProxyConn(c net.Conn) *ProxyConn {
	pc := &ProxyConn{}
	pc.conn = c

	return pc
}

func (pc *ProxyConn) CloseConn() {
	pc.conn.Close()
}

func (pc *ProxyConn) ReadData(b []byte) (int, error) {
	return pc.conn.Read(b)
}

func (pc *ProxyConn) WriteData(b []byte) (int, error) {
	return pc.conn.Write(b)
}
