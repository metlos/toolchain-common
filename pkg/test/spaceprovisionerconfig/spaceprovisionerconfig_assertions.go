package spaceprovisionerconfig

import (
	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	"github.com/codeready-toolchain/toolchain-common/pkg/test/assertions"
	corev1 "k8s.io/api/core/v1"
)

func getConditions(spc *toolchainv1alpha1.SpaceProvisionerConfig) *[]toolchainv1alpha1.Condition {
	return &spc.Status.Conditions
}

func Ready() assertions.Predicate[*toolchainv1alpha1.SpaceProvisionerConfig] {
	return assertions.ConditionIn[*toolchainv1alpha1.SpaceProvisionerConfig]().AccessedUsing(getConditions).That(assertions.MatchesCondition(toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionTrue,
		Reason: toolchainv1alpha1.SpaceProvisionerConfigValidReason,
	}))
}

func NotReady() assertions.Predicate[*toolchainv1alpha1.SpaceProvisionerConfig] {
	return assertions.ConditionIn[*toolchainv1alpha1.SpaceProvisionerConfig]().AccessedUsing(getConditions).That(assertions.MatchesCondition(toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionFalse,
	}))
}

func NotReadyWithReason(reason string) assertions.Predicate[*toolchainv1alpha1.SpaceProvisionerConfig] {
	return assertions.ConditionIn[*toolchainv1alpha1.SpaceProvisionerConfig]().AccessedUsing(getConditions).That(assertions.MatchesCondition(toolchainv1alpha1.Condition{
		Type:   toolchainv1alpha1.ConditionReady,
		Status: corev1.ConditionFalse,
		Reason: reason,
	}))
}
