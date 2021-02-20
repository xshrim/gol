package tk

import (
	"encoding/json"
	"reflect"
)

type link struct {
	v interface{}
	n *link
	p *link
}

type orderedMap struct {
	m map[interface{}]interface{}
	h *link
	t *link
}

func NewMap(elts ...interface{}) *orderedMap {
	m := &orderedMap{
		m: make(map[interface{}]interface{}),
		h: nil,
		t: nil,
	}

	if len(elts) == 1 && reflect.TypeOf(elts[0]).Kind() == reflect.Map {
		v := reflect.ValueOf(elts[0])
		for _, e := range v.MapKeys() {
			m.Set(e.Interface(), v.MapIndex(e).Interface())
		}
	}

	return m
}

func (om *orderedMap) Set(key, value interface{}) {
	om.m[key] = value
	var prev *link
	cur := om.h
	for {
		if cur == nil {
			break
		}
		i := compareInterface(cur.v, key)
		if i >= 0 {
			break
		} else {
			prev = cur
			cur = cur.n
		}
	}

	var node *link
	if prev == nil {
		node = &link{v: key, n: cur}
		om.h = node
	} else {
		node = &link{v: key, n: cur, p: prev}
		prev.n = node
	}

	if cur == nil {
		om.t = node
	} else {
		cur.p = node
	}
}

func (om *orderedMap) Remove(key interface{}) {
	delete(om.m, key)
	var prev *link
	cur := om.h
	for {
		if cur == nil {
			return
		}
		i := compareInterface(cur.v, key)
		if i == 0 {
			break
		} else if i > 0 {
			return
		} else {
			prev = cur
			cur = cur.n
		}
	}

	if prev == nil {
		om.h = cur.n
	} else {
		prev.n = cur.n
	}

	if cur.n != nil {
		cur.n.p = prev
	} else {
		om.t = cur.p
	}
}

func (om *orderedMap) Clear() {
	om.m = make(map[interface{}]interface{})
	om.h = nil
	om.t = nil
}

func (om *orderedMap) Keys(reverse ...bool) []interface{} {
	keys := []interface{}{}
	if len(reverse) > 0 && reverse[0] {
		for p := om.t; p != nil; p = p.p {
			keys = append(keys, p.v)
		}
	} else {
		for p := om.h; p != nil; p = p.n {
			keys = append(keys, p.v)
		}
	}

	return keys
}

func (om *orderedMap) Values() []interface{} {
	values := []interface{}{}
	for _, v := range om.m {
		values = append(values, v)
	}

	return values
}

func (om *orderedMap) Contain(k interface{}) bool {
	_, ok := om.m[k]
	return ok
}

func (om *orderedMap) Get(k interface{}) interface{} {
	return om.m[k]
}

func (om *orderedMap) Len() int {
	return len(om.m)
}

func (om *orderedMap) Pop(dequeue ...bool) (interface{}, interface{}) {
	var p *link
	if len(dequeue) > 0 && dequeue[0] {
		p = om.t
	} else {
		p = om.h
	}

	if p == nil {
		return nil, nil
	}

	key := p.v
	value := om.m[key]

	om.Remove(key)

	return key, value
}

func (om *orderedMap) Iter(reverse ...bool) <-chan [2]interface{} {
	r := false
	if len(reverse) > 0 && reverse[0] {
		r = true
	}

	c := make(chan [2]interface{})
	p := om.h
	if r {
		p = om.t
	}

	go func() {
		for p != nil {
			c <- [2]interface{}{p.v, om.m[p.v]}
			if r {
				p = p.p
			} else {
				p = p.n
			}
		}
		close(c)
	}()

	return c
}

func (om *orderedMap) Clone() *orderedMap {
	nm := NewMap()
	for kv := range om.Iter() {
		nm.Set(kv[0], kv[1])
	}

	return nm
}

func (om *orderedMap) Equal(other *orderedMap) bool {
	if om.Len() != other.Len() {
		return false
	}

	return reflect.DeepEqual(om.m, other.m)

	// for k, v := range om.m {
	// 	if mv, ok := mp.m[k]; !ok || compareInterface(v, mv) != 0 {
	// 		return false
	// 	}
	// }

	// return true
}

func (om *orderedMap) Join(mp *orderedMap) {
	for kv := range mp.Iter() {
		om.Set(kv[0], kv[1])
	}
}

func (om *orderedMap) Unmarshal(js string) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(js), &m); err == nil {
		om.Clear()
		for k, v := range m {
			om.Set(k, v)
		}
	}
}

func (om *orderedMap) Marshal() string {
	m := make(map[string]interface{})
	for k, v := range om.m {
		if key, ok := k.(string); ok {
			m[key] = v
		}
	}
	js, _ := json.Marshal(m)
	return string(js)
}
