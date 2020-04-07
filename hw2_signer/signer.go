package main

// сюда писать кодpackage main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {

	wg := &sync.WaitGroup{}
	prev := make(chan interface{})
	for _, j := range jobs {
		wg.Add(1)
		next := make(chan interface{})
		go func(in, out chan interface{}, j job) {
			j(in, out)
			wg.Done()
			close(out)
		}(prev, next, j)
		prev = next
	}
	wg.Wait()

}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(dataStr, dataMD5 string) {
			buf1 := make(chan string)
			go func(buf chan string) {
				buf <- DataSignerCrc32(dataStr)
			}(buf1)
			buf2 := make(chan string)
			go func(buf chan string) {
				buf <- DataSignerCrc32(dataMD5)
			}(buf2)
			hashSum := <-buf1 + "~" + <-buf2
			close(buf1)
			close(buf2)
			out <- hashSum
			wg.Done()
		}(strconv.Itoa(data.(int)), DataSignerMd5(strconv.Itoa(data.(int))))
	}
	wg.Wait()
}

type hashRes struct {
	idx  int
	hash string
}

func MultiHash(ch2, ch3 chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range ch2 {
		wg.Add(1)
		go func(data string) {

			buffer := make(chan hashRes)
			for i := 0; i < 6; i++ {
				go func(i int) {
					buffer <- hashRes{i, DataSignerCrc32(strconv.Itoa(i) + data)}
				}(i)
			}
			count := 0
			arr := [6]string{}
			for r := range buffer {
				arr[r.idx] = r.hash
				count++
				if count == 6 {
					close(buffer)
				}
			}
			var sum string
			for _, h := range arr {
				sum += h
			}
			ch3 <- sum
			wg.Done()
		}(data.(string))
	}
	wg.Wait()
}

func CombineResults(ch3, ch4 chan interface{}) {
	var results []string
	for data := range ch3 {
		results = append(results, data.(string))
	}
	sort.Strings(results)
	ch4 <- strings.Join(results, "_")
}