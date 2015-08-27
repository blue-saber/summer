package sub

import (
	"fmt"
)

type ICat interface {
	Purr()
}

type Cat struct {
}

type Tiger struct {}


func (c *Cat) Purr() {
	fmt.Println("Cat purr");
}

func (t Tiger) Purr() {
	fmt.Println("Tiger purr");
}