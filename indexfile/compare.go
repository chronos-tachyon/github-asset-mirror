package indexfile

import (
	"math/big"
	"sort"
)

func CompareBool[T ~bool](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case b == true:
		return LT
	default:
		return GT
	}
}

func CompareByte[T ~byte](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareRune[T ~rune](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareInt[T ~int](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareInt64[T ~int64](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareUint[T ~uint](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareUint64[T ~uint64](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareString[T ~string](a T, b T) CompareResult {
	switch {
	case a == b:
		return EQ
	case a < b:
		return LT
	default:
		return GT
	}
}

func CompareBigInt(a *big.Int, b *big.Int) CompareResult {
	cmp := a.Cmp(b)
	switch {
	case cmp < 0:
		return LT
	case cmp > 0:
		return GT
	default:
		return EQ
	}
}

type ComparableTo[T any] interface {
	CompareTo(T) CompareResult
}

type SortableList[T ComparableTo[T]] []T

func (list SortableList[T]) Len() int {
	return len(list)
}

func (list SortableList[T]) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableList[T]) Less(i, j int) bool {
	return list[i].CompareTo(list[j]) < 0
}

func (list SortableList[T]) Sort() {
	sort.Sort(list)
}
