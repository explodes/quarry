package samplelib

import (
	"context"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/sample/samplepb"
	"github.com/explodes/quarry/examples/sample/samplequarry"
)

func init() {
	graph := samplequarry.Default()

	graph.MustAddFactory("response", buildResponse)
	graph.MustAddDependency("response", "user")
	graph.MustAddDependency("response", "inbox")
}

func buildResponse(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	response := &samplepb.SampleResponse{
		User:  deps["user"].(*samplepb.User),
		Inbox: deps["inbox"].(*samplepb.Inbox),
	}
	return response, nil
}
