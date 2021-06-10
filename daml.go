package daml

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

func stripPackage(s string) string {
	split := strings.Split(s, ".")
	if len(split) == 2 {
		return split[1]
	}
	return s
}

func damlPrimitiveTypeMap() map[string]string {
	return map[string]string{
		"string":  "Text",
		"int":     "Int",
		"int32":   "Int",
		"int64":   "Int",
		"float":   "Decimal",
		"float32": "Decimal",
		"float64": "Decimal",
		"bool":    "Bool",
		"Date":    "Date",
	}
}

func toDAMLType(s string) string {
	t, _ := toDAMLTypeWithBool(s)
	return t
}

func toDAMLTypeWithBool(s string) (string, bool) {
	pt := damlPrimitiveTypeMap()
	t := pt[s]
	if len(t) > 0 {
		return t, true
	}
	return s, false
}

type marshalDAMLMap struct {
	Type     interface{}
	TypeName string
	Seen     map[string][]byte
}

func defaultPrimitiveTypes() map[string][]byte {
	return map[string][]byte{
		"string":  []byte{},
		"int":     []byte{},
		"int32":   []byte{},
		"int64":   []byte{},
		"float":   []byte{},
		"float32": []byte{},
		"float64": []byte{},
		"bool":    []byte{},
	}
}

// Marshal : convert given type to DAML type
func Marshal(x interface{}, typeMap func() map[string]interface{}) ([]byte, error) {
	r := reflect.ValueOf(x)
	typeOfS := r.Type()
	log.Printf("typeOfS: %s", typeOfS)

	var buf bytes.Buffer
	baseTypeString := stripPackage(fmt.Sprintf("%s", typeOfS))
	if typeOfS.String() != "daml.marshalDAMLMap" {
		// buf.WriteString(fmt.Sprintf("module %s where\n\n", strings.ToUpper(baseTypeString)))
		buf.WriteString(fmt.Sprintf("module %s where\n\n", baseTypeString))
	}

	seenTypes := make(map[string][]byte)
	// primitiveTypes := defaultPrimitiveTypes()

	if typeOfS == reflect.ValueOf(marshalDAMLMap{}).Type() {
		seenTypes = x.(marshalDAMLMap).Seen
		r = reflect.ValueOf(x.(marshalDAMLMap).Type)
		if !r.IsValid() {
			log.Printf("\n\nINVALID!\n\n")
			return buf.Bytes(), errors.New(fmt.Sprintf("invalid type. '%s' likely does not exist in provided typeMap", r))
		}
		typeOfS = r.Type()
		log.Printf("[changed] typeOfS: %s", typeOfS)
		baseTypeString = stripPackage(fmt.Sprintf("%s", typeOfS))
	}
	enum := r.MethodByName("Enum")
	if enum.IsValid() {
		// if enum {
		// enum.Call()
		log.Printf("Is enum")
		rCalled := r.MethodByName("EnumOptions").Call([]reflect.Value{})
		enumValues := baseTypeString
		if len(rCalled) > 0 {
			v := rCalled[0]
			// enumValues = strings.Join(v.Interface().([]string), " | ")
			var enumsValsNoSpaces []string
			for _, val := range v.Interface().([]string) {
				cleanedVal := strings.Join(strings.Split(val, " "), "")
				enumsValsNoSpaces = append(enumsValsNoSpaces, cleanedVal)
			}
			enumValues = strings.Join(enumsValsNoSpaces, " | ")
		}
		if len(enumValues) == 0 {
			enumValues = "Text"
			buf.WriteString(fmt.Sprintf("type %v = %v\n\n", baseTypeString, enumValues))
			return buf.Bytes(), nil
		}
		// strings.Join(, )
		buf.WriteString(fmt.Sprintf("data %v = %v\n", baseTypeString, enumValues))
		buf.WriteString(fmt.Sprintf("  deriving(Eq, Show)\n\n"))
		return buf.Bytes(), nil
	} else if fmt.Sprintf("%s", typeOfS.Kind()) == "string" {
		buf.WriteString(fmt.Sprintf("type %v = %v\n\n", baseTypeString, "Text"))
		return buf.Bytes(), nil
	}
	log.Printf("underlying type? --> %s", typeOfS.Kind())

	// if typeOfS.Elem().AssignableTo(reflect.TypeOf("")) {
	// 	log.Fatalf("STRINGGGGGG")
	// }

	log.Printf("data %v = %v\n", baseTypeString, baseTypeString)
	buf.WriteString(fmt.Sprintf("data %v = %v\n", baseTypeString, baseTypeString))
	// fmt.Printf("  with\n")
	buf.WriteString(fmt.Sprintf("  with\n"))

	// for i := 0; i < r.NumField(); i++ {
	// 	useType := stripPackage(fmt.Sprintf("%s", typeOfS.Field(i).Type))
	// 	log.Printf("%#v", toDAMLType(useType))
	// 	fmt.Printf("    Field: %s\tValue: %v - %v\n", typeOfS.Field(i).Name, r.Field(i).Interface())
	// }
	// seenTypes := make(map[string]interface{})
	// seenTypes := make(map[string]bool)
	renderTypes := []string{}
	// toExecute := make(map[string]func(interface{}) []byte)
	log.Printf("interface type: %v", r.CanInterface())
	for i := 0; i < r.NumField(); i++ {
		// log.Printf("seenTypes: %v", seenTypes)
		field := typeOfS.Field(i)
		log.Println(field.PkgPath)
		currentType := typeOfS.Field(i).Type
		isList := strings.HasPrefix(fmt.Sprintf("%s", currentType), "[]")
		trimmedCurrentType := strings.TrimPrefix(fmt.Sprintf("%s", currentType), "[]")
		typeString := stripPackage(trimmedCurrentType)
		damlType, basicType := toDAMLTypeWithBool(typeString)
		if seenTypes[typeString] == nil && !basicType {
			log.Printf("Current type: %s", currentType)
			log.Printf("Base type: %s", typeString)
			if typeString == "rtype" {
				log.Fatalf("Stop.")
			}
			renderTypes = append(renderTypes, typeString)
			log.Printf("renderTypes: %s", renderTypes)
			// log.Fatalf("Stop")
			var err error
			log.Printf("given type for %s = %#v", typeString, typeMap()[typeString])
			seenTypes[typeString], err = Marshal(marshalDAMLMap{Type: typeMap()[typeString], Seen: seenTypes}, typeMap)
			if err != nil {
				log.Fatalf("ERROR: %#v", err)
				return buf.Bytes(), err
			}
			// toExecute[typeString] = func(x interface{}) []byte {
			// 	return MarshallDAML(x)
			// }
		}
		// fmt.Printf("    %v : %v\n", typeOfS.Field(i).Name, toDAMLType(useType))
		typeFieldName := typeOfS.Field(i).Name
		tag := field.Tag.Get("daml")
		required := strings.Contains(field.Tag.Get("validate"), "required")
		var optionalText string
		if !required {
			optionalText = "Optional "
		}
		if tag == "" && strings.Index(string(field.Tag), ":") < 0 {
			tag = string(field.Tag)
		}
		if tag == "-" {
			continue
		} else if tag == "" {
			tag = strings.ToLower(typeFieldName)
			if len(tag) > 1 {
				tag = string(tag[0]) + typeFieldName[1:]
			}
		}

		log.Printf("Field Name: %s | Tag: %s", typeFieldName, tag)
		if isList {
			// buf.WriteString(fmt.Sprintf("    %v : [%v]\n", typeFieldName, damlType))
			buf.WriteString(fmt.Sprintf("    %v : %s[%v]\n", tag, optionalText, damlType))
		} else {
			// buf.WriteString(fmt.Sprintf("    %v : %v\n", typeFieldName, damlType))
			buf.WriteString(fmt.Sprintf("    %v : %s%v\n", tag, optionalText, damlType))
		}
	}
	// fmt.Printf("  deriving(Eq, Show)\n")
	buf.WriteString(fmt.Sprintf("  deriving(Eq, Show)\n"))
	buf.WriteString(fmt.Sprintf("\n"))
	// fmt.Println("")

	for _, t := range renderTypes {
		log.Printf("Processing: %s", t)
		// buf.WriteString(string(MarshallDAML(seenTypes[t])))
		buf.WriteString(string(seenTypes[t]))
	}

	return buf.Bytes(), nil
}

func toDAML(x interface{}, typeMap func() map[string]interface{}) {
	d, _ := Marshal(x, typeMap)
	fmt.Printf("\n\n%s\n\n", string(d))
}
