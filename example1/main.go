package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type KVStorage interface {
	Get(context.Context, string) (interface{}, error)
	Put(context.Context, string, interface{}) error
	Delete(context.Context, string) error
}

type IValue interface{}

type ThreadSafeStorage struct {
	items map[string]IValue
	lock  sync.RWMutex
}

func main() {

	var strg KVStorage = &ThreadSafeStorage{}
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	wg.Add(3)
	//Три Goрутины две обращаются к записям с одним ключом и третья запускает отложенную отмену задания
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			err := strg.Put(ctx, "key1", i)
			if err != nil {
				log.Println("Ошибка:", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			err := strg.Put(ctx, "key1", i)
			if err != nil {
				log.Println("Ошибка:", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(100000 * time.Microsecond)
		cancel()
		log.Println("Cancelled")
	}()

	wg.Wait()
	log.Println("main")
}

// Get Имплементация метода получения данных из хранилища
func (tss *ThreadSafeStorage) Get(ctx context.Context, str string) (interface{}, error) {
	tss.lock.RLock()
	defer tss.lock.RUnlock()

	//Проверка отмены задания
	select {
	case <-ctx.Done():
		log.Println("Cancel from context")
		return nil, errors.New("cancelled")
	default:
		if _, check := tss.items[str]; !check {
			return nil, errors.New("key doesn't exist")
		}
		return tss.items[str], nil
	}

}

// Put Имплементация метода внесения данных в хранилище
func (tss *ThreadSafeStorage) Put(ctx context.Context, str string, val interface{}) error {
	tss.lock.Lock()
	defer tss.lock.Unlock()
	if tss.items == nil {
		tss.items = make(map[string]IValue)
	}

	//Проверка отмены задания
	select {
	case <-ctx.Done():
		log.Println("Cancel from context")
		return errors.New("cancelled")
	default:
		tss.items[str] = val
		log.Println("Put in")
		return nil
	}

}

// Delete Имплементация метода удаления записей из хранилища
func (tss *ThreadSafeStorage) Delete(ctx context.Context, str string) error {
	tss.lock.Lock()
	defer tss.lock.Unlock()

	//Проверка отмены задания
	select {
	case <-ctx.Done():
		log.Println("Cancel from context")
		return errors.New("Error: cancelled")
	default:
		_, check := tss.items[str]
		if check {
			delete(tss.items, str)
			log.Println("Deleted")
			return nil
		}
		return errors.New("key doesn't exist")
	}
}
