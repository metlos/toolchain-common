package assertions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNamePredicate(t *testing.T) {
	pred := Name("expected")

	t.Run("positive", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetName("expected")

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("negative", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetName("different")

		// when & then
		assert.False(t, pred.Matches(obj))
	})
}

func TestInNamespacePredicate(t *testing.T) {
	pred := InNamespace("expected")

	t.Run("positive", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetNamespace("expected")

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("negative", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetNamespace("different")

		// when & then
		assert.False(t, pred.Matches(obj))
	})
}

func TestWithKeyPredicate(t *testing.T) {
	pred := ObjectKey(client.ObjectKey{Name: "expected", Namespace: "expected"})

	t.Run("positive", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetName("expected")
		obj.SetNamespace("expected")

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("different name", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetName("different")
		obj.SetNamespace("expected")

		// when & then
		assert.False(t, pred.Matches(obj))
	})

	t.Run("different namespace", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetName("expected")
		obj.SetNamespace("different")

		// when & then
		assert.False(t, pred.Matches(obj))
	})
}

func TestLabelsPredicate(t *testing.T) {
	pred := Labels(map[string]string{"ka": "va", "kb": "vb"})

	t.Run("exact match", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetLabels(map[string]string{"ka": "va", "kb": "vb"})

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("subset match", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetLabels(map[string]string{"ka": "va", "kb": "vb", "kc": "vc"})

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("nil", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetLabels(nil)

		// when & then
		assert.False(t, pred.Matches(obj))
	})
}

func TestAnnotationsPredicate(t *testing.T) {
	pred := Annotations(map[string]string{"ka": "va", "kb": "vb"})

	t.Run("exact match", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetAnnotations(map[string]string{"ka": "va", "kb": "vb"})

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("subset match", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetAnnotations(map[string]string{"ka": "va", "kb": "vb", "kc": "vc"})

		// when & then
		assert.True(t, pred.Matches(obj))
	})

	t.Run("nil", func(t *testing.T) {
		// given
		obj := &corev1.ConfigMap{}
		obj.SetAnnotations(nil)

		// when & then
		assert.False(t, pred.Matches(obj))
	})
}
