package inject

import (
	"reflect"
)

type Object struct {
	Value interface{}
	Name  string
	typ   reflect.Type
	val   *reflect.Value
}

func NewObject(name string, obj interface{}) *Object {
	val := reflect.ValueOf(obj)
	return &Object{
		Name:  name,
		Value: obj,
		typ:   reflect.TypeOf(obj),
		val:   &val,
	}
}

func (o *Object) AssignableTo(typ reflect.Type) bool {
	return o.typ.AssignableTo(typ)
}

type Repository struct {
	s []*Object
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) register(obj *Object) {
	if obj.Name == "" {
		r.s = append(r.s, obj)
		return
	}
	for _, o := range r.s {
		if o.Name == obj.Name {
			panic(obj.Name + " already inject")
		}
	}
	r.s = append(r.s, obj)
}

func (r *Repository) Register(name string, obj interface{}) {
	if o, ok := obj.(*Object); ok {
		r.register(o)
		return
	}
	r.register(NewObject(name, obj))
}

func (r *Repository) Key(name string) interface{} {
	if name == "" {
		return nil
	}
	for _, obj := range r.s {
		if obj.Name == name {
			return obj.Value
		}
	}
	return nil
}

func (r *Repository) Type(obj interface{}) interface{} {
	var typ reflect.Type
	if t, ok := obj.(reflect.Type); ok {
		typ = t
	} else {
		typ = reflect.TypeOf(obj)
	}
	for _, o := range r.s {
		//fmt.Println(o.typ.Name(), typ.Name())
		if o.AssignableTo(typ) {
			return o.Value
		}
	}
	return nil
}

func (r *Repository) Produce(obj interface{}) interface{} {
	typ := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if typ.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
		typ = val.Type()
	}
	if typ.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			tf := typ.Field(i)
			if tf.Name[0] <= 'A' || tf.Name[0] >= 'Z' {
				continue
			}
			vf := val.Field(i)
			name := tf.Tag.Get("inject")
			if name == "-" {
				continue
			}
			v := r.Key(name)
			if v != nil {
				vf.Set(reflect.ValueOf(v))
				continue
			}
			v = r.Key(tf.Name)
			if v != nil {
				vf.Set(reflect.ValueOf(v))
				continue
			}
			v = r.Type(vf.Type())
			if v != nil {
				vf.Set(reflect.ValueOf(v))
			}
		}
	}
	return obj
}
