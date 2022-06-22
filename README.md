L1VMgodata
==========
This is my first go project. It will be a simple database like L1VM data, if it is finished.
This database is written for data exchange between programs.
New save/load function added to write/read whole database from disk!
NEW: added IP whitelist: "whitelist.txt" in which you can set all allowed client IPs!

NEW: added "json-export" command to export the data base into a ".json" file.
NEW: now single quotes: ' are used for data!
You can connect via:

```
$ nc localhost 2000
```

Via nc you can send the "store data" command:

```
store data :test '12345'
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
save
load
json-export
erase all
```

Get key/remove:

```
get key :foobar
test 1234
```

Get value:

```
get value 'test 1234'
foobar
```

Save example:

```
save :save 'test.db'
OK
```

Load:

```
load :load 'test.db'
OK
```

Close client connection:

```
close
OK
```

Export data base into ".json" file:

```
json-export :json 'test.json'
```

Erase all data entries. Handle with care:

```
erase all
OK
```
