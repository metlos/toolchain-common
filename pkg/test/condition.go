package test

import (
	"time"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/require"

	conditions "github.com/codeready-toolchain/toolchain-common/pkg/condition"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// AssertConditionsMatch asserts that the specified list A of conditions is equal to specified
// list B of conditions ignoring the order of the elements. We can't use assert.ElementsMatch
// because the LastTransitionTime of the actual conditions can be modified but the conditions
// still should be treated as matched
func AssertConditionsMatch(t T, actual []toolchainv1alpha1.Condition, expected ...toolchainv1alpha1.Condition) {
	require.Equal(t, len(expected), len(actual))
	for _, c := range expected {
		AssertContainsCondition(t, actual, c)
	}
}

// AssertContainsCondition asserts that the specified list of conditions contains the specified condition.
// LastTransitionTime is ignored.
func AssertContainsCondition(t T, conditions []toolchainv1alpha1.Condition, contains toolchainv1alpha1.Condition) {
	AssertThat(t, conditions, Has(ConditionThat(contains.Type, HasStatus(contains.Status), HasReason(contains.Reason), HasMessage(contains.Message))))
}

// AssertConditionsMatchAndRecentTimestamps asserts that the specified list of conditions match AND asserts that the timestamps are recent
func AssertConditionsMatchAndRecentTimestamps(t T, actual []toolchainv1alpha1.Condition, expected ...toolchainv1alpha1.Condition) {
	require.Equal(t, len(expected), len(actual))

	cutoff := time.Now().Add(-5 * time.Second)
	for _, c := range expected {
		AssertThat(t, actual, Has(ConditionThat(c.Type,
			HasStatus(c.Status),
			HasReason(c.Reason),
			HasMessage(c.Message),
			HasTransitionTimeLaterThan(cutoff),
			HasUpdateTimeLaterThan(cutoff))))
	}
}

// ConditionsMatch returns true if the specified list A of conditions is equal to specified
// list B of conditions ignoring the order of the elements
func ConditionsMatch(actual []toolchainv1alpha1.Condition, expected ...toolchainv1alpha1.Condition) bool {
	if len(expected) != len(actual) {
		return false
	}
	for _, c := range expected {
		if !ContainsCondition(actual, c) {
			return false
		}
	}
	for _, c := range actual {
		if !ContainsCondition(expected, c) {
			return false
		}
	}
	return true
}

// ContainsCondition returns true if the specified list of conditions contains the specified condition.
// LastTransitionTime is ignored.
func ContainsCondition(conditions []toolchainv1alpha1.Condition, contains toolchainv1alpha1.Condition) bool {
	for _, c := range conditions {
		if c.Type == contains.Type {
			return contains.Status == c.Status && contains.Reason == c.Reason && contains.Message == c.Message
		}
	}
	return false
}

func ConditionThat(conditionType toolchainv1alpha1.ConditionType, preds ...Predicate[toolchainv1alpha1.Condition]) Predicate[[]toolchainv1alpha1.Condition] {
	return &conditionsPredicate{conditionType: conditionType, predicates: preds}
}

func ConditionOnObject[T client.Object](accessor func(T) *[]toolchainv1alpha1.Condition, conditionType toolchainv1alpha1.ConditionType, preds ...Predicate[toolchainv1alpha1.Condition]) Predicate[T] {
	return &conditionsOnObjectPredicate[T]{accessor: accessor, conditionsPredicate: conditionsPredicate{conditionType: conditionType, predicates: preds}}
}

func IsTrue() Predicate[toolchainv1alpha1.Condition] {
	return HasStatus(corev1.ConditionTrue)
}

func IsNotTrue() Predicate[toolchainv1alpha1.Condition] {
	return HasStatusDifferentFrom(corev1.ConditionTrue)
}

func IsFalse() Predicate[toolchainv1alpha1.Condition] {
	return HasStatus(corev1.ConditionFalse)
}

func IsNotFalse() Predicate[toolchainv1alpha1.Condition] {
	return HasStatusDifferentFrom(corev1.ConditionFalse)
}

func IsUnknown() Predicate[toolchainv1alpha1.Condition] {
	return HasStatus(corev1.ConditionUnknown)
}

func IsNotUnknown() Predicate[toolchainv1alpha1.Condition] {
	return HasStatusDifferentFrom(corev1.ConditionUnknown)
}

func HasStatus(status corev1.ConditionStatus) Predicate[toolchainv1alpha1.Condition] {
	return &conditionPredicate{expectedStatus: &status}
}

func HasStatusDifferentFrom(status corev1.ConditionStatus) Predicate[toolchainv1alpha1.Condition] {
	return &conditionPredicate{expectedStatus: &status, negate: true}
}

func HasReason(reason string) Predicate[toolchainv1alpha1.Condition] {
	return &conditionPredicate{expectedReason: pointer.String(reason)}
}

func HasMessage(reason string) Predicate[toolchainv1alpha1.Condition] {
	return &conditionPredicate{expectedMessage: pointer.String(reason)}
}

func HasRecentTransitionTime() Predicate[toolchainv1alpha1.Condition] {
	return HasTransitionTimeLaterThan(time.Now().Add(-5 * time.Second))
}

func HasRecentUpdateTime() Predicate[toolchainv1alpha1.Condition] {
	return HasUpdateTimeLaterThan(time.Now().Add(-5 * time.Second))
}

func HasTransitionTimeLaterThan(t time.Time) Predicate[toolchainv1alpha1.Condition] {
	return &recencyPredicate{oldestAllowed: t, transition: true}
}

func HasUpdateTimeLaterThan(t time.Time) Predicate[toolchainv1alpha1.Condition] {
	return &recencyPredicate{oldestAllowed: t, transition: false}
}

type conditionsPredicate struct {
	conditionType toolchainv1alpha1.ConditionType
	predicates    []Predicate[toolchainv1alpha1.Condition]
}

// Matches implements Predicate.
func (c *conditionsPredicate) Matches(conds []toolchainv1alpha1.Condition) bool {
	condition, found := conditions.FindConditionByType(conds, c.conditionType)
	if !found {
		return false
	}

	for _, predicate := range c.predicates {
		if !predicate.Matches(condition) {
			return false
		}
	}
	return true
}

func (c *conditionsPredicate) FixToMatch(conds []toolchainv1alpha1.Condition) []toolchainv1alpha1.Condition {
	var found bool
	var index int
	var condition toolchainv1alpha1.Condition

	if conds == nil {
		conds = []toolchainv1alpha1.Condition{}
	} else {
		copy := make([]toolchainv1alpha1.Condition, len(conds))
		for i, c := range conds {
			copy[i] = c
		}
		conds = copy
	}

	for i, cond := range conds {
		if cond.Type == c.conditionType {
			found = true
			index = i
			condition = cond
			break
		}
	}

	if !found {
		conds = append(conds, toolchainv1alpha1.Condition{})
		index = len(conds) - 1
		condition.Type = c.conditionType
	}

	for _, predicate := range c.predicates {
		if p, ok := predicate.(PredicateMatchFixer[toolchainv1alpha1.Condition]); ok {
			condition = p.FixToMatch(condition)
		}
	}

	conds[index] = condition

	return conds
}

type conditionPredicate struct {
	expectedStatus  *corev1.ConditionStatus
	expectedReason  *string
	expectedMessage *string
	expectedType    toolchainv1alpha1.ConditionType
	negate          bool
}

func (c *conditionPredicate) Matches(cond toolchainv1alpha1.Condition) bool {
	if c.expectedType != "" && cond.Type != c.expectedType {
		return c.negate
	}
	if c.expectedStatus != nil && cond.Status != *c.expectedStatus {
		return c.negate
	}
	if c.expectedReason != nil && cond.Reason != *c.expectedReason {
		return c.negate
	}
	if c.expectedMessage != nil && cond.Message != *c.expectedMessage {
		return c.negate
	}

	return !c.negate
}

func (c *conditionPredicate) FixToMatch(cond toolchainv1alpha1.Condition) toolchainv1alpha1.Condition {
	if c.expectedType != "" {
		if c.negate {
			cond.Type = "<different from: " + c.expectedType + ">"
		} else {
			cond.Type = c.expectedType
		}
	}
	if c.expectedStatus != nil {
		if c.negate {
			cond.Status = "<different from: " + *c.expectedStatus + ">"
		} else {
			cond.Status = *c.expectedStatus
		}
	}
	if c.expectedReason != nil {
		if c.negate {
			cond.Reason = "<different from: " + *c.expectedReason + ">"
		} else {
			cond.Reason = *c.expectedReason
		}
	}
	if c.expectedMessage != nil {
		if c.negate {
			cond.Message = "<different from: " + *c.expectedMessage + ">"
		} else {
			cond.Message = *c.expectedMessage
		}
	}

	return cond
}

type conditionsOnObjectPredicate[T client.Object] struct {
	accessor func(T) *[]toolchainv1alpha1.Condition
	conditionsPredicate
}

func (c *conditionsOnObjectPredicate[T]) Matches(obj T) bool {
	conds := c.accessor(obj)
	return c.conditionsPredicate.Matches(*conds)
}

func (c *conditionsOnObjectPredicate[T]) FixToMatch(obj T) T {
	obj = obj.DeepCopyObject().(T)
	conds := c.accessor(obj)
	*conds = c.conditionsPredicate.FixToMatch(*conds)
	return obj
}

type recencyPredicate struct {
	oldestAllowed time.Time
	transition    bool
}

// Matches implements ConditionPredicate.
func (r *recencyPredicate) Matches(cond toolchainv1alpha1.Condition) bool {
	var condTime time.Time
	if r.transition {
		condTime = cond.LastTransitionTime.Time
	} else if cond.LastUpdatedTime != nil {
		condTime = cond.LastUpdatedTime.Time
	} else {
		// we're looking for a recent update time, but there was no update
		return false
	}

	return condTime.After(r.oldestAllowed)
}

// FixToMatch implements ConditionPredicate.
func (r *recencyPredicate) FixToMatch(cond toolchainv1alpha1.Condition) toolchainv1alpha1.Condition {
	if r.transition {
		cond.LastTransitionTime.Time = r.oldestAllowed
	} else {
		cond.LastUpdatedTime = &metav1.Time{Time: r.oldestAllowed}
	}
	return cond
}

var (
	_ Predicate[[]toolchainv1alpha1.Condition]           = (*conditionsPredicate)(nil)
	_ PredicateMatchFixer[[]toolchainv1alpha1.Condition] = (*conditionsPredicate)(nil)
	_ Predicate[toolchainv1alpha1.Condition]             = (*conditionPredicate)(nil)
	_ PredicateMatchFixer[toolchainv1alpha1.Condition]   = (*conditionPredicate)(nil)
	_ Predicate[toolchainv1alpha1.Condition]             = (*recencyPredicate)(nil)
	_ PredicateMatchFixer[toolchainv1alpha1.Condition]   = (*recencyPredicate)(nil)
)
