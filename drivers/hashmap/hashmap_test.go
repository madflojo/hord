package hashmap

import (
	"encoding/json"
	"os"
	"testing"
)

var fileTypeCases = []struct {
	extension string
	unmarshal func([]byte, interface{}) error
	marshal   func(interface{}) ([]byte, error)
}{
	{
		"yaml",
		yaml.Unmarshal,
		yaml.Marshal,
	},
	{
		"json",
		json.Unmarshal,
		json.Marshal,
	},
}

func TestSetupCreatesFileIfNotExist(t *testing.T) {
	tests := []struct {
		ext string
	}{
		{"yaml"},
		{"yml"},
		{"json"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			filename := "testdata/empty_file." + tt.ext
			defer os.RemoveAll(filename)

			db, err := Dial(Config{Filename: filename})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Setup()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// make sure file exists
			_, err = os.Stat(filename)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDial(t *testing.T) {
	t.Run("ErrorInvalidFilename", func(t *testing.T) {
		_, err := Dial(Config{Filename: "testdata/empty_file.txt"})
		if err == nil || err.Error() != "filename must have yaml, yml, or json extension" {
			t.Errorf("error did not match: %v", err)
		}
	})
}

func TestSaveToLocalFileAfterSet(t *testing.T) {
	for _, tt := range fileTypeCases {
		t.Run(tt.extension, func(t *testing.T) {
			filename := "testdata/save_test." + tt.extension
			defer os.RemoveAll(filename)

			db, err := Dial(Config{Filename: filename})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Setup()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Set("key", []byte("value"))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			db.Close()

			data, err := readFile(filename, tt.unmarshal)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			result := string(data["key"])
			if result != "value" {
				t.Errorf("unexpected value: %v", result)
			}
		})
	}
}

func TestLoadingFromExistingFile(t *testing.T) {
	for _, tt := range fileTypeCases {
		t.Run("ValidFileContents_"+tt.extension, func(t *testing.T) {
			filename := "testdata/load_data_test." + tt.extension

			db, err := Dial(Config{Filename: filename})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Setup()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			value, err := db.Get("key")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if string(value) != "value" {
				t.Errorf("unexpected value: %s", string(value))
			}
		})

		t.Run("InvalidFileContents_"+tt.extension, func(t *testing.T) {
			filename := "testdata/load_data_invalid_test." + tt.extension

			db, err := Dial(Config{Filename: filename})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Setup()
			if err == nil {
				t.Error("expected error but was nil")
			}

			// different error expectations depending on format
			var expectedErr string
			switch tt.extension {
			case "yaml":
				expectedErr = "unable to unmarshal data from file: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `this is...` into map[string]hashmap.ByteSlice"
			case "json":
				expectedErr = "unable to unmarshal data from file: invalid character 'h' in literal true (expecting 'r')"
			}

			if err.Error() != expectedErr {
				t.Errorf("error did not match: %v", err)
			}
		})
	}
}

func TestComplexObjectSaveToFile(t *testing.T) {
	data := map[string]interface{}{
		"string":        "string",
		"integer":       1,
		"float":         1.5,
		"string_array":  []string{"a", "b", "c"},
		"integer_array": []int{1, 2, 3},
		"object": map[string]interface{}{
			"key": "value",
		},
	}

	for _, tt := range fileTypeCases {
		t.Run(tt.extension, func(t *testing.T) {
			filename := "testdata/complex_data_test." + tt.extension
			defer os.RemoveAll(filename)

			db, err := Dial(Config{Filename: filename})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Setup()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Save objects to DB as JSON and YAML
			dataJSON, err := json.Marshal(data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Set("data_json", dataJSON)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			dataYAML, err := yaml.Marshal(data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			err = db.Set("data_yaml", dataYAML)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			db.Close()

			t.Run("CompareSavedFilesToExpectations", func(t *testing.T) {
				data, err := readFile(filename, tt.unmarshal)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				expectedData, err := readFile("testdata/complex_data_test_expected."+tt.extension, tt.unmarshal)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if string(data["data"]) != string(expectedData["data"]) {
					t.Errorf("data did not match: %v", string(data["data"]))
				}
			})

			t.Run("RecreateDBAndGetData", func(t *testing.T) {
				db, err := Dial(Config{Filename: filename})
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				err = db.Setup()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				newDataJSON, err := db.Get("data_json")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(newDataJSON) != string(dataJSON) {
					t.Errorf("data did not match after reading from file: %s", string(newDataJSON))
				}

				newDataYAML, err := db.Get("data_yaml")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(newDataYAML) != string(dataYAML) {
					t.Errorf("data did not match after reading from file: %s", string(newDataYAML))
				}
			})
		})
	}
}

func readFile(filename string, unmarshal func([]byte, interface{}) error) (map[string]ByteSlice, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var parsedData map[string]ByteSlice
	err = unmarshal(data, &parsedData)
	if err != nil {
		return nil, err
	}

	return parsedData, nil
}
