{{- $tableNameSingular := .Table.Name | singular | titleCase -}}
{{- $varNameSingular := .Table.Name | singular | camelCase -}}

// MarshalJSONFilter calls the generic MarshalJSONFilter which will return a json object
// excludeFields will be matched to column names and removed
// "ID" is excluded by default
func (o *{{$tableNameSingular}}) MarshalJSONFilter(includeFields []string, excludeFields []string) ([]byte, error) {
    return marshal.MarshalJSONWrapper(o, includeFields, excludeFields)
}

// MarshalJSON will marshal the object into JSON and include all fields
func (o *{{$tableNameSingular}}) MarshalJSON() ([]byte, error) {
   return o.MarshalJSONFilter(nil, nil)
}

func (t *{{$tableNameSingular}}) JSONFilter(includeFields []string, excludeFields []string) (res map[string]interface{}, err error) {
	return marshal.MarshalJSONFilter(t, nil, nil)
}

func (o *{{$tableNameSingular}}) UnmarshalJSON(data []byte) error {
   return marshal.UnmarshalWrapper(o, data, nil)
}

