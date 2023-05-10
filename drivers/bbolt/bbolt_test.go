package bbolt

import (
	"os"
	"strconv"
	"testing"
	"time"
)

type TestCase struct {
	passDial  bool
	passSetup bool
	cfg       Config
}

func TmpFn() string {
	// Snagged from ioutil.TempFile
	r := uint32(time.Now().UnixNano() + int64(os.Getpid()))
	r = r*1664525 + 1013904223
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func TestBBoltConnectivity(t *testing.T) {
	// Create Directory for Test Execution
	tmpDir := "/tmp/" + TmpFn()
	err := os.Mkdir(tmpDir, 0750)
	if err != nil {
		t.Fatalf("Unable to create test directory - %s", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := map[string]TestCase{
		"Happy Path Config": {
			passDial:  true,
			passSetup: true,
			cfg: Config{
				Bucketname:  "test",
				Filename:    tmpDir + "/" + TmpFn() + "happy",
				Timeout:     time.Duration(15 * time.Second),
				Permissions: 0600,
			},
		},
		"Default Permissions": {
			passDial:  true,
			passSetup: true,
			cfg: Config{
				Bucketname: "test",
				Filename:   tmpDir + "/" + TmpFn() + "perms",
				Timeout:    time.Duration(15 * time.Second),
			},
		},
		"Default Timeout": {
			passDial:  true,
			passSetup: true,
			cfg: Config{
				Bucketname:  "test",
				Filename:    tmpDir + "/" + TmpFn() + "timeout",
				Permissions: 0600,
			},
		},
		"No Bucket Name": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				Filename:    tmpDir + "/" + TmpFn() + "nobucket",
				Timeout:     time.Duration(15 * time.Second),
				Permissions: 0600,
			},
		},
		"No Filename": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				Bucketname:  "test",
				Timeout:     time.Duration(15 * time.Second),
				Permissions: 0600,
			},
		},
		"Non-existent Path": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				Bucketname:  "test",
				Filename:    "/doesnotexist/nope",
				Timeout:     time.Duration(15 * time.Second),
				Permissions: 0600,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := Dial(tc.cfg)
			if err != nil {
				if tc.passDial {
					t.Fatalf("unexpected failure while Dialing database - %s", err)
				}
			}
			if err == nil && !tc.passDial {
				t.Fatalf("unexpected success while Dialing database")
			}

			err = db.Setup()
			if err != nil {
				if tc.passSetup {
					t.Fatalf("unexpected failure while Setting up database - %s", err)
				}
			}
			if err == nil && !tc.passSetup {
				t.Fatalf("unexpected success while Setting up database")
			}
		})
	}
}
