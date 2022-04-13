L1VMgodata
==========
This is my first go project. It will be a simple database like L1VM data, if it is finished.
At the moment you can connect via:

```
$ nc localhost 2000
```

And type in some text, which the server then reads.
TODO: client handling sockets code.

Via nc you can send the "store data" command:

```
store data :test "12345"
OK
get key :test
12345
```