package callback

import (
	"opml-opt/common"
	"opml-opt/db"
	"sync"
)

var CallBack *CallBackService

type CallBackService struct {
	mu        sync.Mutex
	MipsWorks map[string]common.OptQA
}

func init() {
	CallBack = &CallBackService{
		MipsWorks: map[string]common.OptQA{},
	}
}

func (c *CallBackService) callBack() {

}

func DoneWork(qa common.OptQA) {
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
		db.InsertSingleConversation(qaExit)
		CallBack.mu.Unlock()
		CallBack.callBack()
	}
}
