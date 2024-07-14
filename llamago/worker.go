package llamago

import (
	"opml-opt/callback"
	"opml-opt/common"
	"opml-opt/log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gotzmann/llama.go/pkg/llama"
	"github.com/gotzmann/llama.go/pkg/server"
)

var LlamaWorker *Worker

type Worker struct {
	ModelName string
	ModelPath string
	JobsNum   atomic.Int32
	MaxJobs   int32
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
		Predict: 512,
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
	jobsNum := LlamaWorker.JobsNum.Load()
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
	vocab, model, err := llama.LoadModel(params.Model, params, true)
	if err != nil {
		return err
	}

	server.MaxPods = opts.Pods
	server.Host = opts.Host
	server.Port = opts.Port
	server.Vocab = vocab
	server.Model = model
	server.Params = params
	log.Info("starting llamago server")
	go server.Run()
	LlamaWorker = &Worker{
		ModelName: modelName,
		ModelPath: modelPath,
		JobsNum:   atomic.Int32{},
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
	jobsNum := LlamaWorker.JobsNum.Load()
	if jobsNum > LlamaWorker.MaxJobs {
		return common.ErrExceedMaxJobs
	}
	LlamaWorker.JobsNum.Add(1)
	defer LlamaWorker.JobsNum.Add(-1)

	jobID := qa.ReqId
	server.PlaceJob(qa.ReqId, " "+qa.Prompt)
	qa.Answer = ""
	log.Infof("llama go handling job %v", qa)
	for {
		time.Sleep(100 * time.Millisecond)
		log.Debugf("llama go job check %v", qa.Answer)
		if _, ok := server.Jobs[jobID]; !ok {
			break
		}
		if qa.Answer != server.Jobs[jobID].Output {
			if len(server.Jobs[jobID].Output) < len(qa.Answer) {
				break
			}
			diff := server.Jobs[jobID].Output[len(qa.Answer):]
			qa.Answer += diff
		}
		if server.Jobs[jobID].Status == "finished" {
			break
		}
	}
	return nil
}
