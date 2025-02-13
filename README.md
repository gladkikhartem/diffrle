# diffrle
Differential Run Length Encoding(RLE) Set written in Go.

It allows to efficiently store a set of sequential IDs, allowing gaps in between. Data is stored as Btree of ranges, where each range is stored as 3 values: `start`, `step` and `count`.  So 995 IDs `1,2,3,4,5....500,555,560,565,...1000` are stored as 2 non-overlapping ranges `1-500 with step of 1` and `555-1000` with step of 5, providing compression ratios proportional to average lenght of the range. 

In RLE duplicates are compressed, but diffrle compresses the duplicate differences between keys.

### Use-cases
This can be useful to store IDs that follow a sequence with equal distances between each other (DB Serial columns, time slots, memory addresses, etc...).

For example instead of storing and updating boolean value "processed" for each row in database in order to implement outbox pattern - one can setup background processing to process IDs sequentially and store all processed/failed IDs in compressed array of few KB in size.

Same goes for queue processing - if every message has serial sequence number - we can efficiently store the list of sent/failed messages, without having to update each message. And messages can be efficiently deleted by dropping whole partitions.

### Performance
```
GOMAXPROCS=1  AMD Ryzen 5 6600H   
-- diffrle set
BenchmarkRandomInsert        	15491312	       1366 ns/op
BenchmarkSequentialInsert    	76867363	       146.5 ns/op
BenchmarkExist               	347426352	       32.86 ns/op

-- go standard map
BenchmarkGoMapRandomInsert     	100000000	       222.0 ns/op
BenchmarkGoMapSequentialInsert 	100000000	       177.0 ns/op
BenchmarkGoMapExist            	100000000	       101.7 ns/op
```

### Usage

``` go
d := NewSet(100) // new set with btree of dimension 100
d.Set(1) // add new ID to set
d.Delete(1) // delete new ID from set
d.DeleteFromTo(0,100) // delete range of IDs
d.Exists(1) // checks if id exists in the set
d.Ranges() // returns underlying ranges
d.IterAll(func(v int64) bool) { // iterate all IDs in range
    //...
    return true
})

d.IterFromTo(0, 100, func(v int64) bool) l{ // iterate all IDs in range [0,100)
    // ...
    return true
})
```



### Use-case: Outbox
For DB tables with serial primary IDs - it's possible to avoid updating status of each row and keep a list of sent/processed IDs in a separate place with few KB in size.
This can be useful in sutation where it's not possible to update original table, such a legacy systems, DBs you don't own or you already have high write-throughput and don't want make another update for each row.


