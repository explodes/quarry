package quarry

import (
	"context"
	"fmt"
	"sync"
)

// Quarry is a dependency graph to fulfill requirements that can provided
// by Factories.
type Quarry interface {
	// AddFactory registers a Factory by name.
	// Names must be unique.
	AddFactory(name string, factory Factory) error
	// MustAddFactory panics if AddFactory fails.
	MustAddFactory(name string, factory Factory)

	// AddDependency links two factory together as dependencies.
	// By default, dependencies are always fulfilled, but when conditions are present
	// they all must be met before fulfilling a dependency.
	// Dependencies that do not meet their required conditions are filled as nil.
	// Cycles are not allowed and result in an error.
	AddDependency(parent, dependsOn string, conditions ...Condition) error
	// MustAddDependency panics if AddDependency fails.
	MustAddDependency(parent, dependsOn string, conditions ...Condition)

	// Get will fetch an object by name using the parameters provided.
	// If any Factories return an error or the Context is done, the first
	// error encountered will be returned.
	Get(ctx context.Context, params interface{}, name string) (interface{}, error)
	// MustGet panics if Get fails.
	MustGet(ctx context.Context, params interface{}, name string) interface{}
}

// New creates a new Quarry.
func New() Quarry {
	return quarryImpl{
		adjacency: make(map[string]conditionMap),
		factories: make(map[string]Factory),
	}
}

// quarryImpl is the default implementation of Quarry.
type quarryImpl struct {
	// adjacency is a map of names of Factories to a set of names of
	// Factories the parent Factory depends on.
	adjacency map[string]conditionMap

	// factories is a map of names of Factories to Factories.
	factories map[string]Factory
}

func (q quarryImpl) MustAddFactory(name string, factory Factory) {
	if err := q.AddFactory(name, factory); err != nil {
		panic(err)
	}
}

func (q quarryImpl) AddFactory(name string, factory Factory) error {
	_, exists := q.factories[name]
	if exists {
		return fmt.Errorf("duplicate add of factory %s", name)
	}
	q.factories[name] = factory
	return nil
}

func (q quarryImpl) MustAddDependency(parent, dependsOn string, conditions ...Condition) {
	if err := q.AddDependency(parent, dependsOn, conditions...); err != nil {
		panic(err)
	}
}

func (q quarryImpl) AddDependency(parent, dependsOn string, conditions ...Condition) error {
	set, ok := q.adjacency[parent]
	if !ok {
		set = newConditionMap()
		q.adjacency[parent] = set
	}
	if set.Contains(dependsOn) {
		return fmt.Errorf("duplicate add of dependency on %s to %s", parent, dependsOn)
	}
	set.Add(dependsOn, conditions...)
	return q.checkCycles(parent, dependsOn)
}

// checkCycles will check the graph for cycles.
func (q quarryImpl) checkCycles(parent, dependsOn string) error {
	visited := newStringSet()
	stack := newStringSet()
	for name := range q.adjacency {
		if !visited.Contains(name) {
			if q.checkCyclesHelper(visited, stack, name) {
				return fmt.Errorf("depedending %s on %s creates a cycle", parent, dependsOn)
			}
		}
	}
	return nil
}

// checkCyclesHelper performs a recursive DFS to detect cycles.
func (q quarryImpl) checkCyclesHelper(visited, stack stringSet, name string) bool {
	visited.Add(name)
	stack.Add(name)
	neighbors := q.adjacency[name]
	for neighbor := range neighbors {
		if !visited.Contains(neighbor) {
			if q.checkCyclesHelper(visited, stack, neighbor) {
				return true
			}
		} else if stack.Contains(neighbor) {
			return true
		}
	}
	stack.Remove(name)
	return false
}

func (q quarryImpl) MustGet(ctx context.Context, params interface{}, name string) interface{} {
	result, err := q.Get(ctx, params, name)
	if err != nil {
		panic(err)
	}
	return result
}

func (q quarryImpl) Get(ctx context.Context, params interface{}, name string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(ctx)
	once := &onceController{
		q:     q,
		onces: make(map[string]*onceDelegate),
	}
	return once.getOnce(ctx, cancelFunc, params, "", name)
}

type onceController struct {
	q     quarryImpl
	m     sync.Mutex
	rw    sync.RWMutex
	onces map[string]*onceDelegate
}

type onceDelegate struct {
	result interface{}
	err    error
	f      func() (interface{}, error)
	once   sync.Once
}

func newOnceDelegate(f func() (interface{}, error)) *onceDelegate {
	return &onceDelegate{f: f}
}

func (o *onceDelegate) Do() (interface{}, error) {
	o.once.Do(func() {
		o.result, o.err = o.f()
	})
	return o.result, o.err
}

func (o *onceController) getOnce(ctx context.Context, cancelFunc func(), params interface{}, parent, name string) (interface{}, error) {
	o.m.Lock()
	delegate, ok := o.onces[name]
	if !ok {
		delegate = newOnceDelegate(func() (interface{}, error) {
			return o.getHelper(ctx, cancelFunc, params, parent, name)
		})
		o.onces[name] = delegate
	}
	o.m.Unlock()
	return delegate.Do()
}

// getHelper will fetch an object, resolving dependencies, until an error occurs or the Context is done.
func (o *onceController) getHelper(ctx context.Context, cancelFunc func(), params interface{}, parent, name string) (interface{}, error) {
	factory, factoryExists := o.q.factories[name]
	if !factoryExists {
		if parent == "" {
			return abort(cancelFunc, fmt.Errorf("factory %s does not exist", name))
		} else {
			return abort(cancelFunc, fmt.Errorf("factory %s, depended upon by %s, does not exist", name, parent))
		}
	}

	depConditions, hasDeps := o.q.adjacency[name]

	var deps Dependencies
	if hasDeps {
		if thisDeps, depsErr := o.getDependencies(ctx, cancelFunc, params, depConditions, parent, name); depsErr != nil {
			return abort(cancelFunc, depsErr)
		} else {
			deps = thisDeps
		}
	}
	result, err := factory(ctx, params, deps)
	if err != nil {
		return abort(cancelFunc, err)
	}
	return result, ctx.Err()
}

// getDependencies resolves all dependencies for a factory.
// Dependencies are resolved asynchronously.
func (o *onceController) getDependencies(ctx context.Context, cancelFunc func(), params interface{}, depConditions conditionMap, parent, name string) (Dependencies, error) {
	deps := make(Dependencies)
	if len(depConditions) == 0 {
		return deps, nil
	}
	var depsErr error
	m := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	wg.Add(depConditions.Size())
	for depName, conditions := range depConditions {
		go func(depName string, conditions []Condition) {
			select {
			case <-ctx.Done():
				depsErr = ctx.Err()
			default:
				var err error
				var result interface{}
				if !checkConditions(params, conditions) {
					result = nil
				} else {
					result, err = o.getOnce(ctx, cancelFunc, params, name, depName)
				}
				if err != nil {
					_, depsErr = abort(cancelFunc, err)
				} else {
					m.Lock()
					deps[depName] = result
					m.Unlock()
				}
			}
			wg.Done()
		}(depName, conditions)
	}
	wg.Wait()
	return deps, depsErr
}

// abort is a helper that calls a cancel function and returns a nil value and the error.
func abort(cancelFunc func(), err error) (interface{}, error) {
	cancelFunc()
	return nil, err
}
