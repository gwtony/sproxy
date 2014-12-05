package sproxy

import (
	goconf "github.com/msbranco/goconfig"
	"path/filepath"
	"os/user"
	"time"
)

const DEFAULT_CONF_PATH = "../conf"
const DEFAULT_CONF		= "sproxy.conf"

type ProxyConfig struct {
	log		string		/* log file */
	level	string		/* log level */
	timeout time.Duration
	port	string		/* listen port */
	dport	string		/* dest port */
	key		string		/* private key path */
}

func (conf *ProxyConfig) ReadConf(file string) (*ProxyConfig, error) {
	if file == "" {
		file = filepath.Join(DEFAULT_CONF_PATH, DEFAULT_CONF)
	}

	c, err := goconf.ReadConfigFile(file)
	if err != nil {
		return nil, err
	}

	//TODO: check
	conf.log, _			= c.GetString("default", "log")
	conf.level, _		= c.GetString("default", "level")
	timeout, _			:= c.GetInt64("default", "timeout")
	conf.timeout		= time.Duration(timeout) * time.Millisecond
	conf.port, _		= c.GetString("default", "port")
	conf.dport, _		= c.GetString("default", "dport")

	conf.key, err	= c.GetString("default", "private_key")
	if err != nil {
		/* default private key path is $HOME/.ssh/id_rsa */
        user, err := user.Current()
        if err != nil {
			return nil, err
		}
		conf.key = filepath.Join(user.HomeDir, ".ssh", "id_rsa")
	}

	return conf, nil
}

