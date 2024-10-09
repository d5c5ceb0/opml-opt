package vm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"

	llama "mlgo/examples/llama/llama_go"
	"mlgo/ml"
)

func LLAMA(nodeID int, modelFile string, prompt string) ([]byte, int, error) {
	if modelFile == "" {
		modelFile = "./mlgo/examples/llama/models/llama-7b-fp32.bin"
	}
	if prompt == "" {
		prompt = "How to combine AI and blockchain?"
	}

	threadCount := 32
	ctx, err := llama.LoadModel(modelFile, true)
	fmt.Println("Load Model Finish")
	if err != nil {
		fmt.Println("load model error: ", err)
		return nil, 0, err
	}
	defer func() {
		ctx.Model = nil
		ctx.Vocab = nil
		ctx.Embedding = nil
		ctx.Logits = nil
		ctx = nil
		runtime.GC()
	}()
	embd := ml.Tokenize(ctx.Vocab, prompt, true)
	graph, mlctx, err := llama.ExpandGraph(ctx, embd, uint32(len(embd)), 0, threadCount)
	ml.GraphComputeByNodes(mlctx, graph, nodeID)
	envBytes := ml.SaveComputeNodeEnvToBytes(uint32(nodeID), graph.Nodes[nodeID], graph, true)
	return envBytes, int(graph.NodesCount), nil
}

func MNIST(nodeID int, modelFile string, dataFile string) ([]byte, int, error) {
	return nil, 0, errors.New("unsupport")
}

func MNIST_Input(dataFile string, show bool) ([]float32, error) {
	buf, err := ioutil.ReadFile(dataFile)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	digits := make([]float32, 784)

	// render the digit in ASCII
	var c string
	for row := 0; row < 28; row++ {
		for col := 0; col < 28; col++ {
			digits[row*28+col] = float32(buf[row*28+col])
			if buf[row*28+col] > 230 {
				c += "*"
			} else {
				c += "_"
			}
		}
		c += "\n"
	}
	if show {
		fmt.Println(c)
	}

	return digits, nil
}
