// Package mock is a Hord database driver used to assist in testing. This package satisfies the Hord database interface
// and can help applications using a Hord managed database.
//
// In using this package, users can create custom functions executed when calling the Database interface methods. Rather
// than writing a mock for the Hord interface, initialize this driver in place of a standard database driver.
//
// 	func TestMocking(t *testing.T) {
// 	        var db hord.Database
// 	        cfg := mock.Config{
// 	                // Create a fake GET function
// 	                GetFunc: func(key string) ([]byte, error) {
// 	                        if key == "works" {
// 	                                return []byte("Yes"), nil
// 	                        }
// 	                        return []byte{}, hord.ErrNil
// 	                },
// 	                // Create a fake SET function
// 	                SetFunc: func(key string, data []byte) error {
// 	                        if key == "works" {
// 	                                return nil
// 	                        }
// 	                        return fmt.Errorf("Error inserting data")
// 	                },
// 	        }
//
// 	        db, err := mock.Dial(cfg)
// 	        if err != nil {
// 	                t.Errorf("Unexpected error when creating Mock interface - %s", err)
// 	        }
//
// 	        t.Run("Validate Get", func(t *testing.T) {
// 	                _, err := db.Get("works")
// 	                if err != nil {
// 	                        t.Errorf("Get mocked function did not work as expected err returned - %s", err)
// 	                }
// 	        })
//
//	        t.Run("Validate Get Errors", func(t *testing.T) {
//	                _, err := db.Get("doesntwork")
//	                if err != hord.ErrNil {
//	                        t.Errorf("Get mocked function did not work as expected err returned - %s", err)
//	                }
//	        })
//
// 	        t.Run("Validate Set", func(t *testing.T) {
// 	                err := db.Set("works", []byte{})
// 	                if err != nil {
// 	                        t.Errorf("Set mocked function did not work as expected err returned - %s", err)
// 	                }
// 	        })
//
// 	        t.Run("Validate Set Errors", func(t *testing.T) {
// 	                err := db.Set("doesntwork", []byte{})
// 	                if err == nil {
// 	                        t.Errorf("Set mocked function did not work as expected err returned - %s", err)
// 	                }
// 	        })
// 	}
//
// This package, by default, offers a happy path for each mocked function. Custom functions only must be defined to alter
// the default behavior.
//
package mock

// Config is passed to Dial to configure this mock. By default, mocked functions will return with a happy path scenario.
// To override and customize the return use the appropriate functions defined within the Config struct.
type Config struct {
	// SetupFunc allows users to define a custom function executed in place of the default Database Setup method.
	SetupFunc func() error

	// HealthCheckFunc allows users to define a custom function executed in place of the default Database HealthCheck
	// method.
	HealthCheckFunc func() error

	// GetFunc allows users to define a custom function executed in place of the default Database Get method.
	GetFunc func(string) ([]byte, error)

	// SetFunc allows users to define a custom function executed in place of the default Database Set method.
	SetFunc func(string, []byte) error

	// DeleteFunc allows users to define a custom function executed in place of the default Database Delete method.
	DeleteFunc func(string) error

	// KeysFunc allows users to define a custom function executed in place of the default Database Keys method.
	KeysFunc func() ([]string, error)
}

// Database is an object returned by the Dial function. This struct satisfies the Hord Database interface and can
// provide mocked functionality for the Hord interface.
type Database struct {
	// setupFunc allows users to define a custom function executed in place of the default Database Setup method.
	setupFunc func() error

	// healthCheckFunc allows users to define a custom function executed in place of the default Database HealthCheck
	// method.
	healthCheckFunc func() error

	// getFunc allows users to define a custom function executed in place of the default Database Get method.
	getFunc func(string) ([]byte, error)

	// setFunc allows users to define a custom function executed in place of the default Database Set method.
	setFunc func(string, []byte) error

	// deleteFunc allows users to define a custom function executed in place of the default Database Delete method.
	deleteFunc func(string) error

	// keysFunc allows users to define a custom function executed in place of the default Database Keys method.
	keysFunc func() ([]string, error)
}

// Dial will mock connecting to a remote database. Users can use the returned Database object to fake interactions
// with a Hord Database.
//
//          var db hord.Database
//          cfg := mock.Config{
//                  // Create a fake GET function
//                  GetFunc: func(key string) ([]byte, error) {
//                          if key == "works" {
//                                  return []byte("Yes"), nil
//                          }
//                          return []byte{}, hord.ErrNil
//                  },
//                  // Create a fake SET function
//                  SetFunc: func(key string, data []byte) error {
//                          if key == "works" {
//                                  return nil
//                          }
//                          return fmt.Errorf("Error inserting data")
//                  },
//          }
//
//          db, err := mock.Dial(cfg)
//          if err != nil {
//                  t.Errorf("Unexpected error when creating Mock interface - %s", err)
//          }
//
func Dial(c Config) (*Database, error) {
	db := &Database{}
	db.setupFunc = c.SetupFunc
	db.healthCheckFunc = c.HealthCheckFunc
	db.getFunc = c.GetFunc
	db.setFunc = c.SetFunc
	db.deleteFunc = c.DeleteFunc
	db.keysFunc = c.KeysFunc
	return db, nil
}

// Setup provides a mocked function, which, when executed without any configuration, will return nil. If Users have
// defined a custom Setup function, Setup will run the custom function producing the results.
func (db Database) Setup() error {
	if db.setupFunc != nil {
		return db.setupFunc()
	}
	return nil
}

// HealthCheck provides a mocked function, which, when executed without any configuration, will return nil. If Users
// have defined a custom HealthCheck function, HealthCheck will run the custom function producing the results.
func (db Database) HealthCheck() error {
	if db.healthCheckFunc != nil {
		return db.healthCheckFunc()
	}
	return nil
}

// Get provides a mocked function, which will return an empty byte slice and no error when executed without any
// configuration. If Users have defined a custom Get function, Get will run the custom function producing the results.
func (db Database) Get(key string) ([]byte, error) {
	if db.getFunc != nil {
		return db.getFunc(key)
	}
	return []byte{}, nil
}

// Set provides a mocked function, which will return no error when executed without any configuration. If
// Users have defined a custom Set function, Set will run the custom function producing the results.
func (db Database) Set(key string, data []byte) error {
	if db.setFunc != nil {
		return db.setFunc(key, data)
	}
	return nil
}

// Delete provides a mocked function, which will return no error when executed without any configuration.
// If Users have defined a custom Delete function, Delete will run the custom function producing the results.
func (db Database) Delete(key string) error {
	if db.deleteFunc != nil {
		return db.deleteFunc(key)
	}
	return nil
}

// Keys provides a mocked function, which will return an empty string slice with no error when executed
// without any configuration. If Users have defined a custom Keys function, Keys will run the custom
// function producing the results.
func (db Database) Keys() ([]string, error) {
	if db.keysFunc != nil {
		return db.keysFunc()
	}
	return []string{}, nil
}

// Close, when called, will return and not act. Use this function to mock a Close Database call.
func (db Database) Close() {}
