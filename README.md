L1VMgodata
==========
This is my first go project. It will be a simple database like L1VM data, if it is finished.
This database is written for data exchange between programs.
At the moment there is no save/read to disk function.
So the data is only stored in RAM!
At the moment you can connect via:

```
$ nc localhost 2000
```

Via nc you can send the "store data" command:

```
store data :test "12345"
OK
get key :test
12345
```

The commands are:

```
store data
get key
get value
remove
close
```