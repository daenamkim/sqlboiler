package marshal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func TestFunction() {
	fmt.Printf("\n\nTESTING 2\n\n")
}

func MarshalJSONFilter(o interface{}, includeFields []string, excludeFields []string) (res []byte, err error) {
	objValue := reflect.ValueOf(o)
	objValue = reflect.Indirect(objValue) // For pointers
	objType := reflect.TypeOf(objValue.Interface())

	result := make(map[string]interface{}, objType.NumField())
	exclude := map[string]bool{"ID": true}
	for _, f := range excludeFields {
		exclude[f] = true
	}
	for _, f := range includeFields {
		exclude[f] = false
	}

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		jsonKey, jsonValid := field.Tag.Lookup("json")
		if !jsonValid {
			continue
		}
		keys := strings.Split(jsonKey, ",")
		if len(keys) == 1 && keys[0] == "-" {
			continue
		}

		fieldValue := objValue.Field(i)
		fieldValue = reflect.Indirect(fieldValue)
		// fmt.Printf("\nVALUE 2 %v", fieldValue)
		// fieldType := field.Type
		// if fieldValue.Kind() == reflect.Ptr {
		// 	fieldType = reflect.TypeOf(fieldValue.Interface()) // TODO: Ensure this won't panic and cause issues. It shouldn't on pointers
		// }
		fieldName := field.Name

		// fmt.Printf("\nFIELD: %+v -- %v", field.Type == null.String, field.Type == "null.String")
		// fmt.Printf("\name %v", fieldName)

		// if it's in our exclude list, skip it
		if exclude[fieldName] {
			continue
		}

		// Special handlers per struct
		if fieldValue.Kind() == reflect.Struct {
			typeName := getTypeName(fieldValue.Type())
			switch typeName {
			case "null.v6.String", "null.v6.Int", "time.Time":
				result[keys[0]] = fieldValue.Interface()
			default:
				result[keys[0]], err = MarshalJSONFilter(fieldValue.Interface(), includeFields, excludeFields)
			}
		} else {
			result[jsonKey] = fieldValue.Interface()
		}

		// filter include from exclude?
		// if it's in exclude, skip
		// alter name
		// if it's leaf -> json.marshal(leaf)
		// if it's branch -> marshalJSONWrapper(branch)
		// switch fieldType.Kind(){
		// case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		// reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		// reflect.Float32, reflect.Float64, reflect.String:
		// 	// do shit
		// case reflect.Interface, reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Ptr:

		// default:

		// }

		// Handle null.String & null.Int

	}
	return json.Marshal(result)
}

func getTypeName(typ reflect.Type) string {
	typeName := typ.Name()
	pkg := typ.PkgPath()
	if pkg == "" {
		return typeName
	}
	parts := strings.Split(pkg, "/")
	return parts[len(parts)-1] + "." + typeName
}
