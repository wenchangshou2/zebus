package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
)

func getCpuId() (string, error) {
	var (
		stdout io.ReadCloser
		err    error
	)
	cmd := exec.Command("cmd", "/C", "wmic", "cpu", "get", "processorid")
	if stdout, err = cmd.StdoutPipe(); err != nil {
		log.Fatal(err)
		return "", err
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		return "", err
	}
	if opBytes, err := ioutil.ReadAll(stdout); err != nil {
		return "", err
	} else {
		str := string(opBytes)
		strArr := strings.Split(str, "\n")
		return strArr[1], nil
	}
}
func getMachineUUID() (string, error) {
	return "", nil
}
func GetSystemUUID() (string, error) {
	var (
		cpuId     string
		machineId string
		str       string
		err       error
	)
	cpuId, err = getCpuId()
	if err != nil {
		return "", errors.New("获取cpuid失败")
	}
	machineId, err = getMachineUUID()
	if err != nil {
		return "", errors.New("获取machine id制造")
	}
	str = cpuId + machineId
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
