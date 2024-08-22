package assertions

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Predicate is a generic predicate for testing whether some object of type T has some quality.
// It is best used with the `AssertThat` function or with the `wait.For(...).FirstThat(...)`
// function in end to end tests using the Is function as a helper to satisfy the method
// signatures and help Go's type inference.
//
// Note that if you're implementing your own predicates, it is helpful for the constructor
// function to not return a concrete type but the generic Predicate[YourObjectType]. This
// helps the Go compiler to be able to infer and match up the types correctly.
//
// E.g. if one would want to implement a predicate checking that a ToolchainCluster CR has
// the ready condition checked, one might implement a constructor function for that predicate
// like this:
//
//	type toochainClusterReady struct {}
//
//	func (t *toolchainClusterReady) Matches(c *toolchainv1alpha1.ToolchainCluster) bool {
//	  return condition.IsTrue(c.Status.Conditions, toolchainv1alpha1.ConditionReady)
//	}
//
//	func Ready() predicates.Predicate[*toolchainv1alpha1.ToolchainCluster] {
//	  return &toolchainClusterReady{}
//	}
//
// Such predicate can then easily be used with the `AssertThat` function (or
// `wait.For(...).FirstThat(...)` from toolchain-e2e which does something very similar
// but waits for an object that satisfies the predicates to appear in the cluster).
//
//	assertions.AssertThat(t, toolchainCluster, assertions.Is(Ready()))
//
// If you're implementing your own predicate, consider implementing the PredicateMatchFixer,
// too, so that you can benefit from improved failure diagnostics offered by Explain function.
type Predicate[T client.Object] interface {
	Matches(obj T) bool
}

// PredicateMatchFixer is an optional interface that the predicate implementations can also
// implement. If so, the FixToMatch method is used to obtain an object that WOULD
// match the predicate. This would-be-matching object is then used to produce a diff
// between it and the non-matching object of the predicate in case of a test failure
// for logging purposes.
//
// There is no need to copy the provided object.
type PredicateMatchFixer[T client.Object] interface {
	FixToMatch(obj T) T
}

// Is merely casts the generic predicate on type T to a predicate on client.Object. This is
// always valid because T is required to implement client.Object. Using this function helps
// readability of the code by being able to construct expressions like:
//
//	predicates.Is(predicates.Named("whatevs"))
func Is[T client.Object](p Predicate[T]) Predicate[client.Object] {
	if fp, ok := p.(FixingPredicate[T]); ok {
		return &fixingCast[T]{Inner: fp}
	}
	return &cast[T]{Inner: p}
}

// Has is just an alias of Is. It is provided for better readability with certain predicate
// names.
func Has[T client.Object](p Predicate[T]) Predicate[client.Object] {
	if fp, ok := p.(FixingPredicate[T]); ok {
		return &fixingCast[T]{Inner: fp}
	}
	return &cast[T]{Inner: p}
}

type FixingPredicate[T client.Object] interface {
	Predicate[T]
	PredicateMatchFixer[T]
}

type cast[T client.Object] struct {
	// Inner is public so that Explain (in assertions.go) can access it...
	Inner Predicate[T]
}

type fixingCast[T client.Object] struct {
	// Inner is public so that Explain (in assertions.go) can access it...
	Inner FixingPredicate[T]
}

var (
	_ Predicate[client.Object]           = (*cast[client.Object])(nil)
	_ Predicate[client.Object]           = (*fixingCast[client.Object])(nil)
	_ PredicateMatchFixer[client.Object] = (*fixingCast[client.Object])(nil)
)

func (c *cast[T]) Matches(obj client.Object) bool {
	return c.Inner.Matches(obj.(T))
}

func (c *fixingCast[T]) Matches(obj client.Object) bool {
	return c.Inner.Matches(obj.(T))
}

func (c *fixingCast[T]) FixToMatch(obj client.Object) client.Object {
	return c.Inner.FixToMatch(obj.(T))
}

type named struct {
	name string
}

var (
	_ Predicate[client.Object]           = (*named)(nil)
	_ PredicateMatchFixer[client.Object] = (*named)(nil)
)

func (n *named) Matches(obj client.Object) bool {
	return obj.GetName() == n.name
}

func (n *named) FixToMatch(obj client.Object) client.Object {
	obj.SetName(n.name)
	return obj
}

// Name returns a predicate checking that an Object has given name.
func Name(name string) Predicate[client.Object] {
	return &named{name: name}
}

type inNamespace struct {
	namespace string
}

var (
	_ Predicate[client.Object]           = (*inNamespace)(nil)
	_ PredicateMatchFixer[client.Object] = (*inNamespace)(nil)
)

func (i *inNamespace) Matches(obj client.Object) bool {
	return obj.GetNamespace() == i.namespace
}

func (i *inNamespace) FixToMatch(obj client.Object) client.Object {
	obj.SetNamespace(i.namespace)
	return obj
}

// InNamespace returns a predicate checking that an Object is in the given namespace.
func InNamespace(name string) Predicate[client.Object] {
	return &inNamespace{namespace: name}
}

type withKey struct {
	types.NamespacedName
}

var (
	_ Predicate[client.Object]           = (*withKey)(nil)
	_ PredicateMatchFixer[client.Object] = (*withKey)(nil)
)

func (w *withKey) Matches(obj client.Object) bool {
	return obj.GetName() == w.Name && obj.GetNamespace() == w.Namespace
}

func (w *withKey) FixToMatch(obj client.Object) client.Object {
	obj.SetName(w.Name)
	obj.SetNamespace(w.Namespace)
	return obj
}

// ObjectKey returns a predicate checking that an Object has given NamespacedName (aka client.ObjectKey).
func ObjectKey(key types.NamespacedName) Predicate[client.Object] {
	return &withKey{NamespacedName: key}
}

type hasLabels struct {
	requiredLabels map[string]string
}

var (
	_ Predicate[client.Object]           = (*hasLabels)(nil)
	_ PredicateMatchFixer[client.Object] = (*hasLabels)(nil)
)

func (h *hasLabels) Matches(obj client.Object) bool {
	objLabels := obj.GetLabels()
	for k, v := range h.requiredLabels {
		value, present := objLabels[k]
		if !present || value != v {
			return false
		}
	}
	return true
}

func (h *hasLabels) FixToMatch(obj client.Object) client.Object {
	if len(h.requiredLabels) == 0 {
		return obj
	}
	objLabels := obj.GetLabels()
	if objLabels == nil {
		objLabels = map[string]string{}
	}

	for k, v := range h.requiredLabels {
		objLabels[k] = v
	}

	obj.SetLabels(objLabels)
	return obj
}

// Labels returns a predicate checking that an Object has provided labels and their values.
func Labels(requiredLabels map[string]string) Predicate[client.Object] {
	return &hasLabels{requiredLabels: requiredLabels}
}

type hasAnnotations struct {
	requiredAnnotations map[string]string
}

var (
	_ Predicate[client.Object]           = (*hasAnnotations)(nil)
	_ PredicateMatchFixer[client.Object] = (*hasAnnotations)(nil)
)

func (h *hasAnnotations) Matches(obj client.Object) bool {
	objAnnos := obj.GetAnnotations()
	for k, v := range h.requiredAnnotations {
		value, present := objAnnos[k]
		if !present || value != v {
			return false
		}
	}
	return true
}

func (h *hasAnnotations) FixToMatch(obj client.Object) client.Object {
	if len(h.requiredAnnotations) == 0 {
		return obj
	}
	objAnnos := obj.GetAnnotations()
	if objAnnos == nil {
		objAnnos = map[string]string{}
	}

	for k, v := range h.requiredAnnotations {
		objAnnos[k] = v
	}

	obj.SetAnnotations(objAnnos)
	return obj
}

// Annotations returns a predicate checking that an Object has provided annotations and their values.
func Annotations(requiredAnnotations map[string]string) Predicate[client.Object] {
	return &hasAnnotations{requiredAnnotations: requiredAnnotations}
}
