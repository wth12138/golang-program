package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
)

//@brief：耗时统计函数
// func timeCost() func() {
// 	start := time.Now()
// 	return func() {
// 		tc := time.Since(start)
// 		fmt.Printf("time cost = %v\n", tc)
// 	}
// }

func quicksort(array []int, begin, end int) {
	var i, j int
	if begin < end {
		i = begin + 1 // 将array[begin]作为基准数，因此从array[begin+1]开始与基准数比较！
		j = end       // array[end]是数组的最后一位

		for {
			if i >= j {
				break
			}
			if array[i] > array[begin] {
				array[i], array[j] = array[j], array[i]
				j = j - 1
			} else {
				i = i + 1
			}

		}

		/* 跳出while循环后，i = j。
		 * 此时数组被分割成两个部分  -->  array[begin+1] ~ array[i-1] < array[begin]
		 *                           -->  array[i+1] ~ array[end] > array[begin]
		 * 这个时候将数组array分成两个部分，再将array[i]与array[begin]进行比较，决定array[i]的位置。
		 * 最后将array[i]与array[begin]交换，进行两个分割部分的排序！以此类推，直到最后i = j不满足条件就退出！
		 */
		if array[i] >= array[begin] { // 这里必须要取等“>=”，否则数组元素由相同的值时，会出现错误！
			i = i - 1
		}

		array[begin], array[i] = array[i], array[begin]
		//fmt.Printf("%s>%v,%d,%v\n", array[begin:i], array[i], array[j:end])
		quicksort(array, begin, i)
		quicksort(array, j, end)
	}
}

func getArr(n, m int) []int {
	var arr []int
	rand.Seed(9)
	for i := 1; i < n; i++ {
		arr = append(arr, rand.Intn(m)) //生成[0,10000000)的随机数
	}
	return arr
}

func main() {
	file, err := os.Create("./cpu.pprof")
	if err != nil {
		fmt.Println(err)
	}
	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile()
	arr := getArr(9999999, 10000000)
	//fmt.Println(arr)
	//defer timeCost()()
	quicksort(arr, 0, len(arr)-1)
	//fmt.Println(arr)

}
