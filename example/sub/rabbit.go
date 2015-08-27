package sub

import "fmt"

type Rabbit struct {
	// Husky *Dog `@Autowired:"*"`
}

func (r *Rabbit) Jump() {
	fmt.Println("Rabbit jump");
}