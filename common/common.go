package common

// {
//     "node_id": "",
//     "req_id": "bab34bd7-8415-4522-bb4a-6f62f3398b50", //对应请求的req_id
//     "model": "llama-7b",
//     "prompt": "hello",
//     "answer": "hello",
//     "state_root": "0x130b06b347409671f3125f3c21b7fbeb720aba7bd2a8bd1b102634750a111686"
// }

type OptQA struct {
	ReqId     string `json:"req_id" bson:"reqId"`
	Model     string `json:"model" bson:"model"`
	Prompt    string `json:"prompt" bson:"prompt"`
	Answer    string `json:"answer" bson:"answer"`
	StateRoot string `json:"state_root" bson:"stateRoot"`
	StartTime int64  `json:"startTime" bson:"startTime"`
	CallBack  string `json:"callback"`
}

func (qa *OptQA) Done() bool {
	return qa.ReqId != "" && qa.Model != "" && qa.Answer != "" && qa.StateRoot != "" && qa.CallBack != ""
}
