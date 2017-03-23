// Package rod is a simple way to put/get values to/from a BoltDB (https://github.com/boltdb/bolt) store. It can deal
// with deep-bucket hierarchies easily and is therefore a lightning-rod straight to the value you want. Hence the name.
//
// Whilst this package won't solve all of your problems or use-cases, it does make a few things simple and is used
// successfully https://publish.li/ and https://weekproject.com/ and various other applications.
//
// Whilst rod just uses strings for it's location and key values (unlike Bolt using []byte), rod helps with the
// following value types:
//
// • []byte
//
// • string
//
// • JSON
//
// Again, everything is a convenience and you should be aware of any overhead rod introduces. However, since rod is
// designed to be minimal we try not to add much overhead at all (in terms of both code size and run-time overhead).
//
// (Ends)
package rod
