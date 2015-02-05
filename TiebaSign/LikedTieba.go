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
