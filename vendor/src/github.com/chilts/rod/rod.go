package rod

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/boltdb/bolt"
)

var (
	// ErrLocationMustHaveAtLeastOneBucket is returned if any location given hasn't got anything in it, ie. it is
	// empty.
	ErrLocationMustHaveAtLeastOneBucket = errors.New("location must specify at least one bucket")

	// ErrInvalidLocationBucket is returned if any location is blank, ie. you specified something like "user..field".
	ErrInvalidLocationBucket = errors.New("invalid location bucket")

	// ErrKeyNotProvided is returned if key was not specified, ie. it is empty.
	ErrKeyNotProvided = errors.New("key must be specified")
)

// Del will find your bucket location and delete the key specified. It doesn't matter what is in the key's value, since
// it is ignored during this operation. If the bucket doesn't exist, no error is returned since technically you've
// already got what you asked for. Similarly, if the key doesn't exist, no error is returned for the same reason.
func Del(tx *bolt.Tx, location, key string) error {
	if location == "" {
		return ErrLocationMustHaveAtLeastOneBucket
	}
	if key == "" {
		return ErrKeyNotProvided
	}

	b, err := GetBucket(tx, location)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	// now delete the key
	return b.Delete([]byte(key))
}

// Put will find your bucket location and put your value into the key specified. The location is specified as a
// hierarchy of bucket names such as "users", "users.chilts", or "users.chilts.posts" and will be split on the period
// for each bucket name.
//
// At every bucket specified in the location, CreateBucketIfNotExists() is called to make sure it exists. If any of these
// fail, the error is returned.
//
// Once the final bucket is found, the value is put using the key.
//
// Examples:
//
//    rod.Put(tx, "social", "twitter-123456", []byte("chilts"))
//    rod.Put(tx, "users.chilts", "email", []byte("andychilton@gmail.com"))
//    rod.Put(tx, "users.chilts.posts", "hello-world", []byte("Hello, World!"))
//
// The location must have at least one bucket ("" is not allowed), and the key must also be a non-empty string. The
// transaction must be a writeable one otherwise an error is returned.
func Put(tx *bolt.Tx, location, key string, value []byte) error {
	if location == "" {
		return ErrLocationMustHaveAtLeastOneBucket
	}
	if key == "" {
		return ErrKeyNotProvided
	}

	// split the 'bucket' on '.'
	buckets := strings.Split(location, ".")
	if buckets[0] == "" {
		return ErrInvalidLocationBucket
	}

	// get the first bucket
	b, errCreateTopLevel := tx.CreateBucketIfNotExists([]byte(buckets[0]))
	if errCreateTopLevel != nil {
		return errCreateTopLevel
	}

	// now, only loop through if we have more than 2
	if len(buckets) > 1 {
		for _, name := range buckets[1:] {
			if name == "" {
				return ErrInvalidLocationBucket
			}
			var err error
			b, err = b.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}
	}

	return b.Put([]byte(key), value)
}

// PutString converts the string to []byte and calls Put. Everything that applies there applies here too.
func PutString(tx *bolt.Tx, location, key, value string) error {
	return Put(tx, location, key, []byte(value))
}

// PutJson calls json.Marshal() to serialise the value into []byte and calls rod.Put with the result.
func PutJson(tx *bolt.Tx, location, key string, v interface{}) error {
	// now put this value in this key
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return Put(tx, location, key, value)
}

// Get will fetch the raw bytes from the BoltDB. If any bucket doesn't exist it will return nil. If the key doesn't
// exist it will also return nil.
//
// Error returned from this function are:
// * ErrLocationMustHaveAtLeastOneBucket if no location was specified
// * ErrKeyNotProvided if no key was specified
func Get(tx *bolt.Tx, location, key string) ([]byte, error) {
	b, err := GetBucket(tx, location)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}

	if key == "" {
		return nil, ErrKeyNotProvided
	}

	// get this key
	return b.Get([]byte(key)), nil
}

// GetString calls Get and converts the []byte to a string before returning it to you. Everything that applies there
// applies here too.
func GetString(tx *bolt.Tx, location, key string) (string, error) {
	raw, err := Get(tx, location, key)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// GetJson calls Get and then json.Unmarshal() with the result to deserialise the value into interface{}. If any bucket
// doesn't exist we just return nil with nothing placed into v. The same if the key doesn't exist.
func GetJson(tx *bolt.Tx, location, key string, v interface{}) error {
	// get this key
	raw, err := Get(tx, location, key)
	if err != nil {
		return err
	}
	if raw == nil {
		// no key exists
		return nil
	}

	// decode to the v interface{}
	return json.Unmarshal(raw, &v)
}

// GetBucket returns this nested bucket from the store. If any bucket along the way does not exist, then no bucket is
// returned (nil) but not error is returned either.
func GetBucket(tx *bolt.Tx, location string) (*bolt.Bucket, error) {
	if location == "" {
		return nil, ErrLocationMustHaveAtLeastOneBucket
	}

	// split the 'bucket' on '.'
	buckets := strings.Split(location, ".")
	if buckets[0] == "" {
		return nil, ErrInvalidLocationBucket
	}

	// get the first bucket
	b := tx.Bucket([]byte(buckets[0]))
	if b == nil {
		return nil, nil
	}

	// loop through if we have more than 2
	if len(buckets) > 1 {
		for _, name := range buckets[1:] {
			if name == "" {
				return nil, ErrInvalidLocationBucket
			}
			b = b.Bucket([]byte(name))
			if b == nil {
				return nil, nil
			}
		}
	}

	return b, nil
}

// SelAll will give you everything inside the bucket specified by location. The newItem function you pass in will be
// called for every key in the bucket and should just return an empty instance of your type. Append will also be called
// for every item once unmarshalling has taken place.
//
//   animals := make([]*Animal, 0)
//   err := SelAll(tx, "animal", func() interface{} {
//       return Animal{}
//   }, func(v interface{}) {
//       a, _ := v.(Animal)
//       animals = append(animals, &a)
//   })
//
// It's a bit of boilerplate but you could just pass in a newItem function declared earlier in the program. This API
// is subject to change since it could probably be improved upon.
func SelAll(tx *bolt.Tx, location string, newItem func() interface{}, append func(interface{})) error {
	b, err := GetBucket(tx, location)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	// use a cursor to iterate through this bucket
	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		// get a new thing
		item := newItem()
		err := json.Unmarshal(v, &item)
		if err != nil {
			return err
		}

		// now call the append function
		append(item)
	}

	return nil
}
