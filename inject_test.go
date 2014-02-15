package inject

import (
	"fmt"
	"time"
	"reflect"
	"testing"
	"math/rand"
)

type SpecialString interface {
}

type TestStruct struct {
	Dep1 string        `inject`
	Dep2 SpecialString `inject`
	Dep3 string
}

func init() {
	rand.Seed(time.Now().Unix())
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected: %v (type %v) - Got: %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got: %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func Test_InjectorInvoke(t *testing.T) {
	injector := New()
	expect(t, injector == nil, false)

	dep := "some dependency"
	injector.Map(dep)
	dep2 := "another dep"
	injector.MapTo(dep2, (*SpecialString)(nil))

	_, err := injector.Invoke(func(d1 string, d2 SpecialString) {
		expect(t, d1, dep)
		expect(t, d2, dep2)
	})

	expect(t, err, nil)
}

func Test_InjectorInvokeReturnValues(t *testing.T) {
	injector := New()
	expect(t, injector == nil, false)

	dep := "some dependency"
	injector.Map(dep)
	dep2 := "another dep"
	injector.MapTo(dep2, (*SpecialString)(nil))

	result, err := injector.Invoke(func(d1 string, d2 SpecialString) string {
		expect(t, d1, dep)
		expect(t, d2, dep2)
		return "Hello world"
	})

	expect(t, err, nil)
	expect(t, result[0].String(), "Hello world")
}

func Test_InjectorApply(t *testing.T) {
	injector := New()

	injector.Map("a dep").MapTo("another dep", (*SpecialString)(nil))

	s := TestStruct{}
	err := injector.Apply(&s)
	expect(t, err, nil)

	expect(t, s.Dep1, "a dep")
	expect(t, s.Dep2, "another dep")
}

func Test_InterfaceOf(t *testing.T) {
	iType := InterfaceOf((*SpecialString)(nil))
	expect(t, iType.Kind(), reflect.Interface)

	iType = InterfaceOf((**SpecialString)(nil))
	expect(t, iType.Kind(), reflect.Interface)

	// Expecting nil
	defer func() {
		rec := recover()
		refute(t, rec, nil)
	}()
	iType = InterfaceOf((*testing.T)(nil))
}

func Test_InjectorGet(t *testing.T) {
	injector := New()

	injector.Map("some dependency")

	expect(t, injector.Get(reflect.TypeOf("string")).IsValid(), true)
	expect(t, injector.Get(reflect.TypeOf(11)).IsValid(), false)
}

func Test_InjectorSetParent(t *testing.T) {
	injector := New()
	injector.MapTo("another dep", (*SpecialString)(nil))

	injector2 := New()
	injector2.SetParent(injector)

	expect(t, injector2.Get(InterfaceOf((*SpecialString)(nil))).IsValid(), true)
}

func Test_InjectorInvokeFactory(t *testing.T) {
	injector := New()

	dep := "some dependency"
	injector.Map(func() string {
		return dep
	})
	dep2 := "another dep"
	injector.MapTo(func() string {
		return dep2
	}, (*SpecialString)(nil))

	res, err := injector.Invoke(func(d1 string, d2 SpecialString) string {
		expect(t, d1, dep)
		expect(t, d2, dep2)
		return dep
	})

	expect(t, err, nil)
	expect(t, res[0].String(), dep)
}

func Test_InjectorInvokeCascadingFactory(t *testing.T) {
	injector := New()

	answer := 42
	injector.Map(func() int {
		return answer
	})
	question := "What do you get if you multiply six by nine?"
	injector.Map(func(answer int) string {
		return fmt.Sprintf("%v %v", question, answer)
	})

	sentence := fmt.Sprintf("%v %v", question, answer)
	res, err := injector.Invoke(func(d1 string) string {
		expect(t, d1, sentence)
		return sentence
	})

	expect(t, err, nil)
	expect(t, res[0].String(), sentence)
}

func Test_InjectorInvokeDependencyLoop(t *testing.T) {
	injector := New()

	dep := "some dependency"
	injector.Map(func(d2 string) string {
		return dep
	})

	_, err := injector.Invoke(func(d string) {
		t.Errorf("expected an error, not %v", d)
	})

	if err == nil {
		t.Errorf("expected an error")
	}
}

func Test_InjectorInvokeWithParentDependency(t *testing.T) {
	injector := New()
	dep := "some dependency"
	injector.Map(func(d2 int) string {
		return dep
	})

	injector2 := New()
	injector2.Map(42)
	injector2.SetParent(injector)

	res, err := injector2.Invoke(func(d1 string) string {
		expect(t, d1, dep)
		return dep
	})

	expect(t, err, nil)
	expect(t, res[0].String(), dep)
}

func Test_InjectorInvokeCaching(t *testing.T) {
	injector := New()

	injector.Map(func() int {
		return rand.Intn(1000000)
	})
	injector.MapTo(func() string {
		return "!"
	}, (*SpecialString)(nil))
	injector.Map(func(c SpecialString, n int) string {
		return fmt.Sprintf("%v%v", n, c)
	})

	_, err := injector.Invoke(func(s string, c SpecialString, n int) {
		expect(t, s, fmt.Sprintf("%v%v", n, c))
	})

	expect(t, err, nil)
}
