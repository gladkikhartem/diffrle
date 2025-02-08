# diffrle
Differential Run Length Encoding Map (DLE) written in Pure Go

It allows to efficiently store a set of sequential IDs that may have gaps in between. Data is stored as Btree of ranges, where each range is stored as 3 values: `start`, `step` and `count`.  So 995 IDs `1,2,3,4,5....500,555,560,565,...1000` are stored as 6 values `1,1,500|555,5,99`.


### Performance
Storing IDs in such a way is ~4 times slower than simply storing them in a map. 

Compression ratio is proportial to the average length of ranges.

Storing, Accessing and Deleting ranges is O(log n) operation, since Btree is used under the hood.


### Usage

```

d := NewSet(1000) // btree dimension 1000
d.Set(1) // add new ID to set
d.Delete(1) // delete new ID from set
d.Exists(1) // checks if id exists in the set
d.Ranges() // returns underlying ranges
d.IterAll(func(v int64) bool { // iterate all IDs in range
    //
    return true
})

```


### Possible use-case: Outbox
For DB tables with serial primary IDs - it's possible to avoid updating status of each row and keep a list of sent/processed IDs in a separate place with few KB in size.