{{- $tableNameSingular := .Table.Name | singular | titleCase -}}
{{- $varNameSingular := .Table.Name | singular | camelCase -}}

// MarshalJSONFilter calls the generic MarshalJSONFilter which will return a json object
// excludeFields will be matched to column names and removed
func (o *{{$tableNameSingular}}) MarshalJSONFilter(includeFields []string, excludeFields []string) ([]byte, error) {
    return marshal.MarshalJSONFilter(o, includeFields, excludeFields)
}

