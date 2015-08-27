package sub

import "fmt"

type IDog interface {
	Bark()
}

type Dog struct {
	Icat *ICat   `@Autowired:"kitty"`
	Rabb *Rabbit `@Autowired:"*"`
}

func (d *Dog) DoSomething() {
	(*d.Icat).Purr()
	d.Rabb.Jump()
}

func (d *Dog) PostSummerConstruct() {
	fmt.Println("Post Construct")
}

func (d *Dog) SetIcat(icat interface{}) {

	if origional, ok := icat.(ICat); ok {
		d.Icat = &origional
	}
}