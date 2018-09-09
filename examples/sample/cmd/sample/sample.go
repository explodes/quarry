package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/explodes/quarry"

	_ "github.com/explodes/quarry/examples/sample/samplelib"
	"github.com/explodes/quarry/examples/sample/samplepb"
	"github.com/explodes/quarry/examples/sample/samplequarry"
)

func main() {
	graph := samplequarry.Default()

	request := &samplepb.SampleRequest{
		Token:      "0xdeadbeef",
		ShowUnread: true,
	}
	makeRequest("with unread notifications", graph, request)

	request.ShowUnread = false
	makeRequest("without unread notifications", graph, request)
}

func makeRequest(name string, graph quarry.Quarry, request interface{}) {
	fmt.Println(name)
	fmt.Println(strings.Repeat("-", 15))

	result := graph.MustGet(context.Background(), request, "response")
	response := result.(*samplepb.SampleResponse)

	protoString := response.String()
	fmt.Println(protoString)
	fmt.Println()
}
