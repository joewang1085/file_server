func quickSortByKey(arr []bson.M, key string, start, end int) []bson.M {
	if start < end {
		i, j := start, end
		value := arr[(start+end)/2][key].(string)
		for i <= j {
			for arr[i][key].(string) > value {
				i++
			}
			for arr[j][key].(string) < value {
				j--
			}

			if i <= j {
				arr[i], arr[j] = arr[j], arr[i] //666
				i++
				j--
			}
		}

		if start < j {
			quickSortByKey(arr, key, start, j)
		}
		if end > i {
			quickSortByKey(arr, key, i, end)
		}
	}

	return arr
}

func quickSortByKeyAsc(arr []bson.M, key string, start, end int) []bson.M {
	if start < end {
		i, j := start, end
		value := arr[(start+end)/2][key].(string)
		for i <= j {
			for arr[i][key].(string) < value {
				i++
			}
			for arr[j][key].(string) > value {
				j--
			}

			if i <= j {
				arr[i], arr[j] = arr[j], arr[i] //666
				i++
				j--
			}
		}

		if start < j {
			quickSortByKeyAsc(arr, key, start, j)
		}
		if end > i {
			quickSortByKeyAsc(arr, key, i, end)
		}
	}

	return arr
}
