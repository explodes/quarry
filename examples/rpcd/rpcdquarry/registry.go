package rpcdquarry

import "github.com/explodes/scratch/quarry"

var graph = quarry.New()

func Default() quarry.Quarry {
	return graph
}
