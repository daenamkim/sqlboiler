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

	// Prep Filter
	includeFields := []string{"ID"}
	var mWithID []byte
	if mWithID, err = {{$varNameSingular}}.MarshalJSONFilter(includeFields, nil); err != nil {
		t.Errorf("Unable to marshal struct: %s", err)
	}

	fmt.Printf("{{$tableNameSingular}} marshaled 2 %+v\n", string(mWithID))

	// Step 1: Check if adding ID back works
	var rawWithID map[string]*json.RawMessage
	err = json.Unmarshal(mWithID, &rawWithID)
	if err != nil {
		t.Errorf("Couldn't unmarshal []byte to RawMessage: %s", err)
	}

	// Step 2: Prep Default json.Marshal
	var mDefault []byte
	if mDefault, err = json.Marshal({{$varNameSingular}}); err != nil {
		t.Errorf("Unable to marshal struct: %s", err)
	}
	// fmt.Printf("\n\n%+v\n", {{$varNameSingular}})
	fmt.Printf("{{$tableNameSingular}} marshaled orig %+v\n", string(mDefault))
	var rawDefault map[string]*json.RawMessage
	err = json.Unmarshal(mDefault, &rawDefault)
	if err != nil {
		t.Errorf("Couldn't unmarshal []byte to RawMessage: %s", err)
	}

	// If ID was ignored in step 1, and then not filtered by step 2 
	if _, valid := rawWithID["id"]; valid {
		if _, valid = rawDefault["id"]; valid {
			t.Errorf("Default ID wasn't filtered: %s", err)
		}
	}

	{{$varNameSingular}}unmarshal := &{{$tableNameSingular}}{}
	if err = json.Unmarshal(mWithID, {{$varNameSingular}}unmarshal); err != nil {
		t.Errorf("Couldn't unmarshal: %s", err)
	}

	fmt.Printf("{{$tableNameSingular}} unmarshaled unfiltered %+v\n", {{$varNameSingular}}unmarshal)

	// Check to see if ID added back and original struct are equivalent
	if !reflect.DeepEqual({{$varNameSingular}}, {{$varNameSingular}}unmarshal) {
		t.Errorf("Unmarshaled object is not equal: %v != %v", {{$varNameSingular}}, {{$varNameSingular}}unmarshal)
	}



	/////////////////////// 

	eWithID := &embedded{{$tableNameSingular}}{
		{{$tableNameSingular}}: {{$varNameSingular}}unmarshal,
	}

	var emWithID []byte
	var emWithID2 []byte
	emWithID, err = json.Marshal(eWithID)
	emWithID2, err = marshal.MarshalJSONWrapper(eWithID, nil, nil)

	fmt.Printf("{{$tableNameSingular}} emWithID %+v\n", eWithID)
	fmt.Printf("{{$tableNameSingular}} emWithID json %+v\n", string(emWithID))
	fmt.Printf("{{$tableNameSingular}} emWithID2 json %+v\n", string(emWithID2))

	umWithID := &embedded{{$tableNameSingular}}{}
	json.Unmarshal(emWithID, umWithID)

	fmt.Printf("{{$tableNameSingular}} umWithID %+v\n", umWithID.{{$tableNameSingular}})

	if !reflect.DeepEqual(umWithID, eWithID) {
		t.Errorf("Unmarshaled embedded object is not equal: %+v != %+v", umWithID.{{$tableNameSingular}}, eWithID.{{$tableNameSingular}})
	}

}

// Test embedding models ///////////////////
type embedded{{$tableNameSingular}} struct{
	*{{$tableNameSingular}}
}

func (o *embedded{{$tableNameSingular}}) MarshalJSON() ([]byte, error) {
	includeFields := []string{}
	excludeFields := []string{}
	return marshal.MarshalJSONWrapper(o, includeFields, excludeFields)
}

func (o *embedded{{$tableNameSingular}}) UnmarshalJSON(data []byte) error {
	specialNames := map[string]string{}
	return marshal.UnmarshalWrapper(o, data, specialNames)
}
