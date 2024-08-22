package assertions

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/codeready-toolchain/toolchain-common/pkg/condition"
)

// ConditionIn is the entry point for creating a Predicate on conditions of an Object of some type.
func ConditionIn[T client.Object]() AccessorGetter[T] {
	return AccessorGetter[T]{}
}

// MatchesCondition is condition test that succeeds if it finds a condition that matches the non-zero fields
// in the supplied prototypeCondition.
func MatchesCondition(prototypeCondition toolchainv1alpha1.Condition) ConditionsTest {
	return &matches{prototype: prototypeCondition}
}

// IsReady matches objects that have a condition with the Ready type with a True status.
func IsReady() ConditionsTest {
	return &ready{}
}

// IsReadyWithReason is similar to IsReady but the condition also needs to have the supplied
// reason.
func IsReadyWithReason(reason string) ConditionsTest {
	return &readyWithReason{expectedReason: reason}
}

// ConditionTest is a Predicate on a set of conditions. It can also "fix"
// the set of conditions so that the test would succeed.
//
// This is used to implement new tests on conditions and generally doesn't
// need to be implemented outside of this package.
type ConditionsTest interface {
	Test([]toolchainv1alpha1.Condition) bool
	Fix([]toolchainv1alpha1.Condition) []toolchainv1alpha1.Condition
}

// AccessorGetter is a "transitional" object that is used to provide a
// function to access the conditions in some type (e.g. ToolchainCluster, SpaceProvisionerConfig, etc.)
// This is used to build a predicate that can work on conditions in any type, starting with the ConditionIn function.
type AccessorGetter[T client.Object] struct{}

// ConditionTestBuilder is used to build the predicate on conditions of an Object of some type. Start with
// the ConditionIn function.
type ConditionTestBuilder[T client.Object] struct {
	accessor func(T) *[]toolchainv1alpha1.Condition
}

// AccessedUsing is used to provide a function using which the conditions can be accessed in an Object.
// It returns a "builder" that can be used supplied an actual test on a set of conditions.
func (g AccessorGetter[T]) AccessedUsing(accessor func(T) *[]toolchainv1alpha1.Condition) *ConditionTestBuilder[T] {
	return &ConditionTestBuilder[T]{
		accessor: accessor,
	}
}

// That returns a predicate on an Object of given type that will test it contains conditions matching
// the supplied test.
func (b *ConditionTestBuilder[T]) That(test ConditionsTest) Predicate[T] {
	return &hasConditions[T]{
		accessor: b.accessor,
		test:     test,
	}
}

type hasConditions[T client.Object] struct {
	accessor func(T) *[]toolchainv1alpha1.Condition
	test     ConditionsTest
}

// Matches implements Predicate.
func (h *hasConditions[T]) Matches(obj T) bool {
	conds := h.accessor(obj)
	return h.test.Test(*conds)
}

func (h *hasConditions[T]) FixToMatch(obj T) T {
	conds := h.accessor(obj)
	new_conds := h.test.Fix(*conds)
	*conds = new_conds
	return obj
}

type matches struct {
	prototype toolchainv1alpha1.Condition
}

func (m *matches) Test(conds []toolchainv1alpha1.Condition) bool {
	for _, c := range conds {
		c := c // memory aliasing protection
		if conditionsMatch(&m.prototype, &c) {
			return true
		}
	}
	return false
}

func (m *matches) Fix(conds []toolchainv1alpha1.Condition) []toolchainv1alpha1.Condition {
	conds, _ = condition.AddOrUpdateStatusConditions(conds, m.prototype)
	return conds
}

func conditionsMatch(prototype, val *toolchainv1alpha1.Condition) bool {
	if prototype.Type != "" && prototype.Type != val.Type {
		return false
	}

	if prototype.Reason != "" && prototype.Reason != val.Reason {
		return false
	}

	if prototype.Status != "" && prototype.Status != val.Status {
		return false
	}

	if prototype.Message != "" && prototype.Message != val.Message {
		return false
	}

	return true
}

type ready struct{}

func (*ready) Test(conds []toolchainv1alpha1.Condition) bool {
	return condition.IsTrue(conds, toolchainv1alpha1.ConditionReady)
}

func (*ready) Fix(conds []toolchainv1alpha1.Condition) []toolchainv1alpha1.Condition {
	conds, _ = condition.AddOrUpdateStatusConditions(conds, toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionTrue,
	})
	return conds
}

type readyWithReason struct {
	expectedReason string
}

func (p *readyWithReason) Test(conds []toolchainv1alpha1.Condition) bool {
	return condition.IsTrueWithReason(conds, toolchainv1alpha1.ConditionReady, p.expectedReason)
}

func (p *readyWithReason) Fix(conds []toolchainv1alpha1.Condition) []toolchainv1alpha1.Condition {
	cnd, found := condition.FindConditionByType(conds, toolchainv1alpha1.ConditionReady)
	if !found {
		return condition.AddStatusConditions(conds, toolchainv1alpha1.Condition{
			Type:   toolchainv1alpha1.ConditionReady,
			Status: corev1.ConditionTrue,
			Reason: p.expectedReason,
		})
	} else {
		cnd.Status = corev1.ConditionTrue
		cnd.Reason = p.expectedReason
		ret, _ := condition.AddOrUpdateStatusConditions(conds, cnd)
		return ret
	}
}
