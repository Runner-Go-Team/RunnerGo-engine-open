package config

import (
	"fmt"
	"os"
	"testing"
)

func TestEnvInitConfig(t *testing.T) {
	name := os.Getenv("Guo_Name")
	fmt.Println(name)
}
