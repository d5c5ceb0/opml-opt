package mips

import (
	"opml-opt/callback"
	"opml-opt/common"
	"opml-opt/mips/vm"
	"sync/atomic"
)

var MipsWork *Worker

type Worker struct {
	ModelName string
	ModelPath string
	JobsNum   atomic.Int32
	MaxJobs   int32
}

func InitWorker(modelName string, modelPath string, programPath string) error {
	MipsWork = &Worker{
		ModelName: modelName,
		ModelPath: modelPath,
		JobsNum:   atomic.Int32{},
		MaxJobs:   1,
	}
	vm.ModelPath = modelPath
	vm.MIPS_PROGRAM = programPath
	return nil
}

func Inference(qa common.OptQA) error {
	defer func() {
		if qa.StateRoot == "" && qa.Err == nil {
			qa.Err = common.ErrJobDownUnknow
		}
		callback.DoneWork(qa)
	}()
	jobsNum := MipsWork.JobsNum.Load()
	if jobsNum > MipsWork.MaxJobs {
		qa.Err = common.ErrExceedMaxJobs
		return common.ErrExceedMaxJobs
	}
	rootHash, err := vm.RunCheckPointZeroRoot(qa.Prompt)
	if err != nil {
		qa.Err = err
		return err
	}
	qa.StateRoot = rootHash.String()
	return nil
}
