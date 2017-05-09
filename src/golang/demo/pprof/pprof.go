package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	cpuf, err := os.Create("cpu_profile")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(cpuf)
	defer pprof.StopCPUProfile()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	test(ctx)

	time.Sleep(time.Second * 3)
	memf, err := os.Create("mem_profile")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	if err := pprof.WriteHeapProfile(memf); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	memf.Close()
}

func test(c context.Context) {
	i := 0
	j := 0
	go func() {
		m := map[int]int{}
		for {
			i++
			m[i] = i
		}
	}()
	go func() {
		m := map[int]int{}
		for {
			j++
			m[i] = i
		}
	}()
	select {
	case <-c.Done():
		fmt.Println("done, i", i, "j", j)
		return
	}
}
