package errorstats

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// EncoderFunc translates v into string.
// v is guaranteed to always receive variable value, never
// a pointer to value, in other words, *v is automatically converted into v.
// Its ok to return an empty string.
type EncoderFunc func(v interface{}) string

type Stats struct {
	counters   map[string]uint64
	countersMu sync.RWMutex

	keyFuncsByType   map[reflect.Type]EncoderFunc
	keyFuncsByString map[string]EncoderFunc // by string allows to intercept private errors
	keyFuncsMu       sync.RWMutex
}

// New returns a new Stats instance
func New() *Stats {
	s := &Stats{
		counters:         make(map[string]uint64),
		keyFuncsByType:   make(map[reflect.Type]EncoderFunc),
		keyFuncsByString: make(map[string]EncoderFunc),
	}

	return s
}

// SetEncoder sets a EncoderFunc for a typ.
// typ can be given as string, which allows to intercept private structs and errors.
func (s *Stats) SetEncoder(typ interface{}, fn EncoderFunc) {
	s.keyFuncsMu.Lock()

	if str, ok := typ.(string); ok {
		s.keyFuncsByString[str] = fn
	} else {
		s.keyFuncsByType[typeOf(typ)] = fn
	}

	s.keyFuncsMu.Unlock()
}

// DeleteEncoder removes EncoderFunc set for typ
func (s *Stats) DeleteEncoder(typ interface{}) {
	s.keyFuncsMu.Lock()

	if str, ok := typ.(string); ok {
		delete(s.keyFuncsByString, str)
	} else {
		delete(s.keyFuncsByType, typeOf(typ))
	}

	s.keyFuncsMu.Unlock()
}

// Log increases counter for v
func (s *Stats) Log(v interface{}) {
	if v == nil {
		return
	}

	key := s.Visit("", v)
	if key == "" {
		return
	}

	s.countersMu.Lock()
	s.counters[key]++
	s.countersMu.Unlock()
}

// Visit visits v's (in order provided) and if v != nil, it will
// append v's key to key.
func (s *Stats) Visit(key string, v ...interface{}) string {
	for _, x := range v {
		if x == nil {
			continue
		}

		if key != "" {
			key += "/ "
		}

		// set type of x and the value of it
		typ := reflect.TypeOf(x)
		val := x

		// if is pointer, use value instead
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			val = reflect.ValueOf(x).Elem().Interface()
		}

		// call encoder func for typ or if not available, just use typ's name.
		s.keyFuncsMu.RLock()

		if fn, ok := s.keyFuncsByType[typ]; ok {
			key += fn(val)

		} else if fn, ok := s.keyFuncsByString[typ.String()]; ok {
			key += fn(val)

		} else {
			key += typ.String()
		}

		s.keyFuncsMu.RUnlock()
	}

	return key
}

// Reset resets counters
func (s *Stats) Reset() {
	s.countersMu.Lock()
	s.counters = make(map[string]uint64)
	s.countersMu.Unlock()
}

// Err returns an error if any errors have been logged, otherwise nil.
func (s *Stats) Err() error {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()
	if len(s.counters) == 0 {
		return nil
	}

	return errors.New(s.json())
}

// ErrAndReset atomically returns an error and resets them
// if any errors have been logged, otherwise nil.
func (s *Stats) ErrAndReset() error {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	if len(s.counters) == 0 {
		return nil
	}

	err := errors.New(s.json())
	s.counters = make(map[string]uint64)
	return err
}

// String returns stats
func (s *Stats) String() string {
	return s.JSON()
}

// String returns stats as JSON
func (s *Stats) JSON() string {
	s.countersMu.RLock()
	defer s.countersMu.RUnlock()
	return s.json()
}

// json returns stats as JSON, it must be safeguarded
// with a read lock on s.countersMu
func (s *Stats) json() string {
	buf := bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	e.SetEscapeHTML(false)
	if err := e.Encode(s.counters); err != nil {
		// json marshal can return UnsupportedTypeError and UnsupportedValueError,
		// both should not happen based on the input we are marshalling.
		panic(err)
	}

	return buf.String()
}

type PrettyFormat int

const (
	CounterDescFormat PrettyFormat = iota + 1
	KeyAscFormat
)

// Pretty returns stats in a really pretty format, useful for console output.
func (s *Stats) Pretty(f PrettyFormat) string {
	counters := s.counterSlice()

	// decide how to sort counter slice
	switch f {
	default:
		fallthrough
	case CounterDescFormat:
		counters.sortByCounterDesc()

	case KeyAscFormat:
		counters.sortByKeyAsc()
	}

	// max string length of counters
	max := uint64(0)
	for _, c := range counters {
		if max < c.counter {
			max = c.counter
		}
	}
	maxLen := len(strconv.FormatUint(max, 10))

	// prepare output string with padded counters
	out := ""
	for _, c := range counters {
		cstr := strconv.FormatUint(c.counter, 10)
		out += strings.Repeat(" ", maxLen-len(cstr)) + cstr + " " + c.key + "\n"
	}

	return out
}

type counterSlice []counter

type counter struct {
	key     string
	counter uint64
}

// counterSlice turns map[string]uint64 into []counter
func (s *Stats) counterSlice() counterSlice {
	s.countersMu.RLock()
	counters := make([]counter, 0, len(s.counters))
	for k, c := range s.counters {
		counters = append(counters, counter{key: k, counter: c})
	}
	s.countersMu.RUnlock()
	return counters
}

func (c counterSlice) sortByCounterDesc() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].counter > c[j].counter
	})
}

func (c counterSlice) sortByKeyAsc() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].key < c[j].key
	})
}

// typeOf returns type of v, unwrapping *v if necessary.
func typeOf(v interface{}) reflect.Type {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		return typ.Elem()
	}
	return typ
}
