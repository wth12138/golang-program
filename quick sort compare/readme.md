# 不做交换的快排


快速排序的原理想必大部分人早就已经背的滚瓜烂熟了，实现起来虽然有区别但万变不离其宗，无非就是基准值选择上的区别以及交换方式的区别（左中右，左右中）

先前偶然间看到了一种快速排序的思路，声明三个数组，分别放入大于，等于和小于基准值的数，再将他们组合起来，这种写法逻辑上极其简单，同时也完美的贯彻了分治的思路，传统的快排因为无法实现每次都选中中间值，所以交换的时候需要有一些不太合逻辑的处理，理解起来十分费力

直接上代码吧，用go分别写了两版出来

最常见的快排
``` go
func quicksort(array []int, begin, end int) {
	var i, j int
	if begin < end {
		i = begin + 1 
		j = end       

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
		if array[i] >= array[begin] { 
			i = i - 1
		}
		array[begin], array[i] = array[i], array[begin]
		quicksort(array, begin, i)
		quicksort(array, j, end)
	}
}
```

引入新数组的快排
``` go
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
```
接下来写一个计时器，使用计时器直接用defer关键词就好了

``` go
func timeCost() func() {
	start := time.Now()
	return func() {
		tc := time.Since(start)
		fmt.Printf("time cost = %v\n", tc)
	}
}
```

Go官方文档这样描述defer的执行
- A "defer" statement invokes a function whose execution is deferred to the moment the surrounding function returns, either because the surrounding function executed a return statement, reached the end of its function body, or because the corresponding goroutine is panicking.

实际使用时常见的用法有延迟释放文件句柄，或是并发场景下某个函数执行完成后的解锁操作以及异常处理了



随机数组生成器，注意seed要是一个固定值，不然每次生成的数组都不一样，比较起来没啥意义
``` go
func getArr(n, m int) []int {
	var arr []int
	rand.Seed(9)
	for i := 1; i < n; i++ {
		arr = append(arr, rand.Intn(m)) //生成(0,m)的随机数
	}
	return arr
}
```


那么接下来就是分析一下这两种排序思路在性能上到底有什么差异

### 空间
按照最优情况考虑，传统的快排循环一般递归情况下是 n(logn) ，循环情况下为 n(1)，而新的这种快排每一次递归都需要完整的拷贝一次数组，所以复杂度为 n(nlogn)

### 时间

分别在main函数里面运行一下，注意统一defer的位置，先生成随机数再开始计时
``` go
// quick_sort_a
func main() {
	arr := getArr(999999, 1000000)
	defer timeCost()()
	quicksort(arr, 0, len(arr)-1)
}

// quick_sort
func main() {
	arr := getarr(999999, 1000000)
	defer timeCost()()
	quickSort(arr)
}


PS C:\Users\Administrator\go\src\quick sort compare> go run .\quick_sort_a.go
time cost = 75.7974ms
PS C:\Users\Administrator\go\src\quick sort compare> go run .\quick_sort.go  
time cost = 270.2776ms
```
运行时间上没有数量级的差别，但是明显新的方式耗费了更多的时间，这是为什么呢？

“数组是一段地址连续的存储空间”  这句话已经背了很多遍了，对于go 的分片slice来说，可以容纳的空间是已经确定好了的，扩容操作一般需要开辟一个新的内存空间（指数增加），再进行一次复制操作，quick_sort_a的代码中存在大量的数组添加的操作，我认为大量时间的花费就在这里

如何解决呢，马上能想到的方法就是声明左中右分片的时候直接使用当前传入数组的len，但是这样一来，又要浪费两倍的内存空间了，能不能少一点？不可以，极端情况下很有可能发生越界的错误

之后我又用python写了一遍，发现了一个奇怪的情况，两者的差异不大

很懵逼，python的list结构底层是指针数组，难道是有优化？之后又用go重新写了一遍，顺便换了个计时方法，同时这次数组里存的不是值，而是指针，跑一下发现没啥区别。。。。顺便也验证了一下之前的判断，拖后腿的就是append这兄弟!!

不过想来也是，一个指针占用4个字节，我这里最大的数字是1000000，还够不着int64，应该是8个字节，感觉还好

``` go
PS C:\Users\Administrator\go\src\quick sort compare> go run .\quicksortPointer.go
PS C:\Users\Administrator\go\src\quick sort compare> go tool pprof cpu.pprof
Type: cpu
Duration: 1.12s, Total samples = 1.78s (158.79%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) list quick
Total: 1.78s
ROUTINE ======================== main.quickSort in C:\Users\Administrator\go\src\quick sort compare\quicksortPointer.go
     110ms      2.18s (flat, cum) 122.47% of Total
         .          .     26:   var left []*int
         .          .     27:   var right []*int
         .          .     28:   var middle []*int
         .          .     29:   middle = append(middle, arr[0])
         .          .     30:   for i := 1; i < len(arr); i++ {
      30ms       30ms     31:           if *arr[i] < *arr[0] {
      30ms      290ms     32:                   right = append(right, arr[i])
         .          .     33:           } else if *arr[i] > *arr[0] {
      10ms      220ms     34:                   left = append(left, arr[i])
         .          .     35:           } else {
      10ms       40ms     36:                   middle = append(middle, arr[i])
         .          .     37:           }
         .          .     38:   }
         .      720ms     39:   right = quickSort(right)
         .      680ms     40:   left = quickSort(left)
      30ms      200ms     41:   myarr := append(append(left, middle...), right...)
         .          .     42:
         .          .     43:   return myarr
         .          .     44:}

```
``` go
PS C:\Users\Administrator\go\src\quick sort compare> go run .\quicksortPointerA.go
PS C:\Users\Administrator\go\src\quick sort compare> go tool pprof cpu.pprof      
Type: cpu
Time: Mar 4, 2021 at 10:09pm (CST)
Duration: 302.19ms, Total samples = 180ms (59.56%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) list quick
Total: 180ms
ROUTINE ======================== main.quicksort in C:\Users\Administrator\go\src\quick sort compare\quicksortPointerA.go
      80ms      230ms (flat, cum) 127.78% of Total
         .          .     25:
         .          .     26:           for {
         .          .     27:                   if i >= j {
         .          .     28:                           break
         .          .     29:                   }
      30ms       30ms     30:                   if *array[i] > *array[begin] {
      30ms       30ms     31:                           *array[i], *array[j] = *array[j], *array[i]
         .          .     32:                           j = j - 1
         .          .     33:                   } else {
      20ms       20ms     34:                           i = i + 1
         .          .     35:                   }
         .          .     36:
         .          .     37:           }
         .          .     38:
         .          .     39:           /* 跳出while循环后，i = j。
         .          .     40:            * 此时数组被分割成两个部分  -->  array[begin+1] ~ array[i-1] < array[begin]
         .          .     41:            *                           -->  array[i+1] ~ array[end] > array[begin]
         .          .     42:            * 这个时候将数组array分成两个部分，再将array[i]与array[begin]进行比较，决定array[i]的位置。
         .          .     43:            * 最后将array[i]与array[begin]交换，进行两个分割部分的排序！以此类推，直到最后i = j不满足条件就退出！
         .          .     44:            */
         .          .     45:           if *array[i] >= *array[begin] { // 这里必须要取等“>=”，否则数组元素由相同的值时，会出现错误！
         .          .     46:                   i = i - 1
         .          .     47:           }
         .          .     48:
         .          .     49:           *array[begin], *array[i] = *array[i], *array[begin]
         .          .     50:           //fmt.Printf("%s>%v,%d,%v\n", array[begin:i], array[i], array[j:end])
         .       80ms     51:           quicksort(array, begin, i)
         .       70ms     52:           quicksort(array, j, end)
         .          .     53:   }
         .          .     54:}
```

关于指针数组插一句嘴，简单理解就是连续的内存存的不再是值而是指向值的指针了，有什么好处呢，对于一些操作指针变量的函数，使用slice时万一发生了拷贝可就惨了（垃圾无法回收，如果有其它线程也在使用这个值，那你们可能使用的就不是一个东西了），当然缺点也有，cpu加载内存时是按照缓存行来加载的，每一个缓存行加载的都是连续的内存，但由于实际的值并不连续，因此做取值操作就需要多次加载了


再把python的代码也跑一遍吧，用line_profile。这个玩意我在windows上装了俩小时，硬是没撞上，顺便还把vc++ 2012到2019都装了个遍。。。。。直接linux了，五分钟装好

运行一下

```python
[root@localhost quick sort compare]# python3 quick_sort_a.py
Timer unit: 1e-06 s

Total time: 1.8001 s
File: quick_sort_a.py
Function: quicksort at line 6

Line #      Hits         Time  Per Hit   % Time  Line Contents
==============================================================
     6                                           def quicksort(array):
     7     95541      20262.0      0.2      1.1      left = []
     8     95541      21916.0      0.2      1.2      middle = []
     9     95541      19370.0      0.2      1.1      right = []
    10     95541      25302.0      0.3      1.4      if len(array) <= 1:
    11     47771       9483.0      0.2      0.5          return array
    12   1904828     418855.0      0.2     23.3      for i in range(1,len(array)):
    13   1857058     487048.0      0.3     27.1          if array[i] > array[0]:
    14    874067     237007.0      0.3     13.2              right.append(array[i])
    15    982991     251875.0      0.3     14.0          elif array[i] == array[0]:
    16     36719       9969.0      0.3      0.6              middle.append(array[i])
    17                                                   else:
    18    946272     255733.0      0.3     14.2              left.append(array[i])
    19     47770      12866.0      0.3      0.7      left = quicksort(left)
    20     47770      12158.0      0.3      0.7      right = quicksort(right)
    21     47770      18256.0      0.4      1.0      return left + middle +right

Total time: 0 s
File: quick_sort_a.py
```

``` python
[root@localhost quick sort compare]# python3 quick_sort.py 
Timer unit: 1e-06 s

Total time: 2.03766 s
File: quick_sort.py
Function: quicksort at line 5

Line #      Hits         Time  Per Hit   % Time  Line Contents
==============================================================
     5                                           def quicksort(array,begin,end):
     6    240629      50973.0      0.2      2.5  	if begin < end:
     7    120314      25075.0      0.2      1.2  		i = begin + 1 
     8    120314      21273.0      0.2      1.0  		j = end  
     9    120314      21371.0      0.2      1.0  		while 1:
    10   2328761     464968.0      0.2     22.8  			if i >= j:
    11    120314      22246.0      0.2      1.1  				break
    12   2208447     512455.0      0.2     25.1  			if array[i] > array[begin]:
    13   1227713     323598.0      0.3     15.9  				array[i], array[j] = array[j], array[i]
    14   1227713     252726.0      0.2     12.4  				j = j - 1
    15                                           			else:
    16    980734     202389.0      0.2      9.9  				i = i + 1
    17                                           			
    18    120314      29405.0      0.2      1.4  		if array[i] >= array[begin]:
    19     99999      21652.0      0.2      1.1  			i = i - 1
    20                                           
    21    120314      32357.0      0.3      1.6  		array[begin], array[i] = array[i], array[begin]
    22    120314      29370.0      0.2      1.4  		quicksort(array, begin, i)
    23    120314      27804.0      0.2      1.4  		quicksort(array, j, end)


```

emmmmm给人的感觉是python在处理其它逻辑的时候也很慢，相比起来append操作就不算是主要原因了