package utils

func ArrayCrop[T any](array []T, startIdx, endIdx int) []T {
	n := len(array)

	startIdx = max(0, min(startIdx, n))
	endIdx = max(startIdx, min(endIdx, n))

	res := make([]T, endIdx-startIdx)
	copy(res, array[startIdx:endIdx])
	return res
}
