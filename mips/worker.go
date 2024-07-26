package mips

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"opml-opt/callback"
	"opml-opt/common"
	"opml-opt/log"
	"opml-opt/mips/vm"
	"os"
	"os/exec"
	"strings"
	"sync"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

var ConfigPath string

var MipsWork *Worker

type Worker struct {
	ModelName string
	ModelPath string
	JobsNum   int32
	MaxJobs   int32
	mut       sync.Mutex
}

func InitWorker(modelName string, modelPath string, programPath string) error {
	MipsWork = &Worker{
		ModelName: modelName,
		ModelPath: modelPath,
		JobsNum:   0,
		MaxJobs:   1,
	}
	vm.ModelPath = modelPath
	vm.MIPS_PROGRAM = programPath
	return nil
}

func Status() int {
	jobsNum := MipsWork.JobsNum
	if jobsNum > MipsWork.MaxJobs {
		return 1
	} else {
		return 0
	}
}

func Inference(qa common.OptQA) error {
	defer func() {
		if qa.StateRoot == "" && qa.Err == nil {
			qa.Err = common.ErrJobDownUnknow
		}
		callback.DoneWork(qa)
	}()

	MipsWork.mut.Lock()
	if MipsWork.JobsNum >= MipsWork.MaxJobs {
		log.Info("mips jobs exceed")
		return common.ErrExceedMaxJobs
	}
	MipsWork.JobsNum++
	MipsWork.mut.Unlock()
	defer func() {
		MipsWork.JobsNum -= 1
	}()

	log.Debugf("mips worker handling %v", qa)
	cmd := exec.Command(os.Args[0], "mips", "--config", ConfigPath, "--prompt", qa.Prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err)
	}
	log.Info(string(output))
	if cmd.ProcessState.Success() {
		scanner := bufio.NewScanner(bytes.NewBuffer(output))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "ok:") {
				hexStr := strings.TrimLeft(strings.Trim(line, "\n"), "ok: 0x")
				qa.StateRoot = ethCommon.HexToHash(hexStr).String()
				return nil
			}
		}
		if err := scanner.Err(); err != nil {
			qa.Err = fmt.Errorf("error reading child process output:%v", err)
			return fmt.Errorf("error reading child process output:%v", err)
		}
	}
	qa.Err = errors.New("mips run failed")
	return errors.New("mips run failed")
}
