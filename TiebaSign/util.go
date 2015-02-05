package TiebaSign

import (
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"time"
)

func GetTimestampStr() string{
	return fmt.Sprintf("%d", time.Now().Unix());
}

func ToUtf8(gbkString string) string {
	I := bytes.NewReader([]byte(gbkString))
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, _ := ioutil.ReadAll(O)
	return string(d)
}
