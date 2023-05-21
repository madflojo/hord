package nats

import (
	"testing"
)

type TestCase struct {
	passDial  bool
	passSetup bool
	cfg       Config
}

func TestNATSConnectivity(t *testing.T) {
	testCases := map[string]TestCase{
		"Happy Path Config": {
			passDial:  true,
			passSetup: true,
			cfg: Config{
				URL:    "nats",
				Bucket: "test",
			},
		},
		"No Bucket": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				URL: "nats",
			},
		},
		"Empty Config": {
			passDial:  false,
			passSetup: false,
			cfg:       Config{},
		},
		"No URL": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				Bucket: "test",
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
