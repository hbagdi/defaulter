package defaulter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	assert := assert.New(t)

	type Foo struct {
		String  string
		Int     int
		IntP    *int
		IntA    []int
		StringA []*string

		IntKeepEmpty    int
		IntDoNotDefault int

		unexportedInt *int
	}

	Int := 442
	stringA := "stringA"
	stringB := "stringB"
	defaultFoo := Foo{
		String:  "defaultString",
		Int:     42,
		IntP:    &Int,
		IntA:    []int{1, 2, 3},
		StringA: []*string{&stringA, &stringB},

		IntDoNotDefault: 777,
	}

	var arg Foo
	arg.IntDoNotDefault = -42
	err := Set(&arg, defaultFoo)
	assert.Nil(err)
	assert.Equal(defaultFoo.String, arg.String)
	assert.Equal(defaultFoo.Int, arg.Int)
	// values pointed via the pointers are same
	assert.Equal(defaultFoo.IntP, arg.IntP)
	// while pointers themselves are not same
	if defaultFoo.IntP == arg.IntP {
		assert.Fail("Expected IntP pointers to point to different addresses")
	}

	// Arrays are same
	assert.Equal(defaultFoo.IntA, arg.IntA)

	// Field not present in def is remains the same
	assert.Equal(0, arg.IntKeepEmpty)
	// field set in the arg is not overridden
	assert.Equal(-42, arg.IntDoNotDefault)
	// unexported values are not touched
	assert.Nil(arg.unexportedInt)

	// values pointed by the pointers are equal
	assert.Equal(defaultFoo.StringA, arg.StringA)
	// but the pointers are not
	for i := range defaultFoo.StringA {
		if defaultFoo.StringA[i] == arg.StringA[i] {
			assert.Fail("Expected pointers in the array" +
				" to point to different addresses")
		}
	}
}

func TestSanityChecks(t *testing.T) {
	assert := assert.New(t)
	type Foo struct {
		String string
		Int    int
	}
	type Bar struct {
		String string
		Int    int
	}

	assert.Equal(errNilArgs, Set(nil, nil))
	assert.Equal(errNilArgs, Set(nil, Foo{}))
	assert.Equal(errNilArgs, Set(&Foo{}, nil))
	assert.Equal(errArgpNotPointer, Set(Foo{}, Foo{}))
	assert.Equal(errNotSameKind, Set(&Foo{}, Bar{}))
}

func TestMap(t *testing.T) {
	assert := assert.New(t)
	type Foo struct {
		Map map[string]*Foo
	}

	m := make(map[string]*Foo)
	m["key1"] = &Foo{
		Map: make(map[string]*Foo),
	}
	defaultFoo := Foo{
		Map: m,
	}
	var arg Foo

	err := Set(&arg, defaultFoo)
	assert.Nil(err)
	assert.Equal(1, len(arg.Map))
	assert.Equal(defaultFoo.Map, arg.Map)
	if arg.Map["key1"] == defaultFoo.Map["key1"] {
		assert.Fail("Expected pointers in the map" +
			" to point to different addresses")
	}
	assert.Nil(err)
}

func TestStructSlice(t *testing.T) {
	assert := assert.New(t)

	type Bar struct {
		String  string
		StringP *string
	}
	type Foo struct {
		String string
		Slice  []*Bar
	}
	s := "stringPValue"
	defaultFoo := Foo{
		Slice: []*Bar{
			{String: "1", StringP: &s},
		},
	}
	var arg Foo

	err := Set(&arg, defaultFoo)
	assert.Nil(err)
	assert.Equal(defaultFoo, arg)
	assert.Nil(err)
	if arg.Slice[0] == defaultFoo.Slice[0] {
		assert.Fail("Expected pointers in the slice" +
			" to point to different addresses")
	}

	if arg.Slice[0].StringP == defaultFoo.Slice[0].StringP {
		assert.Fail("Expected pointers in the slice" +
			" to point to different addresses")
	}
}

func TestNestedStruct(t *testing.T) {
	assert := assert.New(t)

	type Bar struct {
		String string
		Int    int
		IntP   *int
	}
	type Foo struct {
		String  string
		Int     int
		IntP    *int
		IntA    []int
		StringA []*string
		Bar     Bar
		BarP    *Bar
	}

	Int := 442
	stringA := "stringA"
	stringB := "stringB"
	defaultFoo := Foo{
		String:  "defaultString",
		Int:     42,
		IntP:    &Int,
		IntA:    []int{1, 2, 3},
		StringA: []*string{&stringA, &stringB},
	}

	var arg Foo

	err := Set(&arg, defaultFoo)
	assert.Nil(err)
	assert.Nil(arg.BarP)

	arg = Foo{}
	defaultFoo.Bar.String = "defaultFoo.Bar.StringValue"
	err = Set(&arg, defaultFoo)
	assert.Nil(err)

	assert.Equal(defaultFoo.Bar.String, arg.Bar.String)
	assert.Empty(defaultFoo.Bar.IntP)

	defaultFoo.BarP = &Bar{}
	arg = Foo{}
	err = Set(&arg, defaultFoo)
	assert.Nil(err)
	assert.Equal(defaultFoo.BarP, arg.BarP)
}
