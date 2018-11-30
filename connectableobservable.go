package rxgo

import (
	"github.com/reactivex/rxgo/handlers"
	"github.com/reactivex/rxgo/options"
	"sync"
)

type ConnectableObservable interface {
	Connect() Observer
	Subscribe(handler handlers.EventHandler, opts ...options.Option) Observer
}

type connectableObservable struct {
	observable Observable
	observers  []Observer
}

func NewConnectableObservable(observable Observable) ConnectableObservable {
	return &connectableObservable{
		observable: observable,
	}
}

func (c *connectableObservable) Subscribe(handler handlers.EventHandler, opts ...options.Option) Observer {
	ob := CheckEventHandler(handler)
	c.observers = append(c.observers, ob)
	return ob
}

func (c *connectableObservable) Connect() Observer {
	source := make([]interface{}, 0)

	for {
		item, err := c.observable.Next()
		if err != nil {
			break
		}
		source = append(source, item)
	}

	var wg sync.WaitGroup

	for _, ob := range c.observers {
		wg.Add(1)
		local := make([]interface{}, len(source))
		copy(local, source)

		var e error
		go func(ob Observer) {
		OuterLoop:
			for _, item := range local {
				switch item := item.(type) {
				case error:
					ob.OnError(item)

					// Record error
					e = item
					break OuterLoop
				default:
					ob.OnNext(item)
				}
			}

			wg.Done()
		}(ob)
	}

	ob := NewObserver()
	go func() {
		wg.Wait()
		ob.OnDone()
	}()

	return ob
}
