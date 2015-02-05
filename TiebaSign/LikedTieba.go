package TiebaSign

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type LikedTieba struct {
	tiebaId     int
	name        string
	unicodeName string
	exp         int
}

func (tieba LikedTieba) String() string {
	return fmt.Sprintf("%s (ID:%d, EXP:%d)", ToUtf8(tieba.name), tieba.tiebaId, tieba.exp)
}

func ParseLikedTieba(html string) (LikedTieba, error) {
	likedTieba := LikedTieba{}
	exp := regexp.MustCompile("<a href=\"/f\\?kw=(.*?)\" title=\"(.*?)\"")
	names := exp.FindStringSubmatch(html)
	if names == nil {
		return likedTieba, errors.New("Cannot get parse string")
	}
	likedTieba.unicodeName = names[1]
	likedTieba.name = names[2]
	exp = regexp.MustCompile("<a class=\"cur_exp\".+?>(\\d+)</a>")
	likedTieba.exp, _ = strconv.Atoi(exp.FindStringSubmatch(html)[1])
	exp = regexp.MustCompile("balvid=\"(\\d+)\"")
	likedTieba.tiebaId, _ = strconv.Atoi(exp.FindStringSubmatch(html)[1])
	return likedTieba, nil
}
