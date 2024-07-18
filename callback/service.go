package callback

import (
	"bytes"
	"encoding/json"
	"net/http"
	"opml-opt/common"
	"opml-opt/log"
	"sync"
	"time"
)

var CallBack *CallBackService

const CALLBACK_TIMEOUT = time.Second * 60

type CallBackService struct {
	mu        sync.Mutex
	MipsWorks map[string]common.OptQA
}

func init() {
	CallBack = &CallBackService{
		MipsWorks: map[string]common.OptQA{},
	}
}

func (c *CallBackService) callBack(qa common.OptQA) {
	log.Debugf("work done \n%v\n", qa)
	reqBody, _ := json.Marshal(&common.CallbackReq{
		NodeId:    common.NodeID,
		ReqId:     qa.ReqId,
		Model:     qa.Model,
		Prompt:    qa.Prompt,
		Answer:    qa.Answer,
		StateRoot: qa.StateRoot,
	})
	_, err := DoPost(qa.CallBack, string(reqBody), CALLBACK_TIMEOUT)
	if err != nil {
		log.Errorf("callback post error\n %v \n %v \n %v\n", qa.CallBack, string(reqBody), err)
	}
}

func DoneWork(qa common.OptQA) {
	log.Debugf("done work %v", qa)
	CallBack.mu.Lock()
	defer CallBack.mu.Unlock()
	if qa.CallBack == "" {
		return
	}
	if qa.Model == "" {
		return
	}
	qaExit, ok := CallBack.MipsWorks[qa.ReqId]
	if !ok {
		qaExit = qa
	}
	if qa.Answer != "" {
		qaExit.Answer = qa.Answer
	}
	if qa.StateRoot != "" {
		qaExit.StateRoot = qa.StateRoot
	}
	CallBack.MipsWorks[qa.ReqId] = qaExit
	if qaExit.Done() {
		delete(CallBack.MipsWorks, qa.ReqId)
		// db.InsertSingleConversation(qaExit)
		go CallBack.callBack(qaExit)
	}
}

func DoPost(requrl, body string, timeoutS time.Duration) (*http.Response, error) {
	req, err := http.NewRequest("POST", requrl, bytes.NewBufferString(body))
	if err != nil {
		log.Errorf("Failed to new http request:%s", err.Error())
		return nil, err
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{Timeout: timeoutS}
	resp, err := client.Do(req)
	return resp, err
}
