package llamago

import (
	"context"
	"fmt"
	"opml-opt/callback"
	"opml-opt/common"
	"opml-opt/log"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gotzmann/llama.go/pkg/llama"
)

var LlamaWorker *Worker

type Worker struct {
	Params    *llama.ModelParams
	ModelName string
	ModelPath string
	JobsNum   int32
	MaxJobs   int32
	mut       sync.Mutex
}

type Options struct {
	Prompt  string  `long:"prompt" description:"Text prompt from user to feed the model input"`
	Model   string  `long:"model" description:"Path and file name of converted .bin LLaMA model [ llama-7b-fp32.bin, etc ]"`
	Server  bool    `long:"server" description:"Start in Server Mode acting as REST API endpoint"`
	Host    string  `long:"host" description:"Host to allow requests from in Server Mode [ localhost by default ]"`
	Port    string  `long:"port" description:"Port listen to in Server Mode [ 8080 by default ]"`
	Pods    int64   `long:"pods" description:"Maximum pods or units of parallel execution allowed in Server Mode [ 1 by default ]"`
	Threads int     `long:"threads" description:"Max number of CPU cores you allow to use for one pod [ all cores by default ]"`
	Context uint32  `long:"context" description:"Context size in tokens [ 1024 by default ]"`
	Predict uint32  `long:"predict" description:"Number of tokens to predict [ 512 by default ]"`
	Temp    float32 `long:"temp" description:"Model temperature hyper parameter [ 0.50 by default ]"`
	Silent  bool    `long:"silent" description:"Hide welcome logo and other output [ shown by default ]"`
	Chat    bool    `long:"chat" description:"Chat with user in interactive mode instead of compute over static prompt"`
	Dir     string  `long:"dir" description:"Directory used to download .bin model specified with --model parameter [ current by default ]"`
	Profile bool    `long:"profile" description:"Profe CPU performance while running and store results to cpu.pprof file"`
	UseAVX  bool    `long:"avx" description:"Enable x64 AVX2 optimizations for Intel and AMD machines"`
	UseNEON bool    `long:"neon" description:"Enable ARM NEON optimizations for Apple and ARM machines"`
}

func defaultOpts() Options {
	return Options{
		Prompt:  "",
		Model:   "",
		Server:  false,
		Host:    "localhost",
		Port:    "8080",
		Pods:    1,
		Threads: runtime.NumCPU(),
		Context: 1024,
		Predict: 64,
		Temp:    0,
		Silent:  false,
		Chat:    false,
		Dir:     "",
		Profile: false,
		UseAVX:  false,
		UseNEON: false,
	}
}

func Status() int {
	jobsNum := LlamaWorker.JobsNum
	if jobsNum > LlamaWorker.MaxJobs {
		return 1
	} else {
		return 0
	}
}

func InitWorker(modelName string, modelPath string) error {
	opts := defaultOpts()
	opts.Model = modelPath
	params := &llama.ModelParams{
		Model:         opts.Model,
		MaxThreads:    opts.Threads,
		UseAVX:        opts.UseAVX,
		UseNEON:       opts.UseNEON,
		Interactive:   opts.Chat,
		CtxSize:       opts.Context,
		Seed:          -1,
		PredictCount:  opts.Predict,
		RepeatLastN:   opts.Context, // TODO: Research on best value
		PartsCount:    -1,
		BatchSize:     opts.Context, // TODO: What's the better size?
		TopK:          40,
		TopP:          0.95,
		Temp:          0.5,
		RepeatPenalty: 1.10,
		MemoryFP16:    true,
	}

	LlamaWorker = &Worker{
		Params:    params,
		ModelName: modelName,
		ModelPath: modelPath,
		JobsNum:   0,
		MaxJobs:   1,
	}
	return nil
}

func Inference(qa common.OptQA) error {
	defer func() {
		if qa.Answer == "" && qa.Err == nil {
			qa.Err = common.ErrJobDownUnknow
		}
		callback.DoneWork(qa)
	}()
	LlamaWorker.mut.Lock()
	if LlamaWorker.JobsNum >= LlamaWorker.MaxJobs {
		log.Info("llama go jobs exceed")
		return common.ErrExceedMaxJobs
	}
	LlamaWorker.JobsNum++
	LlamaWorker.mut.Unlock()
	defer func() {
		LlamaWorker.JobsNum -= 1
	}()

	log.Infof("llama go handling job %v", qa)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./llamacpp/llama-cli",
		"-m", "./llama-2-7b-chat.Q2_K.gguf", "-p", qa.Prompt,
		"--temp", "0", "-n", "256")

	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Command timed out: %v", ctx.Err())
	}

	if err != nil {
		fmt.Println("Command execution failed: %v, Output: %s", err, string(output))
	}

	qa.Answer = string(output)

	if cmd.ProcessState.Success() {
		fmt.Println("Command executed successfully.")
	} else {
		fmt.Println("Command execution failed with non-zero exit status.")
	}

	return nil
}
