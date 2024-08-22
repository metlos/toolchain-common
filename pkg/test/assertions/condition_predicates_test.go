package assertions

import (
	"testing"

	toolchainv1alpha1 "github.com/codeready-toolchain/api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestConditionPredicates(t *testing.T) {
	getConds := func(tc *toolchainv1alpha1.ToolchainCluster) *[]toolchainv1alpha1.Condition {
		return &tc.Status.Conditions
	}

	tcWith := func(conds []toolchainv1alpha1.Condition) *toolchainv1alpha1.ToolchainCluster {
		return &toolchainv1alpha1.ToolchainCluster{
			Status: toolchainv1alpha1.ToolchainClusterStatus{
				Conditions: conds,
			},
		}
	}

	t.Run("IsReady", func(t *testing.T) {
		tc := tcWith([]toolchainv1alpha1.Condition{
			{
				Type:   toolchainv1alpha1.ConditionReady,
				Status: corev1.ConditionTrue,
				Reason: "I don't care about the reason",
			},
		})
		AssertThat(t, tc, Has(ConditionIn[*toolchainv1alpha1.ToolchainCluster]().AccessedUsing(getConds).That(IsReady())))
	})

	t.Run("ReadyWithReason", func(t *testing.T) {
		tc := tcWith([]toolchainv1alpha1.Condition{
			{
				Type:   toolchainv1alpha1.ConditionReady,
				Status: corev1.ConditionTrue,
				Reason: "because",
			},
		})
		AssertThat(t, tc, Has(ConditionIn[*toolchainv1alpha1.ToolchainCluster]().AccessedUsing(getConds).That(IsReadyWithReason("because"))))
	})

	t.Run("Matches", func(t *testing.T) {
		tc := tcWith([]toolchainv1alpha1.Condition{
			{
				Type:    "someType",
				Status:  corev1.ConditionUnknown,
				Reason:  "some reason",
				Message: "some message",
			},
		})
		AssertThat(t, tc, Has(ConditionIn[*toolchainv1alpha1.ToolchainCluster]().AccessedUsing(getConds).That(MatchesCondition(toolchainv1alpha1.Condition{
			Type:   "someType",
			Status: corev1.ConditionUnknown,
		}))))
	})
}
