package sproxy

import (
	"golang.org/x/crypto/ssh"
	//"io"
	"io/ioutil"
	"net"
	//"os"
	"os/user"
	//"path/filepath"
	"time"
	"sync"
)

var DefaultTimeout = 30 * time.Second

type SshContext struct {
	username	string
	key			string
	timeout		time.Duration
	config		*ssh.ClientConfig
}

type SshClient struct {
	cli	*ssh.Client
}

type SshConn struct {
	addr	string
	client	*ssh.Client
	lock	sync.Mutex
	conn	*net.Conn
}

//func (sc *SshContext) ConnectWithPassword(host, username, pass string) (*Client, error) {
//	return ConnectWithPasswordTimeout(host, username, pass, DefaultTimeout)
//}
//
//func (sc *SshContext) ConnectWithPasswordTimeout(host, username, pass string, timeout time.Duration) (*Client, error) {
//	authMethod := ssh.Password(pass)
//
//	return connect(username, host, authMethod, timeout)
//}

func initSshContext(sc *SshContext, path string, timeout time.Duration, log *Log) error {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("Read private key from %s failed", path)
		return err
	}
	sc.key = string(key)

	user, err := user.Current()
	if err != nil {
		return err
	}
	sc.username = user.Username

	signer, err := ssh.ParsePrivateKey([]byte(sc.key))
	if err != nil {
		log.Error("Parse private key failed")
		return err
	}

	authMethod := ssh.PublicKeys(signer)

	sc.config = &ssh.ClientConfig{
		User: sc.username,
		Auth: []ssh.AuthMethod{authMethod},
	}

	return nil
}

func initSshConn(sc *SshContext, addr string, log *Log) (*SshConn, error) {
	sconn := &SshConn{}
	sconn.addr = addr
	err := sconn.sshConnect(sc, addr, log)
	if err != nil {
		log.Error("Init ssh connection to %s failed", addr)
		return nil, err
	}

	return sconn, nil
}

func (sconn *SshConn) sshConnect(sc *SshContext, addr string, log *Log) error {
	conn, err := net.DialTimeout("tcp", addr, sc.timeout)
	if err != nil {
		log.Error("Create ssh connection to %s failed", addr)
		return err
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, sc.config)
	if err != nil {
		return err
	}
	client := ssh.NewClient(sshConn, chans, reqs)

	sconn.client = client
	sconn.conn = &conn

	return nil
}

func (sconn *SshConn) sshExec(cmd string) ([]byte, error) {
	session, err := sconn.client.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()

	return session.CombinedOutput(cmd)
}

func (sconn *SshConn) sshLock() {
	sconn.lock.Lock()
}

func (sconn *SshConn) sshUnlock() {
	sconn.lock.Unlock()
}

func (sconn *SshConn) sshClose() {
	(*sconn.conn).Close()
}
