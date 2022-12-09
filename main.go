package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	seedCount = 100000
	dataLen   = 30
)

var dbWriteMutex sync.Mutex

var runtime = flag.Int64("t", 1, "per test duration in seconds")

type testCfg struct {
	// section title
	section string

	// sqlite3 options
	wal   bool
	sync  string
	conns int

	// app options
	writers int
	readers int
	mutex   bool
}

func (cfg testCfg) String() string {
	yn := func(b bool) string {
		if b {
			return "Y"
		}
		return "N"
	}

	return fmt.Sprintf("readers=%-3d writers=%-3d WAL=%s sync=%s conns=%d mutex=%s",
		cfg.readers, cfg.writers, yn(cfg.wal), cfg.sync, cfg.conns, yn(cfg.mutex))
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(65 + rand.Intn(26))
	}
	return string(b)
}

// initDB creates a new sqlite3 database based on the cfg parameters
func initDB(cfg testCfg) (*sql.DB, string, error) {
	dbFilename := filepath.Join(os.TempDir(), randString(10)+".db")

	var db *sql.DB
	var err error

	// ref https://github.com/mattn/go-sqlite3#connection-string
	dsn := dbFilename + "?_timeout=10000&"
	dsn += "_sync=" + cfg.sync + "&"

	if cfg.wal {
		dsn += "_journal_mode=WAL&"
	}

	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}

	if cfg.conns > 0 {
		db.SetMaxOpenConns(cfg.conns)
	}

	_, err = db.Exec("CREATE TABLE foo (id INTEGER NOT NULL PRIMARY KEY, name TEXT)")
	if err != nil {
		panic(err)
	}

	if err := seedDB(db); err != nil {
		panic(err)
	}

	return db, dbFilename, nil
}

// seedDB inserts seedCount rows for use by readTest
func seedDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO FOO(name) VALUES($1)")
	if err != nil {
		return err
	}

	for i := 0; i < seedCount; i++ {
		_, err = stmt.Exec(randString(dataLen))
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// readTest reads as many rows as possible in the given time limit, and returns the number of reads
func readTest(db *sql.DB, doneCh <-chan struct{}) (int, error) {
	var reads int

	for {
		select {
		case <-doneCh:
			return reads, nil
		default:
		}

		i := rand.Int31n(int32(seedCount))
		rows, err := db.Query("SELECT * FROM foo WHERE id=$1", i)
		if err != nil {
			return 0, err
		}
		for rows.Next() {
			var id int
			var name string
			err = rows.Scan(&id, &name)
			if err != nil {
				return 0, err
			}
		}
		reads++
	}
}

// writeTest writes as many rows as possible in the given time limit, and returns the number of writes
func writeTest(db *sql.DB, cfg testCfg, doneCh <-chan struct{}) (int, error) {
	var writes int

	for {
		select {
		case <-doneCh:
			return writes, nil
		default:
		}

		s := randString(dataLen)
		if cfg.mutex {
			dbWriteMutex.Lock()
		}

		_, err := db.Exec("INSERT INTO FOO(name) VALUES($1)", s)

		if cfg.mutex {
			dbWriteMutex.Unlock()
		}

		if err != nil {
			return 0, err
		}
		writes++
	}
}

func runTest(cfg testCfg) {
	db, dbFilename, err := initDB(cfg)
	if err != nil {
		log.Fatal("initDB:", err)
	}
	defer db.Close()
	defer os.Remove(dbFilename)

	startCh := make(chan struct{})
	doneCh := make(chan struct{})

	var totalReads int64
	var totalWrites int64

	// Prepare read goroutines
	var wg sync.WaitGroup
	for i := 0; i < cfg.readers; i++ {
		wg.Add(1)
		go func(i int) {
			<-startCh
			reads, err := readTest(db, doneCh)
			if err != nil {
				log.Fatal("readTest:", err)
			}
			atomic.AddInt64(&totalReads, int64(reads))
			wg.Done()
		}(i)
	}

	// Prepare write goroutines
	for i := 0; i < cfg.writers; i++ {
		wg.Add(1)
		go func(i int) {
			<-startCh
			writes, err := writeTest(db, cfg, doneCh)
			if err != nil {
				log.Fatal("writeTest:", err)
			}
			atomic.AddInt64(&totalWrites, int64(writes))
			wg.Done()
		}(i)
	}

	// start goroutines and time limiter
	close(startCh)
	time.AfterFunc(time.Duration(*runtime)*time.Second, func() {
		close(doneCh)
	})

	wg.Wait()

	readRate := totalReads / *runtime
	writeRate := totalWrites / *runtime
	totalRate := (totalReads + totalWrites) / *runtime
	fmt.Printf("%-57s |   read:%7d, write:%7d, total:%7d\n", cfg, readRate, writeRate, totalRate)
}

func runTests(tests []testCfg) {
	fmt.Print("Results are in reads/writes per second\n")
	for _, test := range tests {
		if test.section != "" {
			fmt.Printf("\n==%s==\n", test.section)
			continue
		}
		runTest(test)
	}
}

func main() {
	flag.Parse()
	runTests(tests)
}
