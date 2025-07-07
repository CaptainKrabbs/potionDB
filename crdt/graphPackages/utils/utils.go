package utils

import "fmt"

//Converts an array slice into its string representation.
//
//Empty array returns ""
//Each element's string representation separated by ", "
//returns: assembled string value.
func ToFieldString[T any](arrSlice []T) string {
	//Empty slice
	if len(arrSlice) == 0 {
		return ""
	}
	//Not empty slice
	str := fmt.Sprintf("%v", arrSlice[0])
	for _, elem := range arrSlice[1:] {
		str = fmt.Sprintf("%v, %v", str, elem)
	}
	return str
}

//Filters elements in an slice that fulfil a certain predicate
//
//Returns elements of slice that fulfil the predicate
func FilterSlice[T any](arrSlice []T, pred func(T) bool) []T {
	newSlice := []T{}
	for _, elem := range arrSlice {
		if pred(elem) {
			newSlice = append(newSlice, elem)
		}
	}
	return newSlice
}

//Maps each element in an array slice to a new element produced by a function
//slice of these mapped elements
func MapSlice[T any, V any](arrSlice []T, mapper func(T) V) []V {
	newSlice := []V{}
	for _, v := range arrSlice {
		newSlice = append(newSlice, mapper(v))
	}
	return newSlice
}