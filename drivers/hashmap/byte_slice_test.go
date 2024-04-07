package hashmap

import (
	"testing"
)

func TestByteSliceMarshalYAML(t *testing.T) {
	data := map[string]ByteSlice{
		"key": []byte("value"),
	}

	dataBytes, err := yaml.Marshal(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "key: value\n"
	if string(dataBytes) != expected {
		t.Errorf("expected %q but got: %v", expected, string(dataBytes))
	}
}

func TestByteSliceUnmarshalYAML(t *testing.T) {
	var data map[string]ByteSlice
	err := yaml.Unmarshal([]byte(`key: value`), &data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(data["key"]) != "value" {
		t.Errorf("unexpected value: %v", string(data["key"]))
	}
}

func TestByteSliceUnmarshalYAMLError(t *testing.T) {
	var data map[string]ByteSlice
	err := yaml.Unmarshal([]byte(`key: 1`), &data)
	if err == nil || err.Error() != "expected string, but got !!int" {
		t.Errorf("error did not match: %v", err)
	}
}
