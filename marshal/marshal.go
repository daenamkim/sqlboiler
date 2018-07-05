package marshal

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
)

// MarshalJSONFilter iterates through an interface, constructs a map[string]interface{} with
// the appropriate fields according to include & exclude.
// Excludes by column name when working with a model from sqlboiler, field name for go-ethereum.
// It currently excludes "ID" column by default.
func MarshalJSONFilter(o interface{}, includeFields []string, excludeFields []string) (res map[string]interface{}, err error) {
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
		fieldName := field.Name
		jsonKey, jsonValid := field.Tag.Lookup("json")
		// Check for a json tag
		// PkgPath != "" means it is an exported field, which we want to marshal
		if !jsonValid && !field.Anonymous && field.PkgPath != "" {
			continue
		}
		fieldValue := objValue.Field(i)
		keys := strings.Split(jsonKey, ",")

		// Handle json tag according to json.Marshal
		if jsonKey == "-," {
			keys[0] = "-"
		} else if len(keys) == 1 {
			if keys[0] == "-" {
				continue
			} else if keys[0] == "omitempty" || keys[0] == "" {
				keys[0] = fieldName
			}
		} else if len(keys) == 2 && keys[1] == "omitempty" && IsEmpty(fieldValue.Interface()) {
			continue
		}
		keys[0] = ToSnakeCase(keys[0])

		// Handle embedded field so the fields are in the proper layer in the JSON object
		if field.Anonymous {
			fieldValue = reflect.Indirect(fieldValue)
			embeddedFields, err := MarshalJSONFilter(fieldValue.Interface(), includeFields, excludeFields)
			if err != nil {
				return nil, err
			}
			for k, v := range embeddedFields {
				result[k] = v
			}
			continue
		}

		// if it's in our exclude list, skip it. Do this after checking for Anonymous field.
		if exclude[fieldName] {
			continue
		}

		// Add field to result array which will be marshalled
		result[keys[0]] = fieldValue.Interface()
	}
	return result, nil
}

// MarshalJSONWrapper calls the MarshalJSONFilter with the appropriate filter lists and
// marshals the result.
func MarshalJSONWrapper(o interface{}, includeFields []string, excludeFields []string) (res []byte, err error) {
	jason, err := MarshalJSONFilter(o, includeFields, excludeFields)
	if err != nil {
		return nil, err
	}
	JSON, err := json.Marshal(jason)
	return JSON, err
}

// GetTypeName strips the path from a type.Name() and returns the type name
func GetTypeName(typ reflect.Type) string {
	typeName := typ.Name()
	pkg := typ.PkgPath()
	if pkg == "" {
		return typeName
	}
	parts := strings.Split(pkg, "/")
	return parts[len(parts)-1] + "." + typeName
}

// IsEmpty returns whether or not an interface is empty
// 0 int, "" string, false bool, empty struct...
func IsEmpty(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// http://www.golangprograms.com/golang-package-examples/golang-convert-string-into-snake-case.html
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func SetIfAvailable(oNew reflect.Value, field string, value *string) {
	if !(value == nil) {
		reflectedField := oNew.Elem().FieldByName(field)
		reflectPointer := reflect.New(reflectedField.Type())
		method := reflectPointer.MethodByName("Scan")
		if method.IsValid() {
			params := []reflect.Value{reflect.ValueOf(*value)}
			method.Call(params)
			reflectedField.Set(reflectPointer.Elem())
		} else {
			reflectedField.Set(reflect.ValueOf(*value))
		}
	}
}

func UnmarshalWrapper(o interface{}, data []byte) {
	// body := map[string]*string
	var structFieldName string
	oType := reflect.TypeOf(o)
	oNew := reflect.New(oType)
	for key, value := range body {
		structFieldName = BoilCase(key)
		SetIfAvailable(oNew, structFieldName, value)
	}
}

// func BoilCase(a string) (b string) {
// 	return strmangle.CamelCase(strmangle.TitleCase(a))
// }
