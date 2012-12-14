package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/ogier/pflag"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type twitresult struct {
	User string `json:"from_user"`
	Text string
	Time string `json:"created_at"`
}
type twitresp struct {
	Results []twitresult
}

var resnum = flag.IntP("number", "n", 20, "number of items to return")
var reverse = flag.BoolP("reverse", "r", false, "reverse order")

func twitsearch(search string) (twitresp, error) {
	var tw twitresp
	search = url.QueryEscape(search)
	query := fmt.Sprintf("http://search.twitter.com/search.json?q=%s&rpp=%d", search, *resnum)
	resp, err := http.Get(query)
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&tw)
	return tw, err
}

var fixes = strings.NewReplacer(
	"\n",`\n`,
	"‘", `'`,
	"’", `'`,
	"“", `"`,
	"”", `"`,
	"—", "-",
	"&amp;", "&",
	"&lt;", "<",
	"&gt;", ">",
	"&quot;", `"`,
	"&apos;", "'",
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("need query")
	}
	tw, err := twitsearch(strings.Join(flag.Args(), " "))
	if err != nil {
		log.Fatal(err)
	}

	if !*reverse {
		r := tw.Results
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
	}

	for _, r := range tw.Results {
		t, _ := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", r.Time)
		tnice := t.Local().Format("Mon 15:04")
		
		text:=fixes.Replace(r.Text)
		fmt.Printf("[%s] <%s> %s\n", tnice, r.User, text)
	}
}
