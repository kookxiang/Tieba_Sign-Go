package TiebaSign

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type LikedTieba struct {
	TiebaId     int
	Name        string
	UnicodeName string
	Exp         int
}

func (tieba LikedTieba) String() string {
	return fmt.Sprintf("%s (ID:%d, EXP:%d)", ToUtf8(tieba.Name), tieba.TiebaId, tieba.Exp)
}

func newParseLikedTieba(bajson string) (LikedTieba, error) {
	likedTieba := LikedTieba{}
	exp := regexp.MustCompile("\"name\":\"(.*?)\"")
	names := exp.FindStringSubmatch(bajson)
	if names == nil {
		return likedTieba, errors.New("Cannot get parse string")
	}
	strstr := []string{`{"tmpp":"`, names[1], `"}`}
 bas := []byte(strings.Join(strstr, ""))
	likedTieba.Name = getBaName(bas)
	exp = regexp.MustCompile("\"cur_score\":\"(\\d+)\"")
	likedTieba.Exp, _ = strconv.Atoi(exp.FindStringSubmatch(bajson)[1])
	exp = regexp.MustCompile("{\"id\":\"(\\d+)\"")
	likedTieba.TiebaId, _ = strconv.Atoi(exp.FindStringSubmatch(bajson)[1])
	return likedTieba, nil
}
