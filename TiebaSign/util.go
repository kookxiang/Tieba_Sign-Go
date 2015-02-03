package TiebaSign

import (
	"fmt"
	"time"
)

func GetTimestampStr() string{
	return fmt.Sprintf("%d", time.Now().Unix());
}
