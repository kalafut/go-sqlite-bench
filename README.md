# go-sqlite-bench

This benchmark tests Go + SQLite performance using a variety of database and application settings:

- Concurrency level
- Read/Write mix
- WAL, SYNC settings
- Connection pool
- Using a write mutex

## Running

```text
$ go run . -h

  -d string
    	database driver to use (mattn, modernc) (default "mattn")
  -t int
    	per test duration in seconds (default 1)
```

## Findings

The full results are below, but here are a few things that stand out:

- Concurrency reduces overall throughput regardless of test
- Setting `_journal=WAL` and leaving `SYNC` at the default of `NORMAL` (not `FULL`) are standard practice and showed to be indeed essential. Not doing those things is so bad I didn't bother with too many test runs with these set badly.
- Using `SetMaxOpenConns(1)` (as recommended in the [mattn/go-sqlite FAQ](https://github.com/mattn/go-sqlite3#faq)) seems to reduce performance in high concurrency write-only and read-only loads.
- Using `SetMaxOpenConns(1)` significantly increases write performance in mixed read/write loads, at the expense of read performance. It seems to keep the read/write mix fairer (relative to the amount of read/write load).
- Using `SetMaxOpenConns(>0)` seems to be required for the [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) version to handle concurrency.
- Using an application level write mutex:
  - had mixed (but fairly small) effects on write-only loads
  - resulted in many more reads (at the expence of writes) in mixed loads
- At a _very_ rough approximation, the pure Go modernc driver was about 75% as fast as mattn's CGO version.

Those are just some immediate takeways from a few tests. If you have suggestions for how to improve the benchmark, want to add some more settings or test combinations (see `testsuite.go`), please open a PR.

## References

There are a lot of knobs to potentially adjust and this topic has been written about quite a bit:

- [Go and SQLite in the cloud](https://www.golang.dk/articles/go-and-sqlite-in-the-cloud) (also see additional links at the bottom of that article)
- [Hacker News thread prompting this work](https://news.ycombinator.com/item?id=33899007)
- [Similar benchmark](https://gist.github.com/markuswustenberg/f35ab7e191137dca5f7ec112bfc887be?permalink_comment_id=4396598#gistcomment-4396598)

## Results

Run on a 2021 MBP

### Driver: https://github.com/mattn/go-sqlite3

```text
Results are in reads/writes per second

==Read only==
readers=1   writers=0   WAL=Y sync=NORMAL conns=1 mutex=N |   read: 193130, write:      0, total: 193130
readers=1   writers=0   WAL=Y sync=NORMAL conns=0 mutex=N |   read: 192884, write:      0, total: 192884
readers=10  writers=0   WAL=Y sync=NORMAL conns=1 mutex=N |   read: 156680, write:      0, total: 156680
readers=10  writers=0   WAL=Y sync=NORMAL conns=0 mutex=N |   read: 247939, write:      0, total: 247939
readers=100 writers=0   WAL=Y sync=NORMAL conns=0 mutex=N |   read: 205042, write:      0, total: 205042
readers=100 writers=0   WAL=Y sync=NORMAL conns=1 mutex=N |   read: 142733, write:      0, total: 142733
readers=100 writers=0   WAL=Y sync=NORMAL conns=2 mutex=N |   read: 221243, write:      0, total: 221243

==Write only==
readers=0   writers=100 WAL=N sync=NORMAL conns=0 mutex=N |   read:      0, write:   3356, total:   3356
readers=0   writers=100 WAL=Y sync=FULL conns=0 mutex=N   |   read:      0, write:  10902, total:  10902
readers=0   writers=1   WAL=Y sync=NORMAL conns=1 mutex=N |   read:      0, write:  97088, total:  97088
readers=0   writers=1   WAL=Y sync=NORMAL conns=0 mutex=N |   read:      0, write:  96447, total:  96447
readers=0   writers=10  WAL=Y sync=NORMAL conns=1 mutex=N |   read:      0, write:  78676, total:  78676
readers=0   writers=10  WAL=Y sync=NORMAL conns=0 mutex=N |   read:      0, write:  94759, total:  94759
readers=0   writers=100 WAL=Y sync=NORMAL conns=1 mutex=N |   read:      0, write:  73426, total:  73426
readers=0   writers=100 WAL=Y sync=NORMAL conns=0 mutex=N |   read:      0, write:  90776, total:  90776
readers=0   writers=100 WAL=Y sync=NORMAL conns=1 mutex=Y |   read:      0, write:  79885, total:  79885
readers=0   writers=100 WAL=Y sync=NORMAL conns=0 mutex=Y |   read:      0, write:  80414, total:  80414

==Read Heavy==
readers=100 writers=10  WAL=Y sync=NORMAL conns=1 mutex=N |   read: 112366, write:  11330, total: 123697
readers=100 writers=10  WAL=Y sync=NORMAL conns=0 mutex=N |   read: 193550, write:   7811, total: 201362
readers=100 writers=10  WAL=Y sync=NORMAL conns=1 mutex=Y |   read: 141928, write:   1465, total: 143393
readers=100 writers=10  WAL=Y sync=NORMAL conns=0 mutex=Y |   read: 205903, write:   1062, total: 206965

==Write Heavy==
readers=10  writers=100 WAL=Y sync=NORMAL conns=1 mutex=N |   read:   7337, write:  72868, total:  80206
readers=10  writers=100 WAL=Y sync=NORMAL conns=0 mutex=N |   read: 156790, write:  16871, total: 173662
readers=10  writers=100 WAL=Y sync=NORMAL conns=1 mutex=Y |   read: 112584, write:  11395, total: 123980
readers=10  writers=100 WAL=Y sync=NORMAL conns=0 mutex=Y |   read: 184286, write:  10634, total: 194920
```

### Driver: https://modernc.org/sqlite

```
Results are in reads/writes per second

==Read only==
readers=1   writers=0   WAL=Y sync=NORMAL conns=1 mutex=N |   read: 151782, write:      0, total: 151782
readers=1   writers=0   WAL=Y sync=NORMAL conns=0 mutex=N |   read: 147356, write:      0, total: 147356
readers=10  writers=0   WAL=Y sync=NORMAL conns=1 mutex=N |   read: 118524, write:      0, total: 118524
readers=10  writers=0   WAL=Y sync=NORMAL conns=0 mutex=N |   read: 120853, write:      0, total: 120853
readers=100 writers=0   WAL=Y sync=NORMAL conns=0 mutex=N |   read: 121335, write:      0, total: 121335
readers=100 writers=0   WAL=Y sync=NORMAL conns=1 mutex=N |   read: 120051, write:      0, total: 120051
readers=100 writers=0   WAL=Y sync=NORMAL conns=2 mutex=N |   read: 180759, write:      0, total: 180759

==Write only==
readers=0   writers=100 WAL=N sync=NORMAL conns=0 mutex=N |   read:      0, write:   3975, total:   3975
readers=0   writers=100 WAL=Y sync=FULL conns=0 mutex=N   |   read:      0, write:  13320, total:  13320
readers=0   writers=1   WAL=Y sync=NORMAL conns=1 mutex=N |   read:      0, write:  77480, total:  77480
readers=0   writers=1   WAL=Y sync=NORMAL conns=0 mutex=N |   read:      0, write:  78759, total:  78759
readers=0   writers=10  WAL=Y sync=NORMAL conns=1 mutex=N |   read:      0, write:  66706, total:  66706
readers=0   writers=10  WAL=Y sync=NORMAL conns=0 mutex=N |   read:      0, write:  65631, total:  65631
readers=0   writers=100 WAL=Y sync=NORMAL conns=1 mutex=N |   read:      0, write:  65836, total:  65836
readers=0   writers=100 WAL=Y sync=NORMAL conns=0 mutex=N |   read:      0, write:  64890, total:  64890
readers=0   writers=100 WAL=Y sync=NORMAL conns=1 mutex=Y |   read:      0, write:  66165, total:  66165
readers=0   writers=100 WAL=Y sync=NORMAL conns=0 mutex=Y |   read:      0, write:  66043, total:  66043

==Read Heavy==
readers=100 writers=10  WAL=Y sync=NORMAL conns=1 mutex=N |   read:  95886, write:   9637, total: 105523
readers=100 writers=10  WAL=Y sync=NORMAL conns=0 mutex=N |   read:  96570, write:   9831, total: 106401
readers=100 writers=10  WAL=Y sync=NORMAL conns=1 mutex=Y |   read: 115469, write:   1172, total: 116642
readers=100 writers=10  WAL=Y sync=NORMAL conns=0 mutex=Y |   read: 120202, write:   1251, total: 121453

==Write Heavy==
readers=10  writers=100 WAL=Y sync=NORMAL conns=1 mutex=N |   read:   6173, write:  61083, total:  67257
readers=10  writers=100 WAL=Y sync=NORMAL conns=0 mutex=N |   read:   6459, write:  62482, total:  68941
readers=10  writers=100 WAL=Y sync=NORMAL conns=1 mutex=Y |   read:  94716, write:   9413, total: 104130
readers=10  writers=100 WAL=Y sync=NORMAL conns=0 mutex=Y |   read:  96367, write:   9755, total: 106122
```
