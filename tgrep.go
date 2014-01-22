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

type twuser struct {
	Name string `json:"screen_name"`
}
type twstatus struct {
	User twuser
	Text string
	Time string `json:"created_at"`
}

type twmeta struct {
	RefreshURL string `json:"refresh_url"`
}
type twitresp struct {
	Statuses []twstatus
	Meta     twmeta `json:"search_metadata"`
}

const twSearchBase = "https://api.twitter.com/1.1/search/tweets.json"

var resnum = flag.IntP("number", "n", 20, "number of items to return")
var reverse = flag.BoolP("reverse", "r", false, "reverse order")
var retweets = flag.BoolP("retweets", "R", false, "include retweeets")
var follow = flag.BoolP("follow", "f", false, "follow mode")
var followDelay = flag.DurationP("followdelay", "F", time.Minute, "refresh delay in follow mode")
var lang = flag.StringP("lang", "l", "", "limit to language")

var fixes = strings.NewReplacer(
	"\n", ` \n `, // spaces to make c&p easier
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

func (tw twstatus) String() string {
	t, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tw.Time)
	tnice := t.Local().Format("Mon 15:04")
	text := fixes.Replace(tw.Text)
	return fmt.Sprintf("[%s] <%s> %s", tnice, tw.User.Name, text)
}

func twitquery(query string) (twitresp, error) {
	var tw twitresp
	auth:="Bearer "+twtoken
	client:=&http.Client{}
	req,_ := http.NewRequest("GET", query, nil)
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		return tw, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tw)
	return tw, err
}

func oauth() {

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
	query := fmt.Sprintf("%s?q=%s&count=%d&lang=%s", twSearchBase, search, *resnum, *lang)
	for {
		tw, err := twitquery(query)
		if err != nil {
			log.Fatal(err)
		}
		r := tw.Statuses
		for i := range r {
			if *reverse {
				fmt.Println(r[i])
			} else {
				fmt.Println(r[len(r)-i-1])
			}
		}
		if !*follow { // if not in follow mode we are done
			break
		}
		query = twSearchBase + tw.Meta.RefreshURL
		time.Sleep(*followDelay)
	}
}
