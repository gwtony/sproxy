package sproxy

import (
	"net"
	"io"
	"encoding/json"
	"errors"
)

const (
	BUF_SIZE	= 4096
	CHAN_SIZE	= 10240
)

type Map map[string] string

type ProxyRequest struct {
	num		int			/* ip num */
	buf		[]byte		/* receive buffer */
	end		int			/* buffer offset */
	cmd		string
	ip		[]string
	result	Map			/* result of cmd */
	conn	*ProxyConn	/* client connection */
	cache	*ConnDict	/* connection cache */
	log		*Log
	port	string		/* dest port */
	ch		chan bool	/* sync */
	sc		*SshContext
}

type RequestData struct {
    Cmd string
    Ip  []string
}

type ResponseData struct {
	Result *Map
}

func initRequest(sc *SshContext, c *net.Conn, cache *ConnDict, port string, log *Log) (*ProxyRequest, error) {
	pr := &ProxyRequest{}

	pr.sc		= sc
	pr.conn		= initProxyConn(*c)
	pr.cache	= cache
	pr.port		= port
	pr.log		= log

	pr.buf		= make([]byte, BUF_SIZE)
	pr.ch		= make(chan bool, CHAN_SIZE)

	return pr, nil
}

func (pr *ProxyRequest) closeRequest() error {
	pr.conn.CloseConn()
	pr.log.Debug("Close request")

	return nil
}

func (pr *ProxyRequest) readData() error {
	conn := pr.conn
	log := pr.log
	buf := pr.buf
	end := pr.end

	dlen, err := conn.ReadData(buf[end:])
	if err == io.EOF {
		log.Debug("Read data to end")
	} else if err != nil {
		log.Error("Read data error: %s", err.Error())

		return err
	}

	if dlen == 0 {
		log.Debug("Client closed")
		pr.closeRequest()

		return errors.New("client close while reading request")
	}
	pr.end += dlen

	return nil
}

func (pr *ProxyRequest) parseData() (int, error) {
	log := pr.log

	rd, err := parseRequest(string(pr.buf[:pr.end]))
	if err != nil {
		pr.log.Error("Parse request data failed")
		return ERROR, err
	}

	pr.cmd = rd.Cmd
	pr.ip  = rd.Ip
	pr.num = len(pr.ip)
	pr.result = make(map[string] string, pr.num)

	log.Debug("Parse done, request is cmd:%s, ip:%s", pr.cmd, pr.ip)

	return OK, nil
}

func (pr *ProxyRequest) executeCmd() error {
	for _, ip := range pr.ip {
		go pr.sshCore(ip)
	}

	pr.waitSignal(pr.num)

	return nil
}

func (pr *ProxyRequest) signal() {
	pr.ch <- true
}

func (pr *ProxyRequest) waitSignal(cycle int) {
	for i := 0; i < cycle; i++ {
		<-pr.ch
	}
}

func (pr *ProxyRequest) sshCore(ip string) error {
	sc		:= pr.sc
	addr	:= ip + ":" + pr.port
	log		:= pr.log
	cmd		:= pr.cmd
	update	:= false	/* need to update cache */
	clean	:= false	/* need to clean cache */

	sconn, err := pr.lookUpCache(addr)

	defer pr.signal()	/* sync between ssh core and execute cmd */

	if sconn == nil {
		log.Debug("Not found %s in dict", addr)
		update = true
		sconn, err = initSshConn(sc, addr, log)
		if err != nil {
			log.Error("Ssh connect to %s failed", addr)
			pr.result[ip] = err.Error()

			return err
		}
	}

	sconn.sshLock()

	result, err := sconn.sshExec(cmd)

	sconn.sshUnlock()

	if err != nil {
		log.Error("Execute %s in %s failed, error: %s", cmd, addr, err.Error())
		clean = true
	}

	if clean == true {
		sconn.sshClose()
		_ = pr.updateCache(addr, nil)
		pr.result[ip] = err.Error()

		return err
	} else {
		if update == true {
			err := pr.updateCache(addr, sconn)
			if err != nil {
				log.Error("Update cache for %s failed", addr)

				pr.result[ip] = err.Error()
			}
		}
	}

	pr.result[ip] = string(result)

	log.Debug("address %s result is %s", ip, string(result))

	return nil
}

func (pr *ProxyRequest) lookUpCache(addr string) (*SshConn, error) {
	return  pr.cache.lookUpDict(addr)
}

func (pr *ProxyRequest) updateCache(addr string, sconn *SshConn) error {
	return pr.cache.updateDict(addr, sconn)
}

func (pr *ProxyRequest) sendResponse() error {
	log := pr.log

	rd := &ResponseData{}
	rd.Result = &pr.result

	b, err := json.Marshal(rd)
	if err != nil {
		log.Error("Generate response data failed")
		return err
	}

	log.Info("Cmd %s to %s result is %s", pr.cmd, pr.ip, string(b))

	pr.conn.WriteData(b)

	return nil
}


