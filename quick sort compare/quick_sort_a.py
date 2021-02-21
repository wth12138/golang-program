import time
import random
import datetime


def quicksort(array):
    left = []
    middle = []
    right = []
    if len(array) <= 1:
        return array
    for i in range(1,len(array)):
        if array[i] > array[0]:
            right.append(array[i])
        elif array[i] == array[0]:
            middle.append(array[i])
        else:
            left.append(array[i])
    left = quicksort(left)
    right = quicksort(right)
    return left + middle +right


def getarr(n, m):
    arr = []
    random.seed(9)
    for i in range(m):
        arr.append(random.randint(1,m))
    return arr

if __name__ == '__main__':
    arr = getarr(99999,100000)
    start=datetime.datetime.now()
    arr_new=quicksort(arr)
    end=datetime.datetime.now()
    #print(arr_new)
    print(end-start)