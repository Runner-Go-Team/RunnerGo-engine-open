package heartbeat

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"sync"
	"testing"
)

type C struct {
	A A `json:"a"`
}

type A struct {
	P []*B `json:"p"`
}

type B struct {
	K string `json:"k"`
	V string `json:"v"`
}

func T(a C) {

}

func TestGetCpuInfo(t *testing.T) {
	str := "{ \"a\": { \"p\": [{\"k\": \"name\", \"v\": \"123\"}]}}"

	var wg = &sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		var a C
		_ = json.Unmarshal([]byte(str), &a)
		wg.Add(1)
		go func(c C, w *sync.WaitGroup) {
			defer w.Done()
			fmt.Println(&c.A.P)
			fmt.Println(&c.A.P[0].V)
		}(a, wg)
	}
	wg.Wait()

	//fmt.Println("err:      ", err)
	//fmt.Println(a)

	//ctx := context.Background()
	//heartbeat = CheckHeartBeat(ctx)
	//by, _ := json.Marshal(heartbeat)
	//fmt.Println(string(by))
}

//
//func TestGetHostInfo(t *testing.T) {
//	GetHostInfo()
//}
//
//func TestGetMemInfo(t *testing.T) {
//	GetMemInfo()
//}
//
func TestGetDiskInfo(t *testing.T) {
	a, _ := disk.IOCounters()
	for k, _ := range a {
		b, _ := disk.Usage(k)
		c, _ := json.Marshal(b)
		fmt.Println(":    ", string(c))
	}
}

//
func TestGetNetInfo(t *testing.T) {
	//infos, _ := net.IOCounters(true)
	//fmt.Println(infos)

}
