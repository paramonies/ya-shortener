package handlers

import (
	"hash/fnv"
	"sync"

	"github.com/paramonies/internal/store"
)

func fanOut(inputCh chan item, n int) []chan item {
	chs := make([]chan item, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan item)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan item) {
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

func newWorker(inputCh <-chan item, rep store.Repository) chan errorItem {
	outCh := make(chan errorItem)

	go func() {
		for item := range inputCh {
			err := rep.Delete(item.URLID, item.UserID)
			outCh <- errorItem{item: item, Err: err}
		}
		close(outCh)
	}()

	return outCh
}

func fanIn(inputChs ...chan errorItem) chan errorItem {
	outCh := make(chan errorItem)

	go func() {
		var wg sync.WaitGroup

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan errorItem) {
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
