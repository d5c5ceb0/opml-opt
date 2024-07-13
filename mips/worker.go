package mips

import (
	"opml-opt/callback"
	"opml-opt/common"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

var MipsWork *Worker

type Worker struct {
	ModelName string
	ModelPath string
	JobsNum   atomic.Int32
	MaxJobs   int32
}

func InitWorker(modelName string, modelPath string) error {
	MipsWork = &Worker{
		ModelName: modelName,
		ModelPath: modelPath,
		JobsNum:   atomic.Int32{},
		MaxJobs:   1,
	}
	return nil
}

func Inference(qa common.OptQA) error {
	qa.StateRoot = crypto.Keccak256Hash([]byte(time.Now().Format(time.RFC1123))).String()
	callback.DoneWork(qa)
	return nil
}
