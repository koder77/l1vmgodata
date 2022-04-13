L1VMgodata
==========
This is my first go project it will be a simple database like L1VM data, if it is finished.
At the moment you can connect via:

<pre>
$ nc localhost 2000
</pre>

And type in some text, which the server then reads.


HELP!
=====
I need some help:
I tried to allocate the data structure:

```
type mem struct {
	s string
	i int64
	d float64
}

type data struct {
	dtype int
	name  string
	size  int64
	mem   mem
}

var maxdata = 10000 // max data number
var pdata *[]data

func main() {
	fmt.Println("l1vmgodata start...")
	servdata := data{}
	pdata = &servdata
	init_data()
	run_server()
}
```

But this doesn't work: I can't assign the servdata to the pdata pointer.
I need the data pointer pdata globally, as it contains the whole database!
How can I do this? I am totally stuck now!
