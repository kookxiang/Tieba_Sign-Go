package TiebaSign

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
)

func ToUtf8(gbkString string) string {
	I := bytes.NewReader([]byte(gbkString))
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, _ := ioutil.ReadAll(O)
	return string(d)
}
