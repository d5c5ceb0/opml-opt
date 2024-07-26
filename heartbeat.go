package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"opml-opt/callback"
	"opml-opt/llamago"
	"opml-opt/log"
	"opml-opt/mips"
	"time"
)

const HEART_BEAT_TIMER = time.Second * 5

func callHeartBeat(config Config) {
	ticker := time.NewTicker(HEART_BEAT_TIMER)
	workerUrl := fmt.Sprintf("%s:%s", config.Host, config.Port)
	hbUrl, err := url.JoinPath(config.DispatcherUrl, "receive_heart_beat")
	if err != nil {
		log.Fatal(err)
	}
	call := func() {
		mipsJobs := mips.MipsWork.JobsNum
		llamagoJobs := llamago.LlamaWorker.JobsNum
		queue := mipsJobs
		if llamagoJobs > mipsJobs {
			queue = llamagoJobs
		}
		body := struct {
			Type    string `json:"string"`
			Default struct {
				WorkerName  string `json:"worker_name"`
				QueueLength int32  `json:"queue_length"`
			}
		}{
			Type: "string",
			Default: struct {
				WorkerName  string "json:\"worker_name\""
				QueueLength int32  "json:\"queue_length\""
			}{
				WorkerName:  workerUrl,
				QueueLength: queue,
			},
		}
		data, _ := json.Marshal(&body)
		_, err = callback.DoPost(hbUrl, string(data), time.Second*3)
		if err != nil {
			log.Error(err)
		}
	}
	for {
		<-ticker.C
		call()
	}
}
