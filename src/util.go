package main

import "fmt"

// RemoveAtIndexFromUnorderedArray removes the element at
// index i of the array by replacing it with the last element
// in the array, then decreasing the array's size.
func RemoveAtIndexFromUnorderedArray(array []interface{}, i uint) []interface{} {
	l := len(array) - 1
	array[i] = array[l]
	return array[:l]
}

// RemoveFromUnorderedArray removes the element of the
// array by replacing it with the last element in the array,
// then decreasing the array's size.
func RemoveFromUnorderedArray(array []interface{}, elem interface{}) ([]interface{}, error) {
	i, err := GetIndex(array, elem)
	if err != nil {
		return array, err
	}

	array = RemoveAtIndexFromUnorderedArray(array, i)
	return array, nil
}

// GetIndex returns the index of an element in an array.
func GetIndex(array []interface{}, elem interface{}) (uint, error) {
	for i, v := range array {
		if v == elem {
			return uint(i), nil
		}
	}

	return 0, fmt.Errorf("Element %s not in array", elem)
}
