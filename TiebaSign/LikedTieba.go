package TiebaSign

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

func ParseLikedTieba(html string) (LikedTieba, error) {
	likedTieba := LikedTieba{}
	exp := regexp.MustCompile("<a href=\"/f\\?kw=(.*?)\" title=\"(.*?)\"")
	names := exp.FindStringSubmatch(html)
	if names == nil {
		return likedTieba, errors.New("Cannot get parse string")
	}
	likedTieba.UnicodeName = names[1]
	likedTieba.Name = names[2]
	exp = regexp.MustCompile("<a class=\"cur_exp\".+?>(\\d+)</a>")
	likedTieba.Exp, _ = strconv.Atoi(exp.FindStringSubmatch(html)[1])
	exp = regexp.MustCompile("balvid=\"(\\d+)\"")
	likedTieba.TiebaId, _ = strconv.Atoi(exp.FindStringSubmatch(html)[1])
	return likedTieba, nil
}

func ParseLikedTiebaNew(tiebalist map[string]interface{}) ([]LikedTieba, error) {
	likedTieba := LikedTieba{}
	likedTiebaList := make([]LikedTieba, 0)
	list := tiebalist["non-gconforum"].([]interface{})
	for _, v := range list {
		tieba := v.(map[string]interface{})
		likedTieba.Name = tieba["name"].(string)
		likedTieba.UnicodeName = ToUtf8(tieba["name"].(string))
		likedTieba.TiebaId, _ = tieba["id"].(int)
		likedTieba.Exp, _ = tieba["cur_score"].(int)
		likedTiebaList = append(likedTiebaList, likedTieba)
	}
	return likedTiebaList, nil
}
