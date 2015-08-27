package summer

import (
	"container/list"
	"fmt"
	"reflect"
)

type populateItem struct {
	bean interface{}
	beanType reflect.Type
	beanValue reflect.Value
	structPtr bool
	autowired bool
}

type Context struct {
	items *list.List
	itemsMap map[string] *populateItem
}

func PrintStruct(something interface{}) {
	s := reflect.ValueOf(something).Elem()
	typeOfT := s.Type()
	fmt.Println(">> TypeOf:", reflect.TypeOf(something))
	for i := 0; i < s.NumField(); i++ {
	    f := s.Field(i)

	    if f.CanInterface() {
	    	fmt.Printf(">>   %d: %s %s = %v `%s`\n", i, 
	    		typeOfT.Field(i).Name,
	    		f.Type(),
	    		f.Interface(),
	    		typeOfT.Field(i).Tag)
	    } else {
	    	fmt.Printf(">>   %d: %s %s `%s`\n", i, 
	    		typeOfT.Field(i).Name, 
	    		f.Type(),
	    		typeOfT.Field(i).Tag)
	    }
	}
}

func (ctx *Context) addBean(bean interface{}) *populateItem {
	item := new(populateItem)
	item.bean = bean
	item.autowired = false
	item.beanType = reflect.TypeOf(bean)
	item.beanValue = reflect.ValueOf(bean)
	item.structPtr = item.beanType.Kind() == reflect.Ptr && item.beanType.Elem().Kind() == reflect.Struct

	ctx.items.PushBack(item)

	return item;
}

func (ctx *Context) Add(beans ...interface{}) {
	for _, bean := range beans {
		ctx.addBean(bean);
	}
}

func (ctx *Context) AddWithName(beanName string, bean interface{}) bool {
	if _, notfound := ctx.iGetByName(beanName); notfound {
		ctx.itemsMap[beanName] = ctx.addBean(bean)
		return true
	} else {
		fmt.Println("Duplicate Key:", beanName);
		return false
	}
}

func (ctx *Context) assignable(item *populateItem, modelType reflect.Type) bool {
	switch modelType.Kind() {
	case reflect.Interface:
		if item.beanType.Implements(modelType) {
			return true
		}

	case reflect.Struct:
		if item.beanType.Elem() == modelType {
			return true
		}
	}
	return false
}

func (ctx *Context) iGet(modelType reflect.Type) (*populateItem, int) {
	var rc *populateItem = nil
	var autowiredRc *populateItem = nil
	matched := 0

	for e := ctx.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*populateItem)
		
		duplicate := false

		switch modelType.Kind() {
		case reflect.Interface:
			if item.beanType.Implements(modelType) {
				matched++
				if item.autowired {
					autowiredRc = e.Value.(*populateItem)
				}

				if rc == nil {
					rc = e.Value.(*populateItem);
				} else {
					duplicate = true
				}
			}

		case reflect.Struct:
			if item.beanType.Elem() == modelType {
				matched++
				if item.autowired {
					autowiredRc = e.Value.(*populateItem)
				}
				if rc == nil {
					rc = e.Value.(*populateItem)
				} else {
					duplicate = true
				}
			}
		}

		if duplicate {
			if (matched == 2) {
				fmt.Println("Multiple match for [", modelType, "]:")
				fmt.Println(">> ", rc.beanType);
			}
			fmt.Println(">> ", e.Value.(*populateItem).beanType);
		}

	}

	return autowiredRc, matched
}

func (ctx *Context) Get(intf interface{}) interface{} {
	if item, matched := ctx.iGet(reflect.TypeOf(intf).Elem()); item != nil {
		if matched > 1 {
			return nil
		} else {
			return item.bean
		}
	} else {
		return nil
	}
}

func (ctx *Context) iGetByName(beanName string) (*populateItem,bool) {
	if item, ok := ctx.itemsMap[beanName]; ok {
		if item.autowired {
			return item, false
		} else {
			return nil, false
		}
	} else {
		return nil, true
	}
}

func (ctx *Context) GetByName(beanName string) interface{} {
	if item, err := ctx.iGetByName(beanName); err {
		return nil
	} else if item != nil {
		return item.bean
	} else {
		return nil
	}
}

func (ctx *Context) autowireFieldByX(item *populateItem, setvalue bool, index int, byType bool, tag string) (bool,bool) {
	settable := false
	err := false

	st := item.beanValue.Elem();
	field := st.Field(index)
	f := st.Type().Field(index)

	var match *populateItem

	if (byType) {
		cnt := 0;
		if match, cnt = ctx.iGet(f.Type.Elem()); cnt > 1 {
			match, err = nil, true
		}
	} else {
		if match, err = ctx.iGetByName(tag); match != nil {
			if ! ctx.assignable(match, f.Type.Elem()) {
				fmt.Println ("Autowiring by Name [", tag, "], but not match by type: ",
					f.Type, "vs", match.beanType)
				match, err = nil, true
			}
		}
	}

	if match != nil {
		setterName := "Set" + f.Name

		setter := item.beanValue.MethodByName(setterName)

		if (setter.IsValid()) {
			settable = true

			if setvalue {
				setter.Interface().(func (interface {}))(match.bean)
				//fmt.Println("Autowire", f.Type, "via", setterName)
			}
		} else if field.CanSet() {
			settable = true

			if setvalue {
				field.Set(match.beanValue)
				//fmt.Println("Autowire", f.Type, "via field.Set")
			}
		} else {
			PrintStruct(item.bean)
			fmt.Println(f.Type, ": No setter or field not settable!")
			err = true
		}
	}

	return settable, err
}

func (ctx *Context) doAutowire(item *populateItem, setvalue bool) (int, int, bool) {
	wireableCounter := 0
	totalCounter := 0

	st := item.beanValue.Elem();

    for i := 0; i < st.NumField(); i++ {
		if tag := st.Type().Field(i).Tag.Get("@Autowired"); tag != "" {
			totalCounter++

			if settable,err := ctx.autowireFieldByX(item, setvalue, i, tag == "*", tag); err {
				return totalCounter, wireableCounter, err
			} else if settable {
				wireableCounter++
			}
		}
    }
    return totalCounter, wireableCounter, false
}

func (ctx *Context) Autowiring(callback func(err bool)) chan bool {
	done := make(chan bool)

	go func () {
		wireSomething := true
		pendingRequest := true

		wiringLoop:
			for pendingRequest && wireSomething {
				wireSomething = false
				pendingRequest = false

				for e := ctx.items.Front(); e != nil; e = e.Next() {
					if item, ok := e.Value.(*populateItem); ok {
						if ! item.autowired {
							if total, wireable,err := ctx.doAutowire(item, false); err {
								pendingRequest = true
								break wiringLoop
							} else if total == 0 {
								wireSomething = true;
								item.autowired = true;
								if summerized, postinit := item.bean.(Summerized); postinit {
									summerized.PostSummerConstruct()
								}
							} else if total == wireable {
								ctx.doAutowire(item, true)
								item.autowired = true
								if summerized, postinit := item.bean.(Summerized); postinit {
									summerized.PostSummerConstruct()
								}
								wireSomething = true
							} else {
								// fmt.Println("Pending: ", total, wireable, item)
								pendingRequest = true
							}
						}
					}
				}
			}
		done <- pendingRequest

		if pendingRequest {
			fmt.Println("Dependency loop?")
		}

		callback(pendingRequest);
	}()

	return done
}

func NewSummer() *Context {
	ctx := new(Context);
	ctx.items = list.New();
	ctx.itemsMap = make(map[string] *populateItem)
	// ctx.done = make(chan bool)
	//ctx := Context{items: list.New(), done: make(chan bool)}
	return ctx;
}