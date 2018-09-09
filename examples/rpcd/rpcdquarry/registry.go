package rpcdquarry

import "github.com/explodes/quarry"

var graph = quarry.New()

func Default() quarry.Quarry {
	return graph
}
