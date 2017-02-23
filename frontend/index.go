package frontend

import (
	"net/http"
	"text/template"
	"fmt"
	. "github.com/Evi1/Tieba_Sign-Go/global"
	"bytes"
	"strconv"
)

type menuT struct {
	Name string
	Url  string
}

type indexT struct {
	Menu   string
	Errors int
	Counts int
	Users  int
	Body   string
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//Creat menu
	menu := "";
	isUser := false
	for x := range CookieList {
		m := menuT{Name: x, Url: "?n=" + x}
		buf := new(bytes.Buffer)
		fName := ""
		if len(r.Form["n"]) > 0 && r.Form["n"][0] == x {
			fName = "template/menuSelect.gtpl"
			isUser = true
		} else {
			fName = "template/menu.gtpl"
		}
		t, e := template.ParseFiles(fName)
		if e != nil {
			fmt.Println(e)
			continue
		}
		e = t.Execute(buf, m)
		if e != nil {
			fmt.Println(e)
			continue
		}
		menu += buf.String()
	}
	in := indexT{Menu: menu}
	//Create Panel
	in.Users = len(RunList)
	in.Errors = 0;
	in.Counts = 0
	for _, x1 := range RunList {
		for _, x2 := range x1 {
			in.Counts++
			if x2 == "Failed" {
				in.Errors++
			}
		}
	}
	//Create body
	if isUser {
		in.Body = userBody(r.Form["n"][0])
	} else {
		in.Body = indexBody()
	}

	t, e := template.ParseFiles("template/index.html") //解析模板文件
	if e != nil {
		fmt.Println(e)
	}
	t.Execute(w, in)
}

func indexBody() (b string) {
	str := ""
	for k, v := range ErrorList {
		if v {
			str += makeListI(k, "Error !", "fa-user")
		} else {
			str += makeListI(k, "Fine !", "fa-user")
		}
	}
	b = makeList("UserList", str)
	str = ""
	for k, v := range RunList {
		n := 0;
		m := 0
		for _, q := range v {
			if q != "none" && q != "Failed" {
				n++
			}
			m++
		}
		str += makeProgressI(float64(n)/float64(m)*100, strconv.Itoa(n)+"/"+strconv.Itoa(m), k)
	}
	b += makeProgress("Finished", str)

	return
}

func userBody(user string) (b string) {
	str := "";
	b = ""
	for k, v := range RunList {
		if k == user {
			for tb, st := range v {
				str += makeListI(tb, st, "fa-comment")
			}
		}
	}
	b = makeList("TiebaList", str)
	return
}
