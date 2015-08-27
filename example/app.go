package main

import (
	"fmt"
	"github.com/blue-saber/summer"
	"./sub"
)

func main() {
	applicationContext := summer.NewSummer();

	applicationContext.Add(new (sub.Dog), new (sub.Tiger))
	applicationContext.AddWithName("kitty", new (sub.Cat))
	// applicationContext.Add(new (sub.Tiger))
	applicationContext.AddWithName("rabbit", new (sub.Rabbit))

	done := applicationContext.Autowiring(func (err bool) {
		if err {
			fmt.Println("Failed to autowiring.")
		} else {
			fmt.Println("Autowired.")

			if result := applicationContext.GetByName("rabbit"); result != nil {
				rabbit := result.(*sub.Rabbit)
				rabbit.Jump()
			}
		}
	});

	err := <-done

	if ! err {
		var icat sub.ICat

		if result := applicationContext.Get(&icat); result != nil {
			icat = result.(sub.ICat)
			icat.Purr()
		}

		var dog *sub.Dog

		if result := applicationContext.Get(dog); result != nil {
			dog = result.(*sub.Dog)
			dog.DoSomething()
		}
	}
}
