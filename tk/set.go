package tk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Set map[interface{}]struct{}

// An OrderedPair represents a 2-tuple of values.
type OrderedPair struct {
	First  interface{}
	Second interface{}
}

type Iterator struct {
	C    <-chan interface{}
	stop chan struct{}
}

func (i *Iterator) Stop() {
	defer func() {
		_ = recover()
	}()

	close(i.stop)

	for range i.C {
	}
}

func newIterator() (*Iterator, chan<- interface{}, <-chan struct{}) {
	itemChan := make(chan interface{})
	stopChan := make(chan struct{})
	return &Iterator{
		C:    itemChan,
		stop: stopChan,
	}, itemChan, stopChan
}

func NewSet(elts ...interface{}) Set {
	s := make(Set)
	if len(elts) == 1 {
		tp := reflect.TypeOf(elts[0]).Kind()
		if tp == reflect.Slice || tp == reflect.Array {
			v := reflect.ValueOf(elts[0])
			for i := 0; i < v.Len(); i++ {
				s.Add(v.Index(i).Interface())
			}
		}
	} else {
		for _, e := range elts {
			s.Add(e)
		}
	}
	return s
}

func (pair *OrderedPair) Equal(other OrderedPair) bool {
	if pair.First == other.First &&
		pair.Second == other.Second {
		return true
	}

	return false
}

func (set Set) Add(i interface{}) bool {
	_, found := set[i]
	if found {
		return false
	}

	set[i] = struct{}{}
	return true
}

func (set Set) Contains(i ...interface{}) bool {
	for _, val := range i {
		if _, ok := set[val]; !ok {
			return false
		}
	}
	return true
}

func (set Set) IsSubset(other Set) bool {
	if set.Len() > other.Len() {
		return false
	}
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set Set) IsProperSubset(other Set) bool {
	return set.IsSubset(other) && !set.Equal(other)
}

func (set Set) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set Set) IsProperSuperset(other Set) bool {
	return set.IsSuperset(other) && !set.Equal(other)
}

func (set Set) Union(other Set) Set {
	unionedSet := NewSet()

	for elem := range set {
		unionedSet.Add(elem)
	}
	for elem := range other {
		unionedSet.Add(elem)
	}
	return unionedSet
}

func (set Set) Intersect(other Set) Set {
	intersection := NewSet()
	if set.Len() < other.Len() {
		for elem := range set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range other {
			if set.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return intersection
}

func (set Set) Diff(other Set) Set {
	difference := NewSet()
	for elem := range set {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return difference
}

func (set Set) SymmetricDiff(other Set) Set {
	aDiff := set.Diff(other)
	bDiff := other.Diff(set)
	return aDiff.Union(bDiff)
}

func (set Set) Clear() {
	for elem := range set {
		delete(set, elem)
	}
}

func (set Set) Remove(i interface{}) {
	delete(set, i)
}

func (set Set) Sort() {
	s := set.ToSlice()
	quickSortInterface(s, 0, len(s)-1)
	set.Clear()
	for _, elem := range s {
		set.Add(elem)
	}
}

func (set Set) Len() int {
	return len(set)
}

func (set Set) Each(fn func(interface{}) bool) {
	for elem := range set {
		if fn(elem) {
			break
		}
	}
}

func (set Set) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		for elem := range set {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

func (set Set) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
	L:
		for elem := range set {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
	}()

	return iterator
}

func (set Set) Equal(other Set) bool {
	if set.Len() != other.Len() {
		return false
	}
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set Set) Clone() Set {
	clonedSet := NewSet()
	for elem := range set {
		clonedSet.Add(elem)
	}
	return clonedSet
}

func (set Set) String() string {
	items := make([]string, 0, len(set))

	for elem := range set {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(items, ", "))
}

func (pair OrderedPair) String() string {
	return fmt.Sprintf("(%v, %v)", pair.First, pair.Second)
}

func (set Set) Pop() interface{} {
	for item := range set {
		delete(set, item)
		return item
	}
	return nil
}

func (set Set) PowerSet() Set {
	powSet := NewSet()
	nullset := NewSet()
	powSet.Add(&nullset)

	for es := range set {
		u := NewSet()
		j := powSet.Iter()
		for er := range j {
			p := NewSet()
			if reflect.TypeOf(er).Name() == "" {
				k := er.(Set)
				for ek := range k {
					p.Add(ek)
				}
			} else {
				p.Add(er)
			}
			p.Add(es)
			u.Add(&p)
		}

		powSet = powSet.Union(u)
	}

	return powSet
}

func (set Set) CartesianProduct(other Set) Set {
	cartProduct := NewSet()

	for i := range set {
		for j := range other {
			elem := OrderedPair{First: i, Second: j}
			cartProduct.Add(elem)
		}
	}

	return cartProduct
}

func (set Set) ToSlice() []interface{} {
	keys := make([]interface{}, 0, set.Len())
	for elem := range set {
		keys = append(keys, elem)
	}

	return keys
}

// Marshal creates a JSON array from the set, it marshals all elements
func (set Set) Marshal() ([]byte, error) {
	items := make([]string, 0, set.Len())

	for elem := range set {
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}

		items = append(items, string(b))
	}

	return []byte(fmt.Sprintf("[%s]", strings.Join(items, ","))), nil
}

// Unmarshal recreates a set from a JSON array, it only decodes primitive types. Numbers are decoded as json.Number.
func (set Set) Unmarshal(b []byte) error {
	var i []interface{}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	for _, v := range i {
		switch t := v.(type) {
		case []interface{}, map[string]interface{}:
			continue
		default:
			set.Add(t)
		}
	}

	return nil
}
