package quarry_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/explodes/quarry"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	q := quarry.New()

	assert.NotNil(t, q)
}

func TestQuarryImpl_MustAddFactory_panicsOnDuplicate(t *testing.T) {
	const factoryName = "some-factory"
	q := quarry.New()
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()
	q.AddFactory(factoryName, factoryOk())

	q.MustAddFactory(factoryName, factoryOk())
}

func TestQuarryImpl_AddFactory_duplicateResultsInError(t *testing.T) {
	const factoryName = "some-factory"
	q := quarry.New()

	err1 := q.AddFactory(factoryName, factoryOk())
	err2 := q.AddFactory(factoryName, factoryOk())

	assert.NoError(t, err1)
	assert.Error(t, err2)
}

func TestQuarryImpl_AddDependency_panicsOnDuplicate(t *testing.T) {
	q := quarry.New()
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()
	q.AddDependency("parent", "child")

	q.MustAddDependency("parent", "child")
}

func TestQuarryImpl_AddDependency_duplicateResultsInError(t *testing.T) {
	q := quarry.New()

	err1 := q.AddDependency("parent", "child")
	err2 := q.AddDependency("parent", "child")

	assert.NoError(t, err1)
	assert.Error(t, err2)
}

func TestQuarryImpl_AddDependency_cycleResultsInError(t *testing.T) {
	q := quarry.New()
	q.AddDependency("a", "b")
	q.AddDependency("b", "c")
	q.AddDependency("c", "d")

	err := q.AddDependency("d", "a")

	assert.Error(t, err)
}

func TestQuarryImpl_Get(t *testing.T) {
	_, q := simpleGraph(nil)

	value, err := q.Get(context.Background(), nil, "root")

	assert.NoError(t, err)
	assert.NotNil(t, value)
}

func TestQuarryImpl_MustGet_panicsOnError(t *testing.T) {
	q := quarry.New()
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()
	q.AddFactory("root", factoryOk())
	q.AddFactory("err", factoryError())
	q.AddDependency("root", "err")

	q.MustGet(context.Background(), nil, "root")
}

func TestQuarryImpl_MustGet_doesntPanicOnOk(t *testing.T) {
	q := quarry.New()
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()
	q.AddFactory("root", factoryOk())
	q.AddFactory("ok", factoryOk())
	q.AddDependency("root", "ok")

	q.MustGet(context.Background(), nil, "root")
}

func TestQuarryImpl_MustGet_visitsAllFactories(t *testing.T) {
	count, counter := factoryCounter()
	size, q := simpleGraph(counter)

	value, err := q.Get(context.Background(), nil, "root")

	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, int32(size), *count)
}

func TestQuarryImpl_Get_parentMissingDepResultsInError(t *testing.T) {
	q := quarry.New()
	q.AddFactory("root", factoryOk())
	q.AddDependency("root", "doesnt-exist")

	_, err := q.Get(context.Background(), nil, "root")

	assert.Error(t, err)
}

func TestQuarryImpl_Get_missingFactoryResultsInError(t *testing.T) {
	q := quarry.New()

	_, err := q.Get(context.Background(), nil, "root")

	assert.Error(t, err)
}

func TestQuarryImpl_Get_doneContextResultsInError(t *testing.T) {
	q := quarry.New()
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := q.Get(ctx, nil, "any")

	assert.Error(t, err)
}

func TestQuarryImpl_Get_contextDoneHalfwayResultsInError(t *testing.T) {
	q := quarry.New()
	ctx, cancelFunc := context.WithCancel(context.Background())
	finishFactory := func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		cancelFunc()
		return nil, nil
	}
	q.AddFactory("root", factoryOk())
	q.AddFactory("a", factoryOk())
	q.AddFactory("b", finishFactory)
	q.AddFactory("c", factoryOk())
	q.AddDependency("root", "a")
	q.AddDependency("a", "b")
	q.AddDependency("b", "c")

	_, err := q.Get(ctx, nil, "root")

	assert.Error(t, err)
}

func TestQuarryImpl_Get_passesParams(t *testing.T) {
	q := quarry.New()
	expectedParams := new(int64)
	count := int32(0)
	paramsFactory := func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		atomic.AddInt32(&count, 1)
		assert.Equal(t, expectedParams, params)
		return params, nil
	}
	q.AddFactory("root", paramsFactory)
	q.AddFactory("a", paramsFactory)
	q.AddFactory("b", paramsFactory)
	q.AddFactory("c", paramsFactory)
	q.AddDependency("root", "a")
	q.AddDependency("a", "b")
	q.AddDependency("b", "c")

	value, err := q.Get(context.Background(), expectedParams, "root")

	assert.NoError(t, err)
	assert.Equal(t, expectedParams, value)
	assert.Equal(t, int32(count), count)
}

func TestQuarryImpl_Get_doesNotResolveDependenciesTwice(t *testing.T) {
	q := quarry.New()
	count, counter := factoryCounter()
	q.AddFactory("root", counter)
	q.AddFactory("a", counter)
	q.AddFactory("b", counter)
	q.AddFactory("c", counter)
	q.AddFactory("d", counter)
	q.AddFactory("e", counter)
	q.AddFactory("f", counter)
	q.AddDependency("root", "a")
	q.AddDependency("root", "e")
	q.AddDependency("e", "f")
	q.AddDependency("e", "d")
	q.AddDependency("a", "b")
	q.AddDependency("a", "c")
	q.AddDependency("b", "c")
	q.AddDependency("c", "d")

	_, err := q.Get(context.Background(), nil, "root")

	assert.NoError(t, err)
	assert.Equal(t, int32(7), *count)
}

func TestQuarryImpl_Get_withFailedConditionsDoesntUseDeps(t *testing.T) {
	q := quarry.New()
	count, counter := factoryCounter()
	q.AddFactory("root", factoryWithNonNilDeps(counter))
	q.AddFactory("a", counter)
	q.AddFactory("b", counter)
	q.AddFactory("c", counter)
	q.AddDependency("root", "a", func(params interface{}) bool {
		return params.(bool)
	})
	q.AddDependency("a", "b")
	q.AddDependency("b", "c")

	_, err := q.Get(context.Background(), false, "root")

	assert.NoError(t, err)
	assert.Equal(t, int32(1), *count)
}

func TestQuarryImpl_Get_withSucceedingConditionsUsesDeps(t *testing.T) {
	q := quarry.New()
	count, counter := factoryCounter()
	q.AddFactory("root", factoryWithNonNilDeps(counter, "a"))
	q.AddFactory("a", counter)
	q.AddFactory("b", counter)
	q.AddFactory("c", counter)
	q.AddDependency("root", "a", func(params interface{}) bool {
		return params.(bool)
	})
	q.AddDependency("a", "b")
	q.AddDependency("b", "c")

	_, err := q.Get(context.Background(), true, "root")

	assert.NoError(t, err)
	assert.Equal(t, int32(4), *count)
}

// ## BENCHMARKS ##

func BenchmarkQuarryImpl_Get(b *testing.B) {
	factory := factoryOk()
	q := quarry.New()
	q.MustAddFactory("root", factory)
	q.MustAddFactory("a", factory)
	q.MustAddFactory("b", factory)
	q.MustAddFactory("c", factory)
	q.MustAddFactory("d", factory)
	q.MustAddFactory("e", factory)
	q.MustAddFactory("f", factory)
	q.MustAddFactory("g", factory)
	q.MustAddFactory("h", factory)
	q.MustAddFactory("i", factory)
	// j is not used by root.
	q.MustAddFactory("j", factory)
	q.MustAddDependency("root", "a")
	q.MustAddDependency("a", "c")
	q.MustAddDependency("c", "d")
	q.MustAddDependency("c", "f")
	q.MustAddDependency("d", "e")
	q.MustAddDependency("root", "b")
	q.MustAddDependency("b", "g")
	q.MustAddDependency("g", "h")
	q.MustAddDependency("h", "i")
	q.MustAddDependency("j", "h")
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		q.Get(ctx, nil, "root")
	}
}

func BenchmarkQuarryImpl_Get_longerKeys(b *testing.B) {
	factory := factoryOk()
	q := quarry.New()
	q.MustAddFactory("rootrootrootrootrootrootrootrootrootroot", factory)
	q.MustAddFactory("aaaaaaaaaa", factory)
	q.MustAddFactory("bbbbbbbbbb", factory)
	q.MustAddFactory("cccccccccc", factory)
	q.MustAddFactory("dddddddddd", factory)
	q.MustAddFactory("eeeeeeeeee", factory)
	q.MustAddFactory("ffffffffff", factory)
	q.MustAddFactory("gggggggggg", factory)
	q.MustAddFactory("hhhhhhhhhh", factory)
	q.MustAddFactory("iiiiiiiiii", factory)
	// j is not used by root.
	q.MustAddFactory("jjjjjjjjjj", factory)
	q.MustAddDependency("rootrootrootrootrootrootrootrootrootroot", "aaaaaaaaaa")
	q.MustAddDependency("aaaaaaaaaa", "cccccccccc")
	q.MustAddDependency("cccccccccc", "dddddddddd")
	q.MustAddDependency("cccccccccc", "ffffffffff")
	q.MustAddDependency("dddddddddd", "eeeeeeeeee")
	q.MustAddDependency("rootrootrootrootrootrootrootrootrootroot", "bbbbbbbbbb")
	q.MustAddDependency("bbbbbbbbbb", "gggggggggg")
	q.MustAddDependency("gggggggggg", "hhhhhhhhhh")
	q.MustAddDependency("hhhhhhhhhh", "iiiiiiiiii")
	q.MustAddDependency("jjjjjjjjjj", "hhhhhhhhhh")
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		q.Get(ctx, nil, "root")
	}
}

// ## UTILS ##

func factoryValue(value interface{}) quarry.Factory {
	return func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		return value, nil
	}
}

func factoryErr(err error) quarry.Factory {
	return func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		return nil, err
	}
}

func factoryOk() quarry.Factory {
	return factoryValue(0)
}

func factoryError() quarry.Factory {
	return factoryErr(errors.New("some-error"))
}

func keys(deps quarry.Dependencies) []string {
	var set []string
	for k := range deps {
		set = append(set, k)
	}
	return set
}

func factoryWithDeps(base quarry.Factory, names ...string) quarry.Factory {
	return func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		if err := checkDeps(names, deps); err != nil {
			return nil, err
		}
		return base(ctx, params, deps)
	}
}

func factoryWithNonNilDeps(base quarry.Factory, names ...string) quarry.Factory {
	return func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		notNilDeps := make(quarry.Dependencies, len(deps))
		for name, dep := range deps {
			if dep != nil {
				notNilDeps[name] = dep
			}
		}
		if err := checkDeps(names, notNilDeps); err != nil {
			return nil, err
		}
		return base(ctx, params, deps)
	}
}

func checkDeps(names []string, deps quarry.Dependencies) error {
	if len(names) != len(deps) {
		return fmt.Errorf("deps mismatch: %v, got %v", names, keys(deps))
	}
	for _, name := range names {
		if !deps.Contains(name) {
			return fmt.Errorf("deps not passed: want %v, got %v", names, keys(deps))
		}
	}
	return nil
}

func simpleGraph(factory quarry.Factory) (numRootDeps int, q quarry.Quarry) {
	q = quarry.New()

	if factory == nil {
		factory = factoryOk()
	}
	q.MustAddFactory("root", factoryWithDeps(factory, "a", "b"))
	q.MustAddFactory("a", factoryWithDeps(factory, "c"))
	q.MustAddFactory("b", factoryWithDeps(factory, "g"))
	q.MustAddFactory("c", factoryWithDeps(factory, "d", "f"))
	q.MustAddFactory("d", factoryWithDeps(factory, "e"))
	q.MustAddFactory("e", factoryWithDeps(factory))
	q.MustAddFactory("f", factoryWithDeps(factory))
	q.MustAddFactory("g", factoryWithDeps(factory, "h"))
	q.MustAddFactory("h", factoryWithDeps(factory, "i"))
	q.MustAddFactory("i", factoryWithDeps(factory))
	// j is not used by root.
	q.MustAddFactory("j", factoryWithDeps(factory, "h"))
	q.MustAddDependency("root", "a")
	q.MustAddDependency("a", "c")
	q.MustAddDependency("c", "d")
	q.MustAddDependency("c", "f")
	q.MustAddDependency("d", "e")
	q.MustAddDependency("root", "b")
	q.MustAddDependency("b", "g")
	q.MustAddDependency("g", "h")
	q.MustAddDependency("h", "i")
	q.MustAddDependency("j", "h")

	return 10, q
}

func factoryCounter() (*int32, quarry.Factory) {
	var count int32
	factory := func(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
		thisCount := atomic.AddInt32(&count, 1)
		return thisCount, nil
	}
	return &count, factory
}
