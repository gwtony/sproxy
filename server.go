package sproxy

import (
	"net"
	"time"
)

const (
	OK		= 0
	ERROR	= -1
	DONE	= 2
)

type ProxyServer struct {
	port		string
	dport		string
	timeout		time.Duration
	active		bool
	listener	*ProxyListener
	cache		*ConnDict
	log			*Log
	key			string
	sc			SshContext
}

func InitServer(conf *ProxyConfig, log *Log) (*ProxyServer, error) {
	ps := &ProxyServer{}
	log.Debug("Init server")
	ln, err := net.Listen("tcp", ":" + conf.port)
	if err != nil {
		log.Error("Listen on %s failed", conf.port)
		return nil, err
	}

	ps.listener = initListenConn(ln)
	ps.cache = initConnDict()
	ps.port = conf.port
	ps.timeout = conf.timeout
	ps.log = log
	ps.key = conf.key
	ps.active = true
	ps.dport = conf.dport

	initSshContext(&ps.sc, ps.key, ps.timeout, log)

	return ps, nil
}

func (ps *ProxyServer) CoreRun() error {
	log		:= ps.log
	ln		:= ps.listener.ln
	cache	:= ps.cache
	port	:= ps.dport
	sc		:= ps.sc

	for ; ps.active; {
		conn, err := ln.Accept()
		if err != nil {
			//TODO: handle eagain or normal fail ?
			log.Error("Accept failed: %s", err.Error())
			continue
		}
		go handleRequest(&sc, &conn, cache, port, log)
	}

	return nil
}

func handleRequest(sc *SshContext, conn *net.Conn, cache *ConnDict, port string, log *Log) {
	var ret int

	log.Debug("In handle request")

	req, err := initRequest(sc, conn, cache, port, log)
	if err != nil {
		log.Error("Init request failed: %s", err.Error())
		return
	}

	defer req.closeRequest()

	for {
		err = req.readData()
		if err != nil {
			log.Error("Read data frome client failed: %s", err.Error())
				return
		}
		log.Debug("Begin parse data")
		ret, err = req.parseData()
		if err != nil {
			log.Error("Parse client data failed: %s", err.Error())
			return
		}
		if ret == OK {
			break
		}
	}

	err = req.executeCmd()
	if err != nil {
		log.Error("Execute command failed: %s", err.Error())
		return
	}

	err = req.sendResponse()
	if err != nil {
		log.Error("Send response to client failed: %s", err.Error())
		return
	}
}
