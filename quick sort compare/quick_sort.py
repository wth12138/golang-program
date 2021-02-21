import time
import random
import datetime


def quicksort(array,begin,end):
	if begin < end:
		i = begin + 1 
		j = end  
		while 1:
			if i >= j:
				break
			if array[i] > array[begin]:
				array[i], array[j] = array[j], array[i]
				j = j - 1
			else:
				i = i + 1
			
		if array[i] >= array[begin]:
			i = i - 1

		array[begin], array[i] = array[i], array[begin]
		quicksort(array, begin, i)
		quicksort(array, j, end)


def getarr(n, m):
    arr = []
    random.seed(9)
    for i in range(m):
        arr.append(random.randint(1,m))
    return arr

if __name__ == '__main__':
	arr = getarr(99999,100000)
	start=datetime.datetime.now()
	quicksort(arr,0,len(arr)-1)
	end=datetime.datetime.now()
	print(end-start)
