package tools

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCallPublicFunc(t *testing.T) {
	InitPublicFunc()
	fmt.Println(CallPublicFunc("RandomFloat0", nil)[0])
	fmt.Println(fmt.Sprintf("%s     ", MD5("fhdssfhibnk146s3")))
}

func TestParsFunc(t *testing.T) {
	//InitPublicFunc()
	//a := "{{__MD5(\"ABc\")__}}"
	//fmt.Println(ParsFunc(a))

	m := new(sync.Map)

	var wg = new(sync.WaitGroup)
	wg.Add(1)
	go func(c *sync.Map, w *sync.WaitGroup) {
		defer w.Done()
		time.Sleep(10 * time.Second)
		c.Range(func(key, value any) bool {
			fmt.Println("key   ", key, "   value:  ", value)
			return true
		})
	}(m, wg)

	m.Store("a", 123)
	wg.Wait()
}

func TestToStandardTime(t *testing.T) {
	fmt.Println("     FJDJSF     ", ToStandardTime(1))
}
