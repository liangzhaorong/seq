package seq

import "cmp"

// Number is the constraint for all numeric types excluding complex. It is the
// type bound used by numeric aggregate operations such as Sum, Product and Mean.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Ordered is an alias for the standard library [cmp.Ordered] constraint. It is
// the type bound used by ordered aggregate operations such as Min and Max.
type Ordered = cmp.Ordered
