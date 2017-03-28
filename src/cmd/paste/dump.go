package main

// From : https://gist.github.com/chilts/687ec1e8c5337213a7e1a5de2d3584ae

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

// Call it with something like:
//
//     go dumpEvery(db, time.Duration(15)*time.Minute, "/var/lib/project/dump")
//
// for every 15 mins.
//
// $ ls -l
// total 20
// -rwx------ 1 chilts chilts 1075 Mar 29 09:59 20170329-095936.db.gz
// -rwx------ 1 chilts chilts 1075 Mar 29 09:59 20170329-095946.db.gz
// -rwx------ 1 chilts chilts 1075 Mar 29 09:59 20170329-095956.db.gz
// -rwx------ 1 chilts chilts 1222 Mar 29 10:00 20170329-100006.db.gz
// -rwx------ 1 chilts chilts 1222 Mar 29 10:00 20170329-100016.db.gz
func dumpEvery(db *bolt.DB, d time.Duration, dir string) {
	ticker := time.NewTicker(d)

	for {
		select {
		case <-ticker.C:
			// do stuff
			log.Println("Dumping the DB now")
			dump(db, dir)
		}
	}
}

func dump(db *bolt.DB, dir string) error {
	filename := path.Join(dir, time.Now().Format("20060102-150405")+".db.gz")
	fmt.Printf("filename=%s\n", filename)

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	// gzip the output
	zw := gzip.NewWriter(f)
	defer zw.Close()

	err = db.View(func(tx *bolt.Tx) error {
		n, err := tx.WriteTo(zw)
		log.Printf("DB Dump written %d bytes\n", n)
		return err
	})
	return err
}
