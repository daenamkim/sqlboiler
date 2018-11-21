{{- $tableNameSingular := .Table.Name | singular | titleCase -}}
{{- $varNameSingular := .Table.Name | singular | camelCase -}}

// MarshalJSON will marshal the struct {{$tableNameSingular}} into JSON
// The ID field is filtered by default
// This is called by default through json.Marshal
func (o {{$tableNameSingular}}) MarshalJSON() ([]byte, error) {
    exclude := map[string]bool{"ID": true}
    return marshal.ToJSON(o, exclude)
 }

 // MarshalJSONFilter will marshal the struct {{$tableNameSingular}} into JSON
// The exclude map will be filtered
func (o {{$tableNameSingular}}) MarshalJSONFilter(exclude map[string]bool) ([]byte, error) {
    temp, err := o.JSONFilter(exclude)
    if err != nil {
        return nil, err
    }
    return json.Marshal(temp)
 }

// JSONFilter is required to be able to filter fields in this struct when it is an anonymous field in another struct
// if there is a special case in the struct (private fields for example)
func (o {{$tableNameSingular}}) JSONFilter(exclude map[string]bool) (res map[string]interface{}, err error) {
	return marshal.JSONFilter(o, exclude)
}
