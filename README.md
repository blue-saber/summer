# summer

The "Summer" is a tiny dependency injection and autowiring package for GO language.
It was deeply inspired by spring framework and try to implements some tiny portions of its functionality.

### About "Summer"

Being a beginner in writting something in Golang,
I was searching for something similar to spring framework in Java's world ... 
and, finally, I decided to write it myself. So, the "Summer" was born.

Go and Java are quite diffent in many ways,
and I have no idea whether "Inversion of Control" or "Dependency Injection"
is idiomatic Go code should be use or how, so ... mimic other language might be a worse idea.

### How to Use
Add tag "@Autowired" on Fields (as following example) need a dependency injection.
Put "*" as it's value which means it require a Type-based injection,
or put a name, "kitty" for example, which means a Name-baed injection.
```go
type Dog struct {
	Icat *ICat   `@Autowired:"kitty"`
	Rabb *Rabbit `@Autowired:"*"`
}
```
For a Interface pointer or a private field, a proper setter is required to make injection working properly.
And the setter is always the first priority to be chosen to inject dependency.
Simply put "Set" in front of the field's name as its setter.
```go
func (d *Dog) SetIcat(icat interface{}) {
	if origional, ok := icat.(ICat); ok {
		d.Icat = &origional
	}
}
```
Like @PostConstruct in spring framework, this package provide a "Summerized" interface and a 
PostSummerConstruct() function should be implemented if your stuct needed to be called
after dependency inject
```go
func (d *Dog) PostSummerConstruct() {
	fmt.Println("Post Construct")
}
```
The main func look simething like this ...
```go
package main

import (
	"github.com/blue-saber/summer"
)

func main() {
	applicationContext := summer.NewSummer();

	applicationContext.Add(new (sub.Dog), new (sub.Tiger))
	applicationContext.AddWithName("kitty", new (sub.Cat))

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
```
