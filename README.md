# Simple-PubSub

A simple implement of pub/sub model in golang.

## Example

```go
package main

import (
	"fmt"
	pubsub "github.com/zetako/simple-pubsub"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	publisher := pubsub.NewPublisher[string]()

	sub1, err := publisher.Subscribe("sub1")
	if err != nil {
		panic(err)
	}
	wg.Add(1)
	go func() {
		i := 0
		for i < 2 {
			tmp := <-sub1.Channel
			fmt.Println("sub1: " + tmp)
			i++
		}
		_ = sub1.Unsubscribe()
		wg.Done()
	}()
	sub2, err := publisher.Subscribe("sub2")
	if err != nil {
		panic(err)
	}
	sub3, err := publisher.Subscribe("sub3")
	if err != nil {
		panic(err)
	}

	publisher.Publish("Hello")
	tmp := <-sub3.Channel
	fmt.Println("sub3: " + tmp)
	publisher.Publish("World")
	i := 0
	for i < 2 {
		tmp := <-sub2.Channel
		fmt.Println("sub2: " + tmp)
		i++
	}
	wg.Wait()
}
```

## Known Issue

- [ ] The sequence of message is not strict FIFO