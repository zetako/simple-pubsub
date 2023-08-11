package simple_subpub

import (
	"fmt"
	"sync"
	"testing"
)

func Test_BasicLogic(t *testing.T) {
	wg := sync.WaitGroup{}
	publisher := NewPublisher[string]()

	sub1, err := publisher.Subscribe("sub1")
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	sub3, err := publisher.Subscribe("sub3")
	if err != nil {
		t.Fatal(err)
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
