package daml

import (
	"io/ioutil"
	"reflect"
	"testing"
)

type CustomStruct struct {
	Name       string
	SomeField  string `daml:"some_field" validate:"required"`
	OmitMe     string `daml:"-"`
	IncludeMe  bool   `daml:"include_me"`
	SomeStruct AnotherStruct
	List       []SliceMe
}

type AnotherStruct struct {
	Foo       string
	Bar       int
	EnumThing EnumExample
}

type SliceMe struct {
	Value bool
}

type EnumExample string

type Foobar struct {
	A string
	B []string
}

const (
	Option1 EnumExample = "A"
	Option2 EnumExample = "B"
)

func (fs EnumExample) Enum() []EnumExample {
	return []EnumExample{
		Option1,
		Option2,
	}
}

func (fs EnumExample) EnumOptions() []string {
	return []string{
		string(Option1),
		string(Option2),
	}
}

func typeMap() map[string]interface{} {
	return map[string]interface{}{
		"CustomStruct":  CustomStruct{},
		"AnotherStruct": AnotherStruct{},
		"SliceMe":       SliceMe{},
		"EnumExample":   EnumExample(""),
	}
}

var exampleDAML = `module CustomStruct where

data CustomStruct = CustomStruct
  with
    name : Optional Text
    some_field : Text
    include_me : Optional Bool
    someStruct : Optional AnotherStruct
    list : Optional [SliceMe]
  deriving(Eq, Show)

data AnotherStruct = AnotherStruct
  with
    foo : Optional Text
    bar : Optional Int
    enumThing : Optional EnumExample
  deriving(Eq, Show)

data EnumExample = A | B
  deriving(Eq, Show)

data SliceMe = SliceMe
  with
    value : Optional Bool
  deriving(Eq, Show)

`

func TestMarshal(t *testing.T) {
	type args struct {
		x       interface{}
		typeMap func() map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "multiTypeTest",
			args: args{
				x:       CustomStruct{},
				typeMap: typeMap,
			},
			want:    []byte(exampleDAML),
			wantErr: false,
		},
		// // Need to find different failure condition
		// {
		// 	name: "failureTest",
		// 	args: args{
		// 		x:       Foobar{},
		// 		typeMap: typeMap,
		// 	},
		// 	want:    []byte(""),
		// 	wantErr: true,
		// },
		// // This needs to be reworked
		// {
		// 	name: "panicTest",
		// 	args: args{
		// 		x:       map[string]interface{}{},
		// 		typeMap: typeMap,
		// 	},
		// 	want:    []byte(""),
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args.x, tt.args.typeMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				have := []byte(got)
				_ = ioutil.WriteFile("have.daml", have, 0644)
				want := []byte(tt.want)
				_ = ioutil.WriteFile("want.daml", want, 0644)

				t.Errorf("Marshal() = %v, want %v", got, tt.want)
				// fmt.Println(string(got))
				// fmt.Printf("\n\n\n\n")
				// fmt.Println(string(tt.want))
				t.Errorf("Marshal() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
