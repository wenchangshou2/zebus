package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/denisbrodbeck/machineid"
)

func getCpuId() (string, error) {
	return "", nil
}
func getMachineUUID() (string, error) {
	return "", nil
}

const appKey = "zoolonlAPP"

func protect(appID, id string) string {
	mac := hmac.New(sha256.New, []byte(id))
	mac.Write([]byte(appID))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
func GetSystemUUID() (string, error) {
	var (
		err error
		id  string
	)
	id, err = machineid.ID()
	if err != nil {
		return "", fmt.Errorf("获取id失败:%v", err)
	}
	id = protect(appKey, id)
	hasher := md5.New()
	hasher.Write([]byte(id))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
