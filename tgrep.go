package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/ogier/pflag" // friendlier than the built-in
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type twitresult struct {
	Id   int64
	User string `json:"from_user"`
	Text string
	Time string `json:"created_at"`
}
type twitresp struct {
	MaxId      int64  `json:"max_id"`
	RefreshUrl string `json:"refresh_url"`
	Results    []twitresult
}

var resnum = flag.IntP("number", "n", 20, "number of items to return")
var reverse = flag.BoolP("reverse", "r", false, "reverse order")
var retweets = flag.BoolP("retweets", "R", false, "include retweeets")
var follow = flag.BoolP("follow", "f", false, "follow mode")
var followDelay = flag.DurationP("followdelay", "F", time.Minute, "refresh delay in follow mode")

var fixes = strings.NewReplacer(
	"\n", `\n`,
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

func (tw twitresult) String() string {
	t, _ := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", tw.Time)
	tnice := t.Local().Format("Mon 15:04")
	text := fixes.Replace(tw.Text)
	return fmt.Sprintf("[%s] <%s> %s", tnice, tw.User, text)
}

func twitquery(query string) (twitresp, error) {
	var tw twitresp
	resp, err := http.Get(query)
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&tw)
	return tw, err
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("need query")
	}

	args := flag.Args()

	if !*retweets {
		args = append(args, "-rt")
	}
	search := url.QueryEscape(strings.Join(args, " "))
	query := fmt.Sprintf("http://search.twitter.com/search.json?q=%s&rpp=%d", search, *resnum)
	for {
		tw, err := twitquery(query)
		if err != nil {
			log.Fatal(err)
		}
		r := tw.Results
		for i := range r {
			if *reverse {
				fmt.Println(r[i])
			} else {
				fmt.Println(r[len(r)-i-1])
			}
		}
		if !*follow {
			break
		}
		query = fmt.Sprintf("http://search.twitter.com/search.json%s", tw.RefreshUrl)
		time.Sleep(*followDelay)
	}
}
