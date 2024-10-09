# OPML on Eigenlayer

OPML (Optimistic Machine Learning) is an optimistic computing model that allows untrusted Operator nodes to execute inference tasks, while leveraging an on-chain dispute resolution mechanism to verify the correctness of the results. By integrating with Eigenlayer, OPML can take advantage of its distributed ledger and cryptoeconomic incentive infrastructure.



# Deploy
## Download model

**LLaMA-7B:** [llama-7b-fp32.bin](https://nogpu.com/llama-7b-fp32.bin)

**llama-2-7b-chat.Q2_K.gguf:** [llama-2-7b-chat.Q2_K.gguf](https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/blob/main/llama-2-7b-chat.Q2_K.gguf)

## Build
### for main branch
```
make build
```

### for fp16 branch

switch fp16 branch, and

```
git submodule update
make build
```

### for fp8 branch

switch fp8 branch, and

```
git submodule update
make build
```

## Config
config file:
```
port: "1234"
host: http://127.0.0.1
model_name: llama
model_path: ./llama-7b-fp32.bin # model filepath
mongo_uri: mongodb://admin:admin@127.0.0.1:27017
mips_program: ./mlgo/ml_mips/ml_mips.bin # mips program path
dispatcher: http://127.0.0.1:21001/ # dispatcher url
```
### Run
```
./opml-opt --config ./config.yml
```

# Operator API

## 1. POST /api/v1/status

Request:

Response:

```
{
    "code": 0, //code==0 success, otherwise failure
    "msg": "1", //1 for busy; 0 for free
    "data": {
		"status": 0 // 0 available, 1 busy
        "node_id": ""
    }
}
```

## 2. POST /api/v1/question

Request:

```
{
    "model": "llama-7b",
    "prompt": "hello",
    "callback": "http://abc.xyz/"
}
```

Response:

```
{
    "code": 0, //code==0 success, otherwise failure
    "msg": "",
    "data": {
        "node_id": "",
        "req_id": "bab34bd7-8415-4522-bb4a-6f62f3398b50"
    }
}
```

# Dispatcher Callback

## POST

Request:

```
{
    "node_id": "",
    "req_id": "bab34bd7-8415-4522-bb4a-6f62f3398b50",
    "model": "llama-7b",
    "prompt": "hello",
    "answer": "hello",
    "state_root": "0x130b06b347409671f3125f3c21b7fbeb720aba7bd2a8bd1b102634750a111686"
}
```
