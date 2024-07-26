package rpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"opml-opt/common"
	"opml-opt/llamago"
	"opml-opt/log"
	"opml-opt/mips"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const InternalError = "internal server error"

const (
	Success            = 200
	ErrorCodeUnknow    = -500
	ErrorCodeReadReq   = -501
	ErrorCodeParseReq  = -502
	ErrorCodeUnmarshal = -503
)

var Host = "127.0.0.1"

const (
	Avalible = 1
	InUse    = 2
	Down     = 0
)

var once sync.Once

var RpcServer *Service

type Service struct {
	port            string
	modelName       string
	modelPath       string
	pendingQuestion atomic.Int32
}

var NodeID = uuid.NewString()

func InitRpcService(port string, modelName, modelPath string) {
	once.Do(func() {
		RpcServer = &Service{}
		RpcServer.port = port
		RpcServer.pendingQuestion.Store(0)
		RpcServer.modelName = modelName
		RpcServer.modelPath = modelPath
	})
}

type LoggerMy struct {
}

func (*LoggerMy) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if strings.Index(msg, `"/healthcheck"`) > 0 {
		return
	}
	log.Debug(msg)
	return
}

func (c *Service) Start(ctx context.Context) error {
	//start gin
	gin.DefaultWriter = &LoggerMy{}
	r := gin.Default()
	pprof.Register(r)
	//cors middleware
	r.Use(Cors())
	r.SetTrustedProxies(nil)
	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	apiV1 := r.Group("/api/v1/")
	apiV1.POST("/question", c.HandleQuestion)
	apiV1.POST("/status", c.HandleStatus)

	address := "0.0.0.0:" + c.port
	r.Run(address)
	log.Info("start rpc on port:" + c.port)
	return nil
}

type Resp struct {
	ResultCode int         `json:"code"`
	ResultMsg  string      `json:"msg"`
	ResultBody interface{} `json:"data"`
}

type QuestionReq struct {
	Prompt   string `json:"prompt"`
	Model    string `json:"model"`
	CallBack string `json:"callback"`
	ReqId    string `json:"req_id"`
}

type QuestionResp struct {
	NodeId string `json:"node_id"`
	ReqId  string `json:"req_id"`
}

func (s *Service) HandleQuestion(c *gin.Context) {
	rep := Resp{
		ResultCode: ErrorCodeUnknow,
		ResultMsg:  InternalError,
		ResultBody: "",
	}
	defer func() {
		if rep.ResultCode == Success {
			c.JSON(http.StatusOK, rep)
		} else {
			c.JSON(http.StatusBadRequest, rep)
		}
	}()
	req := QuestionReq{}
	c.BindJSON(&req)
	reqId := req.ReqId
	qa := common.OptQA{
		ReqId:     reqId,
		Model:     req.Model,
		Prompt:    req.Prompt,
		Answer:    "",
		StateRoot: "",
		StartTime: time.Now().Unix(),
		CallBack:  req.CallBack,
	}

	if llamago.LlamaWorker.JobsNum >= llamago.LlamaWorker.MaxJobs || mips.MipsWork.JobsNum >= mips.MipsWork.MaxJobs {
		log.Info("job exceed")
		rep = Resp{
			ResultCode: -1,
			ResultMsg:  "jobs exceed",
			ResultBody: "",
		}
		return
	}

	go func() {
		err := llamago.Inference(qa)
		if err != nil {
			log.Warn("llamago inference error", err)
		}
	}()

	go func() {
		err := mips.Inference(qa)
		if err != nil {
			log.Warn("mips inference error", err)
		}
	}()

	data, _ := json.Marshal(QuestionResp{
		NodeId: NodeID,
		ReqId:  reqId,
	})

	rep = Resp{
		ResultCode: Success,
		ResultMsg:  "",
		ResultBody: string(data),
	}
}

type StatusResp struct {
	Status int    `json:"status"`
	NodeId string `json:"node_id"`
}

func (s *Service) HandleStatus(c *gin.Context) {
	status := mips.Status() | llamago.Status()
	data, _ := json.Marshal(&StatusResp{
		Status: status,
		NodeId: NodeID,
	})
	rep := Resp{
		ResultCode: 0,
		ResultMsg:  "",
		ResultBody: string(data),
	}
	c.JSON(http.StatusOK, rep)
}

func genWorkerID(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:])
}
