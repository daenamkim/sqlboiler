{{- $tableNameSingular := .Table.Name | singular | titleCase -}}
{{- $tableNamePlural := .Table.Name | plural | titleCase -}}
{{- $varNamePlural := .Table.Name | plural | camelCase -}}
{{- $varNameSingular := .Table.Name | singular | camelCase -}}


func test{{$tableNamePlural}}Marshal(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	random{{$varNameSingular}} := &{{$tableNameSingular}}{}
	if err = randomize.Struct(seed, random{{$varNameSingular}}, {{$varNameSingular}}DBTypes, true, {{$varNameSingular}}ColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize {{$tableNameSingular}} struct: %s", err)
	}

	/////////////////////// Test Models

	// Ignore default ID exclusion, testing MarshalJSONStruct rather than json.Marshal
	marshalFilter{{$varNameSingular}} := []byte{}
	if marshalFilter{{$varNameSingular}}, err = marshal.MarshalJSONStruct(random{{$varNameSingular}}, nil); err != nil {
		t.Errorf("Unable to marshal struct: %s", err)
	}

	// Unmarshal with ID
	unmarshalFilter{{$varNameSingular}} := &{{$tableNameSingular}}{}
	err = json.Unmarshal(marshalFilter{{$varNameSingular}}, &unmarshalFilter{{$varNameSingular}})
	if err != nil {
		t.Errorf("Couldn't unmarshal struct: %s", err)
	}

	// Check if Unmarshaled struct is equal to original
	if !reflect.DeepEqual(unmarshalFilter{{$varNameSingular}}, random{{$varNameSingular}}){
		t.Errorf("Unmarshaled struct is not equal to original: %s", err)
	}

	/////////////////////// Test Embedded Structs 

	embedded{{$varNameSingular}} := &embedded{{$tableNameSingular}}{
		{{$tableNameSingular}}: unmarshalFilter{{$varNameSingular}},
	}

	// Marshal embedded struct
	marshalEmbedded{{$varNameSingular}} := []byte{}
	marshalEmbedded{{$varNameSingular}}, err = json.Marshal(embedded{{$varNameSingular}})

	// Unmarshal embedded struct
	unmarshalEmbedded{{$varNameSingular}} := &embedded{{$tableNameSingular}}{}
	json.Unmarshal(marshalEmbedded{{$varNameSingular}}, unmarshalEmbedded{{$varNameSingular}})

	if !reflect.DeepEqual(embedded{{$varNameSingular}} , unmarshalEmbedded{{$varNameSingular}}) {
		t.Errorf("Unmarshaled embedded object is not equal: %+v != %+v", embedded{{$varNameSingular}}.{{$tableNameSingular}}, unmarshalEmbedded{{$varNameSingular}}.{{$tableNameSingular}})
	}

}

///////////////////// Test embedding models

type embedded{{$tableNameSingular}} struct{
	*{{$tableNameSingular}}
}

func (o *embedded{{$tableNameSingular}}) MarshalJSON() ([]byte, error) {
	exclude := map[string]bool{}
	return marshal.MarshalJSONStruct(o, exclude)
}

func (o *embedded{{$tableNameSingular}}) UnmarshalJSON(data []byte) error {
	specialNames := map[string]string{}
	return marshal.UnmarshalWrapper(o, data, specialNames)
}
