{{- $tableNameSingular := .Table.Name | singular | titleCase -}}
{{- $varNameSingular := .Table.Name | singular | camelCase -}}

// MarshalJSON will marshal the struct {{$tableNameSingular}} into JSON
// The ID field is filtered by default
// This is called by default through json.Marshal
func (o *{{$tableNameSingular}}) MarshalJSON() ([]byte, error) {
    return marshal.MarshalJSONStruct(o, map[string]bool{"ID": true})
 }

// MarshalJSONStruct calls the generic MarshalJSONStruct which will return a json object
// excludeFields will be matched to column names and removed
func (o *{{$tableNameSingular}}) MarshalJSONStruct(exclude map[string]bool) ([]byte, error) {
    return marshal.MarshalJSONStruct(o, exclude)
}

// MarshalJSONSlice will call the generic MarshalJSONSlice which will return a json object
// excludeFields will be matched to column names and removed from each element of the slice
func (o *{{$tableNameSingular}}Slice) MarshalJSONSlice(exclude map[string]bool) ([]byte, error) {
    return marshal.MarshalJSONSlice(o, exclude)
}

// JSONFilter is required to be able to filter this struct when it is an anonymous field in another struct
// if there is a special case in the struct (private fields for example)
func (t *{{$tableNameSingular}}) JSONFilter(exclude map[string]bool) (res map[string]interface{}, err error) {
	return marshal.JSONFilter(t, exclude)
}

// UnmarshalJSON will unmarshal the JSON data into the struct {{$tableNameSingular}}
func (o *{{$tableNameSingular}}) UnmarshalJSON(data []byte) error {
    return marshal.UnmarshalWrapper(o, data, nil)
 }
