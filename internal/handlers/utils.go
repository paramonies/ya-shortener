package handlers

import (
	"hash/fnv"
	"sync"

	"github.com/paramonies/internal/store"
)

func fanOut(inputCh chan Item, n int) []chan Item {
	chs := make([]chan Item, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan Item)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan Item) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			val, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- val
		}

	}()

	return chs
}

func newWorker(inputCh <-chan Item, rep store.Repository) chan ErrorItem {
	outCh := make(chan ErrorItem)

	go func() {
		for item := range inputCh {
			err := rep.Delete(item.URLID, item.UserID)
			outCh <- ErrorItem{Item: item, Err: err}
		}
		close(outCh)
	}()

	return outCh
}

func fanIn(inputChs ...chan ErrorItem) chan ErrorItem {
	outCh := make(chan ErrorItem)

	go func() {
		var wg sync.WaitGroup

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan ErrorItem) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}

		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
