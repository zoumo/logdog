package pythonic

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	list := NewList(1)
	list2 := NewList(2)
	list = list.Append(1, 3, 4, 5, 6)
	list[0] = 2
	list2 = list2.Append(2, 23, 9223372036854775807, "2342", 423.546, "xxx")
	list = list.Extend(list2)
	fmt.Println(list)

}
