L1VMgodata
==========
This is my first go project. It will be a simple database like L1VM data, if it is finished.
This database is written for data exchange between programs.
New save/load function added to write/read whole database from disk!
NEW: added IP whitelist: "whitelist.txt" in which you can set all allowed client IPs!

NEW: added "json-export" command to export the data base into a ".json" file.
NEW: now single quotes: ' are used for data!
NEW: added Json import!
NEW: now you have to set two ports. The first one is for direct access with a program like nc or other programs.
The second is for the new web browser formular to save/load data.

You can connect via:

```
$ nc localhost 2000 2001
```

And in your web browser:

```
127.0.0.1:2001
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
json-import
erase all
usage
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

Import data base from ".json" file:

```
json-import :json 'test.json'
```


Erase all data entries. Handle with care:

```
erase all
OK
```

Get data usage:

```
usage
USAGE 0.00% : 1 of 10000
```

NEW
===
I did add ```web.go``` webinterface for browser.
You can store and load data in the web browser with the commands like above!
The save and load functions use the value entry as the filename!

