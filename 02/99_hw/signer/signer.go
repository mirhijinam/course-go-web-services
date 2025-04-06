package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

/*
Thoughts:
----

There is a loop with the execution of goroutines. All of them start almost
simultaneously in a chaotic order.

However, we want them to form a pipeline, meaning a well-defined order is
established due to the IO-flow between the jobs.

Therefore, based on the aforementioned reasoning we should use
synchronization primitives. In my opinion, the WaitGroup tool is the most
suitable for this situation.

But a new problem appears: we are not allowed to collect the data before
the next job execution.
An output of the previous job in a pipeline is to be the input of the next
as soon as it's added to the channel between of the jobs.

----
*/

func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup

	in := make(chan interface{}, 100)

	for _, j := range jobs {
		out := make(chan interface{}, 100)

		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			j(in, out)
			close(out)
		}(j, in, out)

		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup

	for x := range in {
		data := strconv.Itoa(x.(int))

		md5 := DataSignerMd5(data)

		wg.Add(1)
		go func(md5 string) {
			defer wg.Done()

			var wgLocal sync.WaitGroup
			var crc32, crc32Md5 string

			wgLocal.Add(1)
			go func() {
				defer wgLocal.Done()
				crc32 = DataSignerCrc32(data)
			}()

			wgLocal.Add(1)
			go func() {
				defer wgLocal.Done()
				crc32Md5 = DataSignerCrc32(md5)
			}()

			wgLocal.Wait()

			out <- crc32 + "~" + crc32Md5
		}(md5)
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup

	for x := range in {
		data := x.(string)

		wg.Add(1)
		go func() {
			defer wg.Done()

			var (
				mu sync.Mutex
				wg sync.WaitGroup
			)

			m := make(map[int]string)

			for th := 0; th < 6; th++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					crc32Th := DataSignerCrc32(strconv.Itoa(th) + data)
					mu.Lock()
					m[th] = crc32Th
					mu.Unlock()
				}()
			}

			wg.Wait()

			output := ""
			for i := range 6 {
				output += m[i]
			}

			out <- output
		}()

	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	output := make([]string, 0)

	for x := range in {
		data := x.(string)
		output = append(output, data)
	}

	sort.Slice(output, func(i, j int) bool { return output[i] < output[j] })

	result := strings.Join(output, "_")

	out <- result
}
