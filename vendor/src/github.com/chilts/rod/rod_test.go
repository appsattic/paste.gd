package rod

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
)

type Animal struct {
	Type string
	Name string
}

type User struct {
	Username string
	Logins   int
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestAll(t *testing.T) {
	dir, err := ioutil.TempDir("", "rod-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "rod.db")
	defer os.Remove(filename)

	// Open the database.
	db, err := bolt.Open(filename, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	t.Run("Simple Put and Get", func(t *testing.T) {
		msg := "Hello, World!"

		err := db.Update(func(tx *bolt.Tx) error {
			// put this message
			err := Put(tx, "message", "hello-world", []byte(msg))
			check(err)

			// get it back
			storedMsg, err := Get(tx, "message", "hello-world")
			check(err)
			if string(storedMsg) != msg {
				log.Fatalf("Received msg '%s' is not the same as the original '%s'", string(storedMsg), msg)
			}

			return nil
		})

		check(err)
	})

	t.Run("Simple PutString and GetString", func(t *testing.T) {
		msg := "Hello, World!"

		err := db.Update(func(tx *bolt.Tx) error {
			// put this message
			err := PutString(tx, "strings", "hello-world", msg)
			check(err)

			// get it back
			storedMsg, err := GetString(tx, "strings", "hello-world")
			check(err)
			if storedMsg != msg {
				log.Fatalf("Received msg '%s' is not the same as the original '%s'", storedMsg, msg)
			}

			return nil
		})

		check(err)
	})

	t.Run("Simple PutJson and GetJson", func(t *testing.T) {
		user := User{"chilts", 1}

		err := db.Update(func(tx *bolt.Tx) error {
			// put this message
			errPutJson := PutJson(tx, "user", "chilts", user)
			check(errPutJson)

			// get it back
			storedUser := User{}
			errGetJson := GetJson(tx, "user", "chilts", &storedUser)
			check(errGetJson)
			if storedUser.Username != user.Username {
				log.Fatalf("Received msg '%s' is not the same as the original '%s'", storedUser.Username, user.Username)
			}
			if storedUser.Logins != user.Logins {
				log.Fatalf("Received logins '%d' is not the same as the original '%d'", storedUser.Logins, user.Logins)
			}

			return nil
		})

		check(err)
	})

	t.Run("SelAll", func(t *testing.T) {
		// Start a read-write transaction.
		if err := db.Update(func(tx *bolt.Tx) error {
			dog := Animal{"dog", "rover"}
			cat := Animal{"cat", "willow"}
			horse := Animal{"horse", "ed"}

			_ = PutJson(tx, "animal", "dog", &dog)
			_ = PutJson(tx, "animal", "cat", &cat)
			_ = PutJson(tx, "animal", "horse", &horse)

			animals := make([]*Animal, 0)
			err := SelAll(tx, "animal", func() interface{} {
				return Animal{}
			}, func(v interface{}) {
				a, notOk := v.(Animal)
				if notOk {
					t.Fatal("Thing returned from SelAll is not an Animal")
				}
				animals = append(animals, &a)
			})

			return err
		}); err != nil {
			log.Fatal(err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		location := "delete"
		key := "key"
		val := "val"

		err := db.Update(func(tx *bolt.Tx) error {
			// put a new value and make sure it works
			errPut := PutString(tx, location, key, val)
			check(errPut)

			// now delete it again
			errDel1 := Del(tx, location, key)
			check(errDel1)

			// delete it again (since it now doesn't exist)
			errDel2 := Del(tx, location, key)
			check(errDel2)

			// delete a key from a location that doesn't exist
			errDel3 := Del(tx, "doesnt-exist", key)
			check(errDel3)

			return nil
		})

		check(err)
	})
}
