package marshal

import (
	"encoding/json"
	"errors"
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

// MarshalJSONStruct calls the JSONFilter with the appropriate filter lists and
// marshals the result IFF the given interface is a struct
func MarshalJSONStruct(o interface{}, exclude map[string]bool) (res []byte, err error) {
	val := reflect.ValueOf(o)
	val = reflect.Indirect(val) // For pointers
	if val.Kind() != reflect.Struct {
		return nil, errors.New("Invalid call to MarsahlJSONStruct on non-struct type")
	}

	j, err := JSONFilter(o, exclude)
	if err != nil {
		return nil, err
	}
	JSON, err := json.Marshal(j)
	return JSON, err
}

// MarshalJSONSlice calls the MarshalJSONStruct with the appropriate filter lists and
// marshals the result IFF the given interface is a slice of structs
func MarshalJSONSlice(o interface{}, exclude map[string]bool) (res []byte, err error) {
	val := reflect.ValueOf(o)
	val = reflect.Indirect(val) // For pointers
	if val.Kind() != reflect.Slice {
		return nil, errors.New("Invalid call to MarsahlJSONSlice on non-struct type")
	}

	result := []map[string]interface{}{}
	for i := 0; i < val.Len(); i++ {
		j, err := JSONFilter(val.Index(i).Interface(), exclude)
		if err != nil {
			return nil, err
		}
		result = append(result, j)
	}

	JSON, err := json.Marshal(result)
	return JSON, err
}

// JSONFilter iterates through an interface, constructs a map[string]interface{} with
// the appropriate fields according to the exclude map.
// Exclude is a map of field names and a boolean indicating whether or not to filter that field
// Filter fields of structs inside given interface with the syntax "structField.field"
func JSONFilter(o interface{}, exclude map[string]bool) (res map[string]interface{}, err error) {
	val := reflect.ValueOf(o)
	val = reflect.Indirect(val) // For pointers
	oType := reflect.TypeOf(val.Interface())
	// nextExclude is for filtering fields within structs within o
	nextExclude := GetNextExclude(exclude)

	result := make(map[string]interface{}, oType.NumField())

	for i := 0; i < oType.NumField(); i++ {
		field := oType.Field(i)
		fieldName := field.Name
		jsonKey, jsonValid := field.Tag.Lookup("json")

		// PkgPath != "" means it is an exported field, which we want to marshal
		if !jsonValid && !field.Anonymous && field.PkgPath != "" {
			continue
		}
		fieldValue := val.Field(i)

		// If the field is empty, skip it. TODO: Want a better check for differentiating empty arrays in some cases
		if IsEmpty(fieldValue.Interface()) {
			continue
		}

		keys, skip := ParseJSONKey(jsonKey, fieldName, fieldValue)
		if skip {
			continue
		}

		// Handle embedded field so the fields are in the proper layer in the JSON object
		if field.Anonymous {
			embeddedFields := make(map[string]interface{})

			// Handle go-ethereum structs & Curvegrid structs
			fieldType := fieldValue.Type()
			switch {
			case ethTransaction == fieldType:
				tx := handleTransaction(fieldValue.Interface().(*ethTypes.Transaction))
				embeddedFields, err = JSONFilter(tx, exclude)

			// Handle filtering an embedded field inside a struct
			case fieldValue.MethodByName("JSONFilter").IsValid():
				embeddedFields, err = fieldValue.Interface().(Filterable).JSONFilter(exclude)

			default:
				fieldValue = reflect.Indirect(fieldValue)
				embeddedFields, err = JSONFilter(fieldValue.Interface(), exclude)
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

		// if it's in our exclude list, skip it. Do this after checking for Anonymous field in case
		if exclude[fieldName] {
			continue
		}

		// If it's a struct that implements JSONFilter interface, we filter it here. Otherwise let the json.Marshal handle it.
		if fieldValue.MethodByName("JSONFilter").IsValid() {
			val, err := fieldValue.Interface().(Filterable).JSONFilter(nextExclude)
			if err != nil {
				return nil, err
			}
			result[keys[0]] = val
			continue
		}

		// Add field to result array which will be marshalled
		result[keys[0]] = fieldValue.Interface()
	}
	return result, nil
}

// UnmarshalWrapper handles unmarshaling JSON data into the o interface
// specialNames are needed in the case the user overrides marshalings default naming from BoilCase to SnakeCase
// specialNames are typed with SnakeCase (or custom name) as the key and BoilCase as the value
func UnmarshalWrapper(o interface{}, data []byte, specialNames map[string]string) error {
	var structFieldName string
	oValue := reflect.ValueOf(o).Elem()

	// Unmarshal to raw data so we can set individual fields
	var body map[string]*json.RawMessage
	err := json.Unmarshal(data, &body)
	if err != nil {
		return err
	}

	for i := 0; i < oValue.NumField(); i++ {
		field := oValue.Field(i)
		ft := oValue.Type().Field(i)
		// if there is an anonymous field, we need to initialize it before we can add elements to it
		if field.Kind() == reflect.Ptr && ft.Anonymous && field.IsNil() {
			newField := reflect.New(field.Type().Elem())
			field.Set(newField)
		}
	}

	for key, value := range body {
		// Check if that field name was overridden by the user and supplied in special names
		if name, named := specialNames[key]; named {
			key = name
		}
		structFieldName = ToBoilCase(key)
		SetIfAvailable(oValue, structFieldName, value)
	}
	return nil
}

// SetIfAvailable sets a field in a reflected object if possible
func SetIfAvailable(oValue reflect.Value, fieldName string, value *json.RawMessage) {
	if value != nil {
		fieldV := oValue.FieldByName(fieldName)
		fieldT, _ := oValue.Type().FieldByName(fieldName)

		if !fieldV.IsValid() {
			return
		}

		// if the field we found is an embedded field, then is has the same name as the struct
		if fieldT.Anonymous {
			SetIfAvailable(fieldV.Elem(), fieldName, value)
			return
		}
		// Create new reflect.Value from value
		newValue := reflect.New(fieldV.Type())
		json.Unmarshal(*value, newValue.Interface())
		newValue = newValue.Elem()

		// set field to new Value
		fieldV.Set(newValue)
	}
}

//////////////////////////////// HELPERS

// Filterable interface  is required for JSONFilter on structs that require special handlers ex. Blocks, Transactions
type Filterable interface {
	JSONFilter(map[string]bool) (map[string]interface{}, error)
}

// ParseJSONKey parses the jsonKey value of the field according to json.Marshals definition
func ParseJSONKey(jsonKey, fieldName string, fieldValue reflect.Value) ([]string, bool) {
	keys := strings.Split(jsonKey, ",")

	if jsonKey == "-," {
		keys[0] = "-"
	} else if len(keys) == 1 {
		if keys[0] == "-" {
			return nil, true
		} else if keys[0] == "omitempty" || keys[0] == "" {
			keys[0] = fieldName
		}
	} else if len(keys) == 2 {
		if keys[0] == "" && keys[1] == "omitempty" { // `json:",omitempty"`
			keys[0] = fieldName
		}
		if keys[1] == "omitempty" && IsEmpty(fieldValue.Interface()) {
			return nil, true
		}
	}
	keys[0] = ToSnakeCase(keys[0])
	return keys, false
}

// HandleTransaction is a special handler for go-ethereum transactions due to the hidden fields of a transaction struct
// and is used when marshaling a transaction
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

// GetNextExclude gets the next exclude map for a struct that is a field of a struct
// Exclude is a map of field names and a boolean indicating whether or not to filter that field
// The next exclude map is converting from format "structA.fieldB" = true to "fieldB" = true for
// when we're marshaling structA
func GetNextExclude(exclude map[string]bool) map[string]bool {
	next := map[string]bool{}
	for k, v := range exclude {
		if s := strings.SplitN(k, ".", 2); len(s) == 2 {
			next[s[1]] = v
		}
	}
	return next
}

// IsEmpty returns whether or not an interface is empty
// 0 int, "" string, false bool, empty struct...
func IsEmpty(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// For converting to SnakeCase
// http://www.golangprograms.com/golang-package-examples/golang-convert-string-into-snake-case.html
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase converts a string to snake case
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// ToBoilCase converts a string to the format sqlboiler uses for field names
func ToBoilCase(a string) (b string) {
	return strmangle.CamelCase(strmangle.TitleCase(a))
}
