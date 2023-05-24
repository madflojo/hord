package hashmap

import (
	"encoding/json"
	"os"
	"testing"
)

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
	t.Run("SuccessfullySaveJSON", func(t *testing.T) {
		filename := "testdata/save_test.json"
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

		data, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		var parsedData map[string]ByteSlice
		err = json.Unmarshal(data, &parsedData)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		result := string(parsedData["key"])
		if result != "value" {
			t.Errorf("unexpected value: %v", result)
		}
	})
}

func TestLoadingFromExistingFile(t *testing.T) {
	tests := []struct {
		extension string
	}{
		{"json"},
		{"yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
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
	}
}
