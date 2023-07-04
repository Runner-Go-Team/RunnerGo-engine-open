package tools

import (
	"fmt"
	"testing"
)

func TestEncodeMd5(t *testing.T) {
	data := MD5("ba87446d859f83ffc1a241cc37589a39app_keyzdfa80ddbc38a9c2c5biz_content{\\\"app_id\\\":\\\"wx6f63e7b5f48043a7\\\",\\\"open_id\\\":\\\"oAElv5d8lF6bEX2qrTyLObfKlUak\\\",\\\"scene\\\":0,\\\"is_test\\\":false}methodwxopen.user.risk.rank.gettimestamp1686401814ba87446d859f83ffc1a241cc37589a39")
	fmt.Println(data)
}
