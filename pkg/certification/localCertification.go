package certification

import (
	"errors"
	"strings"
)

type LocalCertification struct {
	Username string
	Password string
}

func (c *LocalCertification) Login(params map[string]interface{}) (bool, error) {
	var (
		username string
		password string
		isOk     bool
	)
	if username, isOk = params["username"].(string); !isOk {
		return false, errors.New("username必须填写")
	}
	if password, isOk = params["password"].(string); !isOk {
		return false, errors.New("password必须填写")
	}
	if strings.Compare(c.Username, username) == 0 && strings.Compare(c.Password, password) == 0 {
		return true, nil
	}
	return false, nil
}
