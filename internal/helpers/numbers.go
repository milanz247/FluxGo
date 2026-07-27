package helpers

import "fmt"

const maximumSequenceSize = 10_000

func add(values ...int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func sub(left, right int) int {
	return left - right
}

func mul(values ...int) int {
	total := 1
	for _, value := range values {
		total *= value
	}
	return total
}

func divide(left, right int) (int, error) {
	if right == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return left / right, nil
}

func modulo(left, right int) (int, error) {
	if right == 0 {
		return 0, fmt.Errorf("cannot calculate modulo by zero")
	}
	return left % right, nil
}

func increment(value int) int {
	return value + 1
}

func decrement(value int) int {
	return value - 1
}

func even(value int) bool {
	return value%2 == 0
}

func odd(value int) bool {
	return value%2 != 0
}

func sequence(start, end int) ([]int, error) {
	size := end - start
	if size < 0 {
		size = -size
	}
	if size+1 > maximumSequenceSize {
		return nil, fmt.Errorf("sequence cannot contain more than %d items", maximumSequenceSize)
	}

	step := 1
	if start > end {
		step = -1
	}

	result := make([]int, 0, size+1)
	for current := start; ; current += step {
		result = append(result, current)
		if current == end {
			break
		}
	}
	return result, nil
}
