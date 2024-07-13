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
	//cors middleware
	r.Use(Cors())
	r.SetTrustedProxies(nil)
	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	apiV1 := r.Group("/api/v1/")
	apiV1.POST("/question", c.HandleQuestion)

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
			c.JSON(http.StatusInternalServerError, rep)
		}
	}()
	req := QuestionReq{}
	c.BindJSON(&req)
	reqId := uuid.NewString()
	qa := common.OptQA{
		ReqId:     reqId,
		Model:     req.Model,
		Prompt:    req.Prompt,
		Answer:    "",
		StateRoot: "",
		StartTime: time.Now().Unix(),
		CallBack:  req.CallBack,
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
		ResultCode: 0,
		ResultMsg:  "",
		ResultBody: data,
	}
}

func genWorkerID(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:])
}
