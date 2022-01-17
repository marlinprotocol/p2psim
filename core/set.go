package core

// Generic set data structure with a simple and intuitive interface

type Set struct {
	elems map[interface{}]struct{}
}

func NewSet(elems ...interface{}) *Set {
	set := map[interface{}]struct{}{}
	for _, elem := range elems {
		set[elem] = struct{}{}
	}
	return &Set{
		elems: set,
	}
}

func (set *Set) Add(elems ...interface{}) {
	for _, elem := range elems {
		set.elems[elem] = struct{}{}
	}
}

func (set *Set) Remove(elems ...interface{}) {
	for _, elem := range elems {
		delete(set.elems, elem)
	}
}

func (set *Set) Exists(elem interface{}) bool {
	_, ok := set.elems[elem]
	return ok
}

func (set *Set) Len() int {
	return len(set.elems)
}

func (set *Set) Clear() {
	set.elems = map[interface{}]struct{}{}
}

func (set *Set) Flatten() []interface{} {
	elems := []interface{}{}
	for elem := range set.elems {
		elems = append(elems, elem)
	}
	return elems
}

func (set *Set) Traverse(visitor func(elem interface{})) {
	for elem := range set.elems {
		visitor(elem)
	}
}
