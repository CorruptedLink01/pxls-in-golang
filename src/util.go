package main

import (
	"crypto/rand"

	"github.com/jinzhu/gorm"
)

// NotFoundError is a generic "not found" error.
type NotFoundError struct {
	msg string
}

// Error returns the error message.
func (e *NotFoundError) Error() string {
	return e.msg
}

// IsNotFoundError returns whenever an error is of the generic NotFoundError type
// or a gorm.RecordNotFoundError.
func IsNotFoundError(err error) (b bool) {
	_, b = err.(*NotFoundError)
	b = b || gorm.IsRecordNotFoundError(err)
	return
}

var (
	tokenDictionary    = []rune("AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz0123456789")
	tokenDictionaryLen = len(tokenDictionary)
)

// GenerateToken creates a token with the given length
func GenerateToken(length uint) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	r := make([]rune, length)
	for i, v := range b {
		r[i] = tokenDictionary[int(v)%tokenDictionaryLen]
	}

	return string(r), nil
}

// TODO(netux): these are unused for now, uncomment when needed
// RemoveAtIndexFromUnorderedArray removes the element at
// index i of the array by replacing it with the last element
// in the array, then decreasing the array's size.
// func RemoveAtIndexFromUnorderedArray(array []interface{}, i uint) []interface{} {
// 	l := len(array) - 1
// 	array[i] = array[l]
// 	return array[:l]
// }

// // RemoveFromUnorderedArray removes the element of the
// // array by replacing it with the last element in the array,
// // then decreasing the array's size.
// func RemoveFromUnorderedArray(array []interface{}, elem interface{}) ([]interface{}, error) {
// 	i, err := GetIndex(array, elem)
// 	if err != nil {
// 		return array, err
// 	}

// 	array = RemoveAtIndexFromUnorderedArray(array, i)
// 	return array, nil
// }

// // GetIndex returns the index of an element in an array.
// func GetIndex(array []interface{}, elem interface{}) (uint, error) {
// 	for i, v := range array {
// 		if v == elem {
// 			return uint(i), nil
// 		}
// 	}

// 	return 0, fmt.Errorf("Element %s not in array", elem)
// }
