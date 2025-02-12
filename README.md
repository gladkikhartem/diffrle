# diffrle
Differential Run Length Encoding Set written in Pure Go

It allows to efficiently store a set of sequential IDs, allowing gaps in between. Data is stored as Btree of ranges, where each range is stored as 3 values: `start`, `step` and `count`.  So 995 IDs `1,2,3,4,5....500,555,560,565,...1000` are stored as 6 values `1,1,500|555,5,99`, providing incredible compression ratios for cases when IDs are typically follow a sequence with equal distances between each other (DB Serial columns)


### Performance
```
GOMAXPROCS=1  AMD Ryzen 5 6600H   
BenchmarkRandomInsert        	15491312	       1366 ns/op
BenchmarkSequentialInsert    	76867363	       146.5 ns/op
BenchmarkExist               	347426352	       32.86 ns/op
BenchmarkGoMapRandomInsert     	100000000	       222.0 ns/op
BenchmarkGoMapSequentialInsert 	100000000	       177.0 ns/op
BenchmarkGoMapExist            	100000000	       101.7 ns/op
```

### Usage

```

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


