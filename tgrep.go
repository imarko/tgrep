package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/ogier/pflag" // friendlier than the built-in
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type twitresult struct {
	User string `json:"from_user"`
	Text string
	Time string `json:"created_at"`
}
type twitresp struct {
	RefreshUrl string `json:"refresh_url"`
	Results    []twitresult
}

const twSearchBase = "http://search.twitter.com/search.json"

var resnum = flag.IntP("number", "n", 20, "number of items to return")
var reverse = flag.BoolP("reverse", "r", false, "reverse order")
var retweets = flag.BoolP("retweets", "R", false, "include retweeets")
var follow = flag.BoolP("follow", "f", false, "follow mode")
var followDelay = flag.DurationP("followdelay", "F", time.Minute, "refresh delay in follow mode")

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

var tco_url_re = regexp.MustCompile(`http://t\.co/[0-9A-Za-z0-9]+`)

func tco_resolve(url string) string {
	tr := &http.Transport{}
	req, _ := http.NewRequest("HEAD", url, nil)
	res, err := tr.RoundTrip(req)

	if err != nil {
		log.Fatal(err)
	}
	loc,_:=res.Location()
	if loc != nil {
		return loc.String()
	}
	return url
}

func (tw twitresult) String() string {
	t, _ := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", tw.Time)
	tnice := t.Local().Format("Mon 15:04")
	text := fixes.Replace(tw.Text)
	text = tco_url_re.ReplaceAllStringFunc(text, tco_resolve)
	return fmt.Sprintf("[%s] <%s> %s", tnice, tw.User, text)
}

func twitquery(query string) (twitresp, error) {
	var tw twitresp
	resp, err := http.Get(query)
	if err != nil {
		return tw, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tw)
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
	query := fmt.Sprintf("%s?q=%s&rpp=%d", twSearchBase, search, *resnum)
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
		if !*follow { // if not in follow mode we are done
			break
		}
		query = twSearchBase + tw.RefreshUrl
		time.Sleep(*followDelay)
	}
}
