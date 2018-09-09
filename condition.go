package quarry

// Condition defines when a dependency should be fulfilled. By default, dependencies
// are always fulfilled, but when conditions are present they all must be met before
// fulfilling a dependency.
// Dependencies that do not meet their required conditions are filled as nil.
type Condition func(params interface{}) bool

type conditionMap map[string][]Condition

func newConditionMap() conditionMap {
	return make(conditionMap)
}

func (c conditionMap) Size() int {
	return len(c)
}

func (c conditionMap) Add(value string, conditions ...Condition) {
	c[value] = conditions
}

func (c conditionMap) Contains(val string) bool {
	_, ok := c[val]
	return ok
}

// checkConditions ensures that all conditions are met.
// If there are no conditions, the conditions are considered to be met.
func checkConditions(params interface{}, conditions []Condition) bool {
	// Quickest path out.
	if len(conditions) == 0 {
		return true
	}
	// If any condition fails, the check fails.
	for _, condition := range conditions {
		if !condition(params) {
			return false
		}
	}
	return true
}
