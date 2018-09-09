package quarry_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/explodes/quarry"
	"github.com/stretchr/testify/assert"
)

func TestProvider_returnsValue(t *testing.T) {
	const expectedValue = "hello-world"
	providerFactory := quarry.Provider(expectedValue)

	value, err := providerFactory(nil, nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, expectedValue, value)

}

func TestSingleton_onlyCallsFactoryOnce(t *testing.T) {
	var callCount int32
	onceFactory := quarry.Singleton(func(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return nil, nil
	})

	onceFactory(nil, nil, nil)
	onceFactory(nil, nil, nil)

	assert.Equal(t, int32(1), callCount)
}

func TestSingleton_returnsSameError(t *testing.T) {
	onceFactory := quarry.Singleton(func(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
		return nil, fmt.Errorf("some error")
	})

	_, err1 := onceFactory(nil, nil, nil)
	_, err2 := onceFactory(nil, nil, nil)

	assert.Equal(t, err1, err2)
}

func TestSingleton_returnsSameValue(t *testing.T) {
	onceFactory := quarry.Singleton(func(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
		return new(int32), nil
	})

	val1, _ := onceFactory(nil, nil, nil)
	val2, _ := onceFactory(nil, nil, nil)

	assert.Equal(t, val1, val2)
}
