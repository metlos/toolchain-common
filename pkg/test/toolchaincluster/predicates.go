package toolchaincluster

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	. "github.com/codeready-toolchain/toolchain-common/pkg/test/assertions"
	corev1 "k8s.io/api/core/v1"
)

func getConditions(spc *toolchainv1alpha1.ToolchainCluster) *[]toolchainv1alpha1.Condition {
	return &spc.Status.Conditions
}

func Ready() Predicate[*toolchainv1alpha1.ToolchainCluster] {
	return ConditionIn[*toolchainv1alpha1.ToolchainCluster]().AccessedUsing(getConditions).That(MatchesCondition(toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionTrue,
	}))
}

func NotReady() Predicate[*toolchainv1alpha1.ToolchainCluster] {
	return ConditionIn[*toolchainv1alpha1.ToolchainCluster]().AccessedUsing(getConditions).That(MatchesCondition(toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionFalse,
	}))
}

func NotReadyWithReason(reason string) Predicate[*toolchainv1alpha1.ToolchainCluster] {
	return ConditionIn[*toolchainv1alpha1.ToolchainCluster]().AccessedUsing(getConditions).That(MatchesCondition(toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionFalse,
		Reason: reason,
	}))
}

func OperatorNamespace(operatorNamespace string) Predicate[*toolchainv1alpha1.ToolchainCluster] {
	return &forOperatorNamespace{ns: operatorNamespace}
}

type forOperatorNamespace struct {
	ns string
}

func (ons *forOperatorNamespace) Matches(tc *toolchainv1alpha1.ToolchainCluster) bool {
	// TODO: remove the check for the legacy field once both host and member operators are
	// updated with the new version of toolchain common.
	return tc.Status.OperatorNamespace == ons.ns || tc.Labels["namespace"] == ons.ns
}

func (ons *forOperatorNamespace) FixToMatch(tc *toolchainv1alpha1.ToolchainCluster) *toolchainv1alpha1.ToolchainCluster {
	tc.Status.OperatorNamespace = ons.ns
	return tc
}
