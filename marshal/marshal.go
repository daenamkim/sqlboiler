package marshal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/curvegrid/sqlboiler/strmangle"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

var (
	ethTransaction = reflect.TypeOf((*ethTypes.Transaction)(nil))
	ethBlock       = reflect.TypeOf((*ethTypes.Block)(nil))
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
		if IsEmpty(fieldValue.Interface()) {
			continue
		}

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
		fmt.Printf("%+v %+v %+v %+v\n", fieldName, fieldValue.Kind(), fieldValue.Type(), field.PkgPath)
		// TODO: Double check it works for embedded fields that aren't pointers
		if field.Anonymous {
			embeddedFields := make(map[string]interface{})

			// Handle go-ethereum structs
			fieldType := fieldValue.Type()
			switch {
			case ethTransaction == fieldType:
				// fieldValue = reflect.Indirect(fieldValue)
				tx := handleTransaction(fieldValue.Interface().(*ethTypes.Transaction))
				embeddedFields, err = MarshalJSONFilter(tx, includeFields, excludeFields)
			case fieldValue.MethodByName("JSONFilter").IsValid(): // OR CHECK IF IT IMPLEMENTS FILTER INTERFACE ?
				embeddedFields, err = fieldValue.Interface().(JSONFilter).JSONFilter(nil, nil)
			default:
				// only indirect if it's a pointer
				fieldValue = reflect.Indirect(fieldValue)
				embeddedFields, err = MarshalJSONFilter(fieldValue.Interface(), includeFields, excludeFields)
			}
			// Add embedded fields to the map that will be marshaled
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

func ToCamelCase(a string) (b string) {
	return strmangle.CamelCase(strmangle.TitleCase(a))
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

func UnmarshalWrapper(o interface{}, body map[string]*string) {
	var structFieldName string
	oValue := reflect.ValueOf(o)
	for key, value := range body {
		structFieldName = ToCamelCase(key)
		SetIfAvailable(oValue, structFieldName, value)
	}
}

func handleTransaction(tx *ethTypes.Transaction) interface{} {
	v, r, s := tx.RawSignatureValues()
	res := &struct {
		Nonce    hexutil.Uint64
		GasPrice *hexutil.Big
		Gas      hexutil.Uint64
		To       *common.Address
		Value    *hexutil.Big
		Input    hexutil.Bytes
		V        *hexutil.Big
		R        *hexutil.Big
		S        *hexutil.Big
		Hash     common.Hash
	}{
		Nonce:    hexutil.Uint64(tx.Nonce()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Gas:      hexutil.Uint64(tx.Gas()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		Input:    tx.Data(),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
		Hash:     tx.Hash(),
	}

	return res
}

type JSONFilter interface {
	JSONFilter([]string, []string) (map[string]interface{}, error)
}
