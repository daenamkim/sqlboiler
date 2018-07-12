{{- $tableNameSingular := .Table.Name | singular | titleCase -}}
{{- $tableNamePlural := .Table.Name | plural | titleCase -}}
{{- $varNamePlural := .Table.Name | plural | camelCase -}}
{{- $varNameSingular := .Table.Name | singular | camelCase -}}


func test{{$tableNamePlural}}Marshal(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	{{$varNameSingular}} := &{{$tableNameSingular}}{}
	if err = randomize.Struct(seed, {{$varNameSingular}}, {{$varNameSingular}}DBTypes, true, {{$varNameSingular}}ColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize {{$tableNameSingular}} struct: %s", err)
	}

	// Marshal object 
	var marshaled []byte
	if marshaled, err = json.Marshal({{$varNameSingular}}); err != nil {
		t.Errorf("Unable to marshal struct: %s", err)
	}

	// Check if default filtering ID worked 
	var body map[string]*json.RawMessage
	err = json.Unmarshal(marshaled, &body)
	if err != nil {
		t.Errorf("Couldn't unmarshal []byte to RawMessage: %s", err)
	}
	if _, valid := body["id"]; valid {
		t.Errorf("Default ID wasn't filtered: %s", err)
	}

	////////////////////// CHECK UNMARSHAL

	// Check filter
	includeFields := []string{"ID"}
	var marshaled2 []byte
	if marshaled2, err = {{$varNameSingular}}.MarshalJSONFilter(includeFields, nil); err != nil {
		t.Errorf("Unable to marshal struct: %s", err)
	}

	// Check if adding ID back works
	var body2 map[string]*json.RawMessage
	err = json.Unmarshal(marshaled2, &body2)
	if err != nil {
		t.Errorf("Couldn't unmarshal []byte to RawMessage: %s", err)
	}
	if _, valid2 := body2["id"]; !valid2 {
		t.Errorf("ID wasn't added back: %s", err)
	}

	{{$varNameSingular}}unmarshal := &{{$tableNameSingular}}{}
	if err = json.Unmarshal(marshaled2, {{$varNameSingular}}unmarshal); err != nil {
		t.Errorf("Couldn't unmarshal: %s", err)
	}

	if !reflect.DeepEqual({{$varNameSingular}}, {{$varNameSingular}}unmarshal) {
		t.Errorf("Unmarshaled object is not equal: %v != %v", {{$varNameSingular}}, {{$varNameSingular}}unmarshal)
	}
}
