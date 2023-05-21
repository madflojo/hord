package nats

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
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
		"No JetStream Config": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				URL:    "nojsnats",
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
		"Invalid Bucket Name": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				URL:    "nats",
				Bucket: "invalid bucket 123",
			},
		},
		"Invalid URL": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				URL:    "",
				Bucket: "test",
			},
		},
		"TLS with no Cert": {
			passDial:  false,
			passSetup: false,
			cfg: Config{
				URL:           "tls://nats",
				Bucket:        "test",
				SkipTLSVerify: true,
				TLSConfig:     &tls.Config{},
				Options: nats.Options{
					AllowReconnect: true,
					MaxReconnect:   10,
					ReconnectWait:  5 * time.Second,
					Timeout:        1 * time.Second,
				},
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
