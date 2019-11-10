# errorstats [![GoDoc](https://godoc.org/github.com/mattes/errorstats?status.svg)](https://godoc.org/github.com/mattes/errorstats)

Store counters for hashed errors.

1. Create a new `errorstats.New()` instance.
2. `SetEncoder(T, EncoderFunc)` to encode/hash error into human read-able string.
3. `Log(T)` error (or any other type).
4. Print counters with `Pretty()`.

See [example](https://godoc.org/github.com/mattes/errorstats#example-Stats-JSON).

