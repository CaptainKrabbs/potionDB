package comparator


type Comparator[T any] interface{
	Compare(a, o *T) int
}