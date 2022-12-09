package main

// Note: conns is always set to at least one for the modernc driver.

var tests = []testCfg{
	{section: "Read only"},
	{wal: true, readers: 1, sync: "NORMAL", conns: 1},
	{wal: true, readers: 1, sync: "NORMAL", conns: 0},
	{wal: true, readers: 10, sync: "NORMAL", conns: 1},
	{wal: true, readers: 10, sync: "NORMAL", conns: 0},
	{wal: true, readers: 100, sync: "NORMAL", conns: 0},
	{wal: true, readers: 100, sync: "NORMAL", conns: 1},
	{wal: true, readers: 100, sync: "NORMAL", conns: 2},

	{section: "Write only"},
	{wal: false, writers: 100, sync: "NORMAL", conns: 0},
	{wal: true, writers: 100, sync: "FULL", conns: 0},
	{wal: true, writers: 1, sync: "NORMAL", conns: 1},
	{wal: true, writers: 1, sync: "NORMAL", conns: 0},
	{wal: true, writers: 10, sync: "NORMAL", conns: 1},
	{wal: true, writers: 10, sync: "NORMAL", conns: 0},
	{wal: true, writers: 100, sync: "NORMAL", conns: 1},
	{wal: true, writers: 100, sync: "NORMAL", conns: 0},
	{wal: true, writers: 100, sync: "NORMAL", conns: 1, mutex: true},
	{wal: true, writers: 100, sync: "NORMAL", conns: 0, mutex: true},

	{section: "Read Heavy"},
	{wal: true, readers: 100, writers: 10, sync: "NORMAL", conns: 1},
	{wal: true, readers: 100, writers: 10, sync: "NORMAL", conns: 0},
	{wal: true, readers: 100, writers: 10, sync: "NORMAL", conns: 1, mutex: true},
	{wal: true, readers: 100, writers: 10, sync: "NORMAL", conns: 0, mutex: true},

	{section: "Write Heavy"},
	{wal: true, readers: 10, writers: 100, sync: "NORMAL", conns: 1},
	{wal: true, readers: 10, writers: 100, sync: "NORMAL", conns: 0},
	{wal: true, readers: 10, writers: 100, sync: "NORMAL", conns: 1, mutex: true},
	{wal: true, readers: 10, writers: 100, sync: "NORMAL", conns: 0, mutex: true},
}
