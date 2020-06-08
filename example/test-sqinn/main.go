/*
A benchmark for sqinn-go.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func testFunctions(sqinnPath, dbFile string, nusers int) {
	funcname := "testFunctions"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nusers=%d", sqinnPath, dbFile, nusers)
	assert := func(c bool) {
		if !c {
			panic("assertion failed")
		}
	}
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	// make sure db does not exist
	os.Remove(dbFile)
	// launch sqinn
	sq, err := sqinn.New(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	err = sq.Open(dbFile)
	check(err)
	t1 := time.Now()
	// prepare schema
	sql := "CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)"
	err = sq.Prepare(sql)
	check(err)
	_, err = sq.Step()
	check(err)
	err = sq.Finalize()
	check(err)
	// insert users
	sql = "BEGIN TRANSACTION"
	err = sq.Prepare(sql)
	check(err)
	_, err = sq.Step()
	check(err)
	err = sq.Finalize()
	check(err)
	sql = "INSERT INTO users (id, name, age, rating) VALUES (?,?,?,?)"
	err = sq.Prepare(sql)
	check(err)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User_%d", id)
		age := 33 + i
		rating := 0.13 * float64(i+1)
		check(sq.Bind(1, id))
		check(sq.Bind(2, name))
		check(sq.Bind(3, age))
		check(sq.Bind(4, rating))
		_, err = sq.Step()
		check(err)
		check(sq.Reset())
		ch, err := sq.Changes()
		check(err)
		assert(ch == 1)
	}
	err = sq.Finalize()
	check(err)
	sql = "COMMIT"
	err = sq.Prepare(sql)
	check(err)
	_, err = sq.Step()
	check(err)
	err = sq.Finalize()
	check(err)
	t2 := time.Now()
	// query users
	sql = "SELECT id, name, age, rating FROM users ORDER BY id"
	err = sq.Prepare(sql)
	check(err)
	var more bool
	more, err = sq.Step()
	check(err)
	var nrows int
	for more {
		nrows++
		idValue, err := sq.Column(0, sqinn.ValInt)
		check(err)
		nameValue, err := sq.Column(1, sqinn.ValText)
		check(err)
		ageValue, err := sq.Column(2, sqinn.ValInt)
		check(err)
		ratingValue, err := sq.Column(3, sqinn.ValDouble)
		check(err)
		_, _, _, _ = idValue, nameValue, ageValue, ratingValue
		// log.Printf("%d | %s | %d | %g", idValue.Value, nameValue.Value, ageValue.Value, ratingValue.Value)
		more, err = sq.Step()
		check(err)
	}
	log.Printf("fetched %d rows", nrows)
	err = sq.Finalize()
	check(err)
	t3 := time.Now()
	// close db
	err = sq.Close()
	check(err)
	// terminate sqinn
	err = sq.Terminate()
	check(err)
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testUsers(sqinnPath, dbFile string, nusers int, bindRating bool) {
	funcname := "testUsers"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nusers=%d, bindRating=%t", sqinnPath, dbFile, nusers, bindRating)
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	// make sure db doesn't exist
	os.Remove(dbFile)
	// launch sqinn
	sq, err := sqinn.New(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	err = sq.Open(dbFile)
	check(err)
	// prepare schema
	_, err = sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)")
	// insert users
	t1 := time.Now()
	_, err = sq.Exec("BEGIN TRANSACTION", 1, 0, nil)
	values := make([]interface{}, 0, nusers*4)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User_%d", id)
		age := 33 + i
		rating := 0.13 * float64(i+1)
		if bindRating {
			values = append(values, id, name, age, rating)
		} else {
			values = append(values, id, name, age, nil)
		}
	}
	_, err = sq.Exec("INSERT INTO users (id, name, age, rating) VALUES (?,?,?,?)", nusers, 4, values)
	_, err = sq.ExecOne("COMMIT")
	t2 := time.Now()
	// query users
	colTypes := []byte{sqinn.ValInt, sqinn.ValText, sqinn.ValInt, sqinn.ValDouble}
	rows, err := sq.Query("SELECT id, name, age, rating FROM users ORDER BY id", nil, colTypes)
	check(err)
	log.Printf("fetched %d rows", len(rows))
	// for _, row := range rows {
	// 	log.Printf("%d | %s | %d | %g",
	// 		row.Values[0].IntValue.Value,
	// 		row.Values[1].StringValue.Value,
	// 		row.Values[2].IntValue.Value,
	// 		row.Values[3].DoubleValue.Value,
	// 	)
	// }
	t3 := time.Now()
	// close db
	err = sq.Close()
	check(err)
	// terminate sqinn
	err = sq.Terminate()
	check(err)
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testComplex(sqinnPath, dbFile string, nprofiles, nusers, nlocations int) {
	funcname := "testComplex"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nprofiles, nusers, nlocations = %d, %d, %d", sqinnPath, dbFile, nprofiles, nusers, nlocations)
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	sq, err := sqinn.New(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// make sure db doesn't exist
	os.Remove(dbFile)
	// open db
	check(sq.Open(dbFile))
	_, err = sq.ExecOne("PRAGMA foreign_keys=1")
	check(err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS locations")
	check(err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS users")
	check(err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS profiles")
	check(err)
	_, err = sq.ExecOne("CREATE TABLE profiles (id VARCHAR PRIMARY KEY NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL)")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_profiles_name ON profiles(name);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_profiles_active ON profiles(active);")
	check(err)
	_, err = sq.ExecOne("CREATE TABLE users (id VARCHAR PRIMARY KEY NOT NULL, profileId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (profileId) REFERENCES profiles(id))")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_users_profileId ON users(profileId);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_users_name ON users(name);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_users_active ON users(active);")
	check(err)
	_, err = sq.ExecOne("CREATE TABLE locations (id VARCHAR PRIMARY KEY NOT NULL, userId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (userId) REFERENCES users(id))")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_locations_userId ON locations(userId);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_locations_name ON locations(name);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_locations_active ON locations(active);")
	check(err)
	// insert
	t1 := time.Now()
	_, err = sq.ExecOne("BEGIN TRANSACTION")
	check(err)
	values := make([]interface{}, 0, nprofiles*3)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("ProfileGo %d", p)
		active := p % 2
		values = append(values, profileID, name, active)
	}
	_, err = sq.Exec("INSERT INTO profiles (id,name,active) VALUES(?,?,?)", nprofiles, 3, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	_, err = sq.ExecOne("BEGIN TRANSACTION")
	check(err)
	values = make([]interface{}, 0, nprofiles*nusers*4)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		for u := 0; u < nusers; u++ {
			userID := fmt.Sprintf("user_%d_%d", p, u)
			name := fmt.Sprintf("User %d %d", p, u)
			active := u % 2
			values = append(values, userID, profileID, name, active)
		}
	}
	_, err = sq.Exec("INSERT INTO users (id,profileId,name,active) VALUES(?,?,?,?)", nprofiles*nusers, 4, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	_, err = sq.ExecOne("BEGIN TRANSACTION")
	check(err)
	values = make([]interface{}, 0, nprofiles*nusers*nlocations*4)
	for p := 0; p < nprofiles; p++ {
		for u := 0; u < nusers; u++ {
			userID := fmt.Sprintf("user_%d_%d", p, u)
			for l := 0; l < nlocations; l++ {
				locationID := fmt.Sprintf("location_%d_%d_%d", p, u, l)
				name := fmt.Sprintf("Location %d %d %d", p, u, l)
				active := l % 2
				values = append(values, locationID, userID, name, active)
			}
		}
	}
	_, err = sq.Exec("INSERT INTO locations (id,userId,name,active) VALUES(?,?,?,?)", nprofiles*nusers*nlocations, 4, values)
	check(err)
	_, err = sq.Exec("COMMIT", 1, 0, nil)
	check(err)
	t2 := time.Now()
	// query
	sql := "SELECT locations.id, locations.userId, locations.name, locations.active, users.id, users.profileId, users.name, users.active, profiles.id, profiles.name, profiles.active " +
		"FROM locations " +
		"LEFT JOIN users ON users.id = locations.userId " +
		"LEFT JOIN profiles ON profiles.id = users.profileId " +
		"WHERE locations.active = ? OR locations.active = ? " +
		"ORDER BY locations.name, locations.id, users.name, users.id, profiles.name, profiles.id"
	rows, err := sq.Query(sql, []interface{}{0, 1}, []byte{sqinn.ValText, sqinn.ValText, sqinn.ValText, sqinn.ValInt, sqinn.ValText, sqinn.ValText, sqinn.ValText, sqinn.ValInt, sqinn.ValText, sqinn.ValText, sqinn.ValInt})
	check(err)
	log.Printf("fetched %d rows", len(rows))
	t3 := time.Now()
	// close and terminate
	check(sq.Close())
	check(sq.Terminate())
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testBlob(sqinnPath, dbFile string) {
	funcname := "testBlob"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s", sqinnPath, dbFile)
	assert := func(c bool, format string, v ...interface{}) {
		if !c {
			panic(fmt.Errorf(format, v...))
			// log.Fatalf(format, v...)
		}
	}
	sq, err := sqinn.New(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	assert(err == nil, "%s", err)
	// open db
	err = sq.Open(dbFile)
	assert(err == nil, "%s", err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS users")
	assert(err == nil, "%s", err)
	_, err = sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, image BLOB)")
	assert(err == nil, "%s", err)
	// insert
	id := 1
	image := make([]byte, 64)
	for i := 0; i < len(image); i++ {
		image[i] = byte(i)
	}
	values := []interface{}{id, image}
	_, err = sq.Exec("INSERT INTO users (id,image) VALUES(?,?)", 1, 2, values)
	assert(err == nil, "%s", err)
	// query
	sql := "SELECT id, image FROM users ORDER BY id"
	rows, err := sq.Query(sql, nil, []byte{sqinn.ValInt, sqinn.ValBlob})
	assert(err == nil, "%s", err)
	assert(len(rows) == 1, "wrong rows %d", len(rows))
	// close and terminate
	err = sq.Close()
	assert(err == nil, "%s", err)
	err = sq.Terminate()
	assert(err == nil, "%s", err)
	log.Printf("TEST %s OK", funcname)
}

func main() {
	// log.SetOutput(ioutil.Discard)
	log.SetFlags(log.Lmicroseconds)
	sqinnPath := "sqinn"
	dbFile := ":memory:"
	flag.StringVar(&sqinnPath, "sqinn", sqinnPath, "name of sqinn executable")
	flag.StringVar(&dbFile, "db", dbFile, "path to db file")
	flag.Parse()
	for _, arg := range flag.Args() {
		if arg == "test" {
			testFunctions(sqinnPath, dbFile, 2)
			testUsers(sqinnPath, dbFile, 2, true)
			testComplex(sqinnPath, dbFile, 2, 2, 2)
			testBlob(sqinnPath, dbFile)
			return
		} else if arg == "bench" {
			testFunctions(sqinnPath, dbFile, 10*1000)
			testUsers(sqinnPath, dbFile, 1000*1000, false)
			testUsers(sqinnPath, dbFile, 1000*1000, true)
			testComplex(sqinnPath, dbFile, 100, 100, 10)
			return
		}
	}
	fmt.Printf("no command, want 'test' or 'bench'\n")
}
