package main

import (
	"fmt"
	"math/rand"
	"time"
)

//@brief：耗时统计函数
func timeCost() func() {
	start := time.Now()
	return func() {
		tc := time.Since(start)
		fmt.Printf("time cost = %v\n", tc)
	}
}

func quickSort(arr []int) []int {

	if len(arr) <= 1 {
		return arr
	}

	var left []int
	var right []int
	var middle []int
	middle = append(middle, arr[0])
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[0] {
			right = append(right, arr[i])
		} else if arr[i] > arr[0] {
			left = append(left, arr[i])
		} else {
			middle = append(middle, arr[i])
		}
	}
	right = quickSort(right)
	left = quickSort(left)
	myarr := append(append(left, middle...), right...)

	return myarr
}

func getarr(n, m int) []int {
	var arr []int
	rand.Seed(9)
	for i := 1; i < n; i++ {
		arr = append(arr, rand.Intn(m)) //生成[0,10000000)的随机数
	}
	return arr
}

func main() {
	defer timeCost()()
	getarr(999999, 1000000)
	//fmt.Println(arr)

	//fmt.Println(quickSort(arr))
	//quickSort(arr)
}
