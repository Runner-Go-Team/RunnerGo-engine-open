package tools

import (
	"fmt"
	"testing"
)

func TestEncodeMd5(t *testing.T) {
	data := SHA224("ABC")
	fmt.Println(data)
}
