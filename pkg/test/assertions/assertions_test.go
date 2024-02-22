package assertions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestExplain(t *testing.T) {
	t.Run("with diff", func(t *testing.T) {
		// given
		actual := &corev1.Secret{}
		actual.SetName("actual")

		pred := Has(Name("expected"))

		// when
		expl := Explain(pred, actual)

		// then
		assert.True(t, strings.HasPrefix(expl, "predicate 'assertions.named' didn't match the object because of the following differences (- indicates the expected values, + the actual values):"))
		assert.Contains(t, expl, "-")
		assert.Contains(t, expl, "\"expected\"")
		assert.Contains(t, expl, "+")
		assert.Contains(t, expl, "\"actual\"")
	})

	t.Run("without diff", func(t *testing.T) {
		// given
		actual := &corev1.Secret{}
		actual.SetName("actual")

		pred := &predicateWithoutFixing{}

		// when
		expl := Explain(pred, actual)

		// then
		assert.Equal(t, expl, "predicate 'assertions.predicateWithoutFixing' didn't match the object")
	})
}

type predicateWithoutFixing struct{}

var _ Predicate[client.Object] = (*predicateWithoutFixing)(nil)

func (*predicateWithoutFixing) Matches(obj client.Object) bool {
	return false
}
