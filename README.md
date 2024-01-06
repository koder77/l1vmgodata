L1VMgodata 0.9.0
================
This is my first go project.
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
csv-export
csv-import
csv-table-export
csv-table-import
erase all
usage
exit
set-link
rem-link
get-links-number
get-link-name
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
save 'test.db'
OK
```

Load:

```
load 'test.db'
OK
```

Close client connection:

```
close
OK
```

Export data base into ".json" file:

```
json-export 'test.json'
```

Import data base from ".json" file:

```
json-import 'test.json'
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
Set links between key values. You can save multiple links between data.
Here are the commands:

Set link:

```
store data :water-n 'water'
OK
store data :water-chem 'H2O' 
OK
set-link :water-n 'water-chem'
OK
```

Get links total number:

```
get-links-number :water-n ''
1
```

Get link index 0:

```
get-link-name :water-n '0'
water-chem
``` 

Remove link:

```
rem-link :water-n 'water-chem'
OK
```

I did add ```web.go``` webinterface for browser.
You can store and load data in the web browser with the commands like above!
The save and load functions use the value entry as the filename!

exit command to quit the database

NEW
===
CSV table export. The data must be stored like this (chem.db):

```
l1vmgodata database
:1-1-substance "water"
:link '0'
:1-2-chemical "H2O"
:link '0'
:1-3-boiling "100"
:link '0'
:2-1-substance "iron"
:link '0'
:2-2-chemical "Fe"
:link '0'
:2-3-boiling "3070"
:link '0'

```
So here 1-1-substance "water" is index 1 key 1.
And the 1-3-boiling is the last key of index 1.

The CSV table looks like this:

```
substance, chemical, boiling
water, H2O, 100
iron, Fe, 3070
```

<b>CSV table import.</b>
Now you can import CSV tables also. The entries are named like in a CSV table export. See above.#
