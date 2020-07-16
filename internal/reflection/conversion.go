package reflection

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrCannotConvert = errors.New("cannot convert")
)

// Convert converts value to the specified type.
//
// If isShallow, this does not convert deepest elements.
func Convert(v interface{}, t reflect.Type, isShallow bool) (reflect.Value, error) {
	return convert(v, t, isShallow)
}

func convert(v interface{}, t reflect.Type, isShallow bool) (reflect.Value, error) {
	return newConverter(v, t, isShallow).convert()
}

type (
	converter struct {
		v         interface{}
		t         reflect.Type
		isShallow bool
	}
)

func newConverter(v interface{}, t reflect.Type, isShallow bool) *converter {
	return &converter{
		v:         v,
		t:         t,
		isShallow: isShallow,
	}
}

func (s *converter) convert() (ret reflect.Value, err error) {
	defer func() {
		if e := recover(); e != nil {
			ret = reflect.Zero(s.t)
			err = fmt.Errorf("%w %v", ErrCannotConvert, e)
		}
	}()
	switch s.t.Kind() {
	case reflect.Array:
		return s.convertArray()
	case reflect.Slice:
		return s.convertSlice()
	case reflect.Chan:
		return s.convertChan()
	case reflect.Map:
		return s.convertMap()
	default:
		if s.isShallow {
			return s.valueOf(), nil
		}
		return s.valueOf().Convert(s.t), nil
	}
}

func (s *converter) valueOf() reflect.Value { return reflect.ValueOf(s.v) }

func (s *converter) convertChan() (reflect.Value, error) {
	return reflect.MakeChan(s.t, s.valueOf().Cap()), nil
}

func (s *converter) convertArray() (reflect.Value, error) {
	var (
		sv = s.valueOf()
		p  = reflect.New(s.t)
	)
	for i := 0; i < p.Len(); i++ {
		v, err := convert(sv.Index(i).Interface(), s.t.Elem(), s.isShallow)
		if err != nil {
			return reflect.Zero(s.t), err
		}
		p.Index(i).Set(v)
	}
	return p, nil
}

func (s *converter) convertSlice() (reflect.Value, error) {
	var (
		sv = s.valueOf()
		p  = reflect.MakeSlice(s.t, sv.Len(), sv.Len())
	)
	for i := 0; i < p.Len(); i++ {
		v, err := convert(sv.Index(i).Interface(), s.t.Elem(), s.isShallow)
		if err != nil {
			return reflect.Zero(s.t), err
		}
		p.Index(i).Set(v)
	}
	return p, nil
}

func (s *converter) convertMap() (reflect.Value, error) {
	var (
		sv = s.valueOf()
		p  = reflect.MakeMapWithSize(s.t, sv.Len())
		it = sv.MapRange()
	)
	for it.Next() {
		k, v := it.Key(), it.Value()
		ck, err := convert(k.Interface(), s.t.Key(), s.isShallow)
		if err != nil {
			return reflect.Zero(s.t), err
		}
		cv, err := convert(v.Interface(), s.t.Elem(), s.isShallow)
		if err != nil {
			return reflect.Zero(s.t), err
		}
		p.SetMapIndex(ck, cv)
	}
	return p, nil
}
