package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type playerID string

// enter new players here in order of playerID, but enter playerName as “Lastname Firstname”
var playerName = map[playerID]string{
	"anne":   "Borelan Anne",
	"annie":  "Lenzer Annie",
	"april":  "McBride April",
	"barb":   "Atkinson Barbara",
	"betsyc": "Conley Betsy",
        "carmen": "Haun Carmen",
	"carol":  "Thomason Carol",
        "carolb": "Beauchamp Carol",
	"chris":  "Braue Chris",
	"char":   "Farley Charlotte",
        "claudia": "Parsons Claudia",
	"dor":    "Hobizal Dorothy",
	"guest1": "Guest 1",
	"guest2": "Guest 2",
	"janl":	  "Luttrell Jan",
	"janp":   "Poujade Jan",
	"jer":    "Johnson Jerrie",
	"joan":   "Wells Joan",
	"joanne": "Maginnis JoAnne",
	"karenm": "McKinney Karen",
	"karens": "Sandberg Karen",
	"karib":  "Beye Kari",
	"kathy":  "Clayton Kathy",
	"kaya":   "Anderson Kay",
	"kayk":   "Klick Kay",
	"laura":  "Welling Laura",
	"lav":    "Howard LaVonne",
	"li":     "Ross Li",
	"liz":    "Christiansen Liz",
	"mavis":  "Varga Mavis",
	"micki":  "Hilliard Micki",
	"nancyc": "Combs Nancy",
	"nancyk": "Killough Nancy",
        "nancys": "Simonsen Nancy",
	"pam":    "Langer Pam",
	"rhonda": "Monroe Rhonda",
	"sandyf": "Fitzpatrick Sandy",
	"sandyp": "Phelps Sandy",
	"sandyz": "Zajdel Sandy",
	"sharron": "Patapoff Sharron",
	"sunny":  "Kwon Sunny",
	"terri":  "Robertson Terri",
	"terry":  "Flaming Terry",
	"tess":   "Holloway Tess",
}

// Historical list of games played before this program existed.
var playHistory = `
# This is a comment. It can go anywhere on the line.
4/8/14
sandyf

`

func init() {
	bs := bufio.NewScanner(strings.NewReader(playHistory))
	var date time.Time
	loc, err := time.LoadLocation("US/Pacific")
	if err != nil {
		panic(err)
	}
	for bs.Scan() {
		line := bs.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "/") {
			t, err := time.ParseInLocation("1/2/06", line, loc)
			if err != nil {
				panic("Invalid date value: " + line)
			}
			date = t.Add(12 * time.Hour) // noonish
			continue
		}
		if date.IsZero() {
			panic("no date before line " + line)
		}
		var group []playerID
		for _, f := range strings.Fields(line) {
			p := playerID(f)
			if _, ok := playerName[p]; !ok {
				panic("Unknown player " + f + " on line: " + line)
			}
			group = append(group, p)
		}
		memHistory = append(memHistory, groupPlay{
			Players: group,
			When:    date,
		})
	}
}

var pairAdjust = map[pair]int{
	pairOf("sandy_fitz", "sandy_z"): 0,
	pairOf("laura", "claire"):       0,
}

func (p playerID) String() string {
	return playerName[p]
}

// players is all player ids, sorted by their display name.
var players []playerID

type byName []playerID

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].String() < s[j].String() }

func init() {
	for p := range playerName {
		players = append(players, p)
	}
	sort.Sort(byName(players))
}

func listPlayers(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `<html>
<head>
<meta content='width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=0' name='viewport' />
<style>
  div.player {
    font-size: 20pt;
    font-family: sans-serif;
    margin-top: 0.5em;
    margin-bottom: 0.5em;
  }
</style>
<script>

function onload() {
   var elements = document.getElementsByClassName("player");
   return; // I give up.
   for (var key in elements) {
       if (key == "length") { break; }
       var el = elements[key];
console.log("key " + key);
console.log(el);
el.childNodes[1].setAttribute("style", "backgroundColor: green");
       el.childNodes[1].addEventListener("touchstart", function(e) {
console.log("touched" + key);
          el.childNodes[1].checked = !el.childNodes[1].checked;
       el.style.backgroundColor = 'yellow';
       }, true);

       el.childNodes[1].addEventListener("mousedown", function(e) {
console.log("clicked" + key);
          el.childNodes[1].checked = !el.childNodes[1].checked;
       el.style.backgroundColor = 'red';
       }, false);
   }
   
}

</script>
</head>
<body onload='onload()'>
<form method=POST action=/makegroups>
`)

	for _, p := range players {
		fmt.Fprintf(w, "<div class=player><input type=checkbox name=playerready value='%s' id='%s_ready'><label for='%s_ready'>%s</label></div>\n", string(p), string(p), string(p), p)
	}
	io.WriteString(w, `<div><input type='submit' value="Let's golf!" style="margin-top: 1em; font-size: 20pt"></form><hr /><a href="/paircounts">Pair counts grid</a> | <a href="/history">History</a></body></html>`)
}

type group []playerID
type outing []group

type byAnneFirst []group

func (s byAnneFirst) Len() int      { return len(s) }
func (s byAnneFirst) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byAnneFirst) Less(i, j int) bool {
	ia := s[i].contains("anne")
	ja := s[j].contains("anne")
	return ia && !ja
}

func (g group) String() string {
	var ps []string
	for _, p := range g {
		ps = append(ps, string(p))
	}
	sort.Strings(ps)
	return strings.Join(ps, ",")
}

func (o outing) String() string {
	var gs []string
	for _, g := range o {
		gs = append(gs, g.String())
	}
	sort.Strings(gs)
	return strings.Join(gs, "/")
}

func (g group) score(hs *historySummary) int {
	var score int
	lastWeek := time.Now().Add(-8 * 24 * time.Hour)
	foreachPair(g, func(p pair) {
		score += pairAdjust[p]

		if hs.count[p] == 0 {
			score += 50
		} else if hs.mostRecent[p].After(lastWeek) {
			score -= 5
		}
	})
	// Anne needs to get to work early. So if she's in a group of
	// 3 (an early group)
	if len(g) == 4 && g.contains("anne") {
		score -= 200
	}
	return score
}

func (g group) contains(who playerID) bool {
	for _, p := range g {
		if p == who {
			return true
		}
	}
	return false
}

func postMakeGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "expected a POST", 400)
		return
	}
	r.ParseForm()

	var readyPlayers []playerID
	for _, v := range r.PostForm["playerready"] {
		readyPlayers = append(readyPlayers, playerID(v))
	}

	hist, err := loadHistory(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	hs := hist.summarize()
	os := makeGroups(hs, readyPlayers)

	io.WriteString(w, `<html>
<head>
<meta content='width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=0' name='viewport' />
<style>
  div.outing {
    font-size: 20pt;
    margin-top: 1em;
    margin-bottom: 1em;
  }

  div.outing input {
      font-size: 30pt;
  }
</style>
</head>
<body>
<form method=POST action=/addhistory>
`)
	for i, o := range os {
		fmt.Fprintf(w, "<div class=outing><b>Score %d:</b>\n", o.score)
		for i, g := range o.o {
			fmt.Fprintf(w, "<div>Group %d: \n", i+1)
			for i, p := range g {
				if i > 0 {
					io.WriteString(w, ", ")
				}
				io.WriteString(w, p.String())
			}
			fmt.Fprintf(w, "</div>")
		}
		b, _ := json.Marshal(o.o)
		fmt.Fprintf(w, "<input type=hidden name='outing%d' value=\"%s\"/>\n", i, html.EscapeString(string(b)))
		fmt.Fprintf(w, "<input type=submit name='submit%d' style='font-size: 18pt;' value=\"Select\"/></div>\n", i)
	}
	io.WriteString(w, `
</form></body>
<body>
`)
}

func postAddHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "expected a POST", 400)
		return
	}
	idx := -1
	for i := 0; i < 50; i++ {
		if r.FormValue(fmt.Sprintf("submit%d", i)) != "" {
			idx = i
			break
		}
	}
	if idx == -1 {
		http.Error(w, "nothing selected", 400)
		return
	}
	var o outing
	if err := json.Unmarshal([]byte(r.FormValue(fmt.Sprintf("outing%d", idx))), &o); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 500)
		return
	}
	for _, g := range o {
		addGroupPlayHistory(r, groupPlay{
			Players: g,
			When:    time.Now(),
		})
	}
	fmt.Fprintf(w, `<html><head>
<meta content='width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=0' name='viewport' />
</head>
<body><h1>Fore! Go golfing.</h1>
`)
}

type groupPlay struct {
	Players []playerID
	When    time.Time
}

type byPlayTime []groupPlay

func (s byPlayTime) Len() int           { return len(s) }
func (s byPlayTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byPlayTime) Less(i, j int) bool { return s[i].When.Before(s[j].When) }

type pair struct {
	a, b playerID // where a < b
}

func pairOf(a, b playerID) pair {
	switch {
	case a == "":
		panic("player a is empty")
	case b == "":
		panic("player b is empty")
	case a == b:
		panic("pair of same player: bogus")
	case a < b:
		return pair{a, b}
	default:
		return pair{b, a}
	}
}

var (
	historyMu  sync.Mutex
	memHistory history
)

type history []groupPlay

func (h history) summarize() *historySummary {
	hs := &historySummary{
		mostRecent: make(map[pair]time.Time),
		count:      make(map[pair]int),
	}
	for _, gp := range h {
		foreachPair(gp.Players, func(p pair) {
			hs.mostRecent[p] = gp.When
			hs.count[p]++
		})
	}
	return hs
}

type historySummary struct {
	mostRecent map[pair]time.Time
	count      map[pair]int
}

type outingAndScore struct {
	o     outing
	score int
}

type byScore []outingAndScore

func (s byScore) Len() int           { return len(s) }
func (s byScore) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byScore) Less(i, j int) bool { return s[j].score < s[i].score }

// makeGroups returns up to 3 possible outings (groups of players).
func makeGroups(hs *historySummary, p []playerID) []outingAndScore {
	const (
		minGolfers  = 6
		numTries    = 1000
		returnCount = 10
	)

	if len(p) < minGolfers {
		return []outingAndScore{
			{[]group{group(p)}, 0},
		}
	}
	sizes := sizesOf(len(p))

	var bestFew []outingAndScore

	seen := map[string]bool{} // outing seen -> true
	for i := 0; i < numTries; i++ {
		o := randOuting(p, sizes)
		if key := o.String(); seen[key] {
			continue
		} else {
			seen[key] = true
		}
		oscore := 0
		for _, g := range o {
			oscore += g.score(hs)
		}
		bestFew = append(bestFew, outingAndScore{o, oscore})
		sort.Sort(byScore(bestFew))
		if len(bestFew) > returnCount {
			bestFew = bestFew[:returnCount]
		}
	}

	// For each possibility being returned, make sure Anne is in
	// group 1:
	for _, os := range bestFew {
		sort.Sort(byAnneFirst(os.o))
	}

	return bestFew
}

// precondition: len(p) == sum of sizes
func randOuting(p []playerID, sizes []int) (o outing) {
	perm := rand.Perm(len(p))
	for _, size := range sizes {
		var g group
		for i := 0; i < size; i++ {
			g = append(g, p[perm[0]])
			perm = perm[1:]
		}
		o = append(o, g)
	}
	return
}

// sizesOf partitions want into 3 and 4 sized player groups.
// want is the sum of the returned parts.
// groups of 3 always come before groups of 4.
//
// For example:
// 6: 3 3
// 7: 3 4
// 8: 4 4
// 9: 3 3 3
// 10: 3 3 4
// 11: 3 4 4
// 12: 4 4 4
// 13: 3 3 3 4
// 14: 3 3 4 4
// 15: 3 4 4 4
// 16: 4 4 4 4
func sizesOf(want int) (parts []int) {
	n := 0 // sum of parts thus far
	for n < want {
		parts = append(parts, 4)
		n += 4
	}
	for i := range parts {
		if n > want {
			parts[i]--
			n--
		}
	}
	return
}

func foreachPair(p []playerID, f func(pair)) {
	for i := 0; i < len(p); i++ {
		for j := i + 1; j < len(p); j++ {
			f(pairOf(p[i], p[j]))
		}
	}
}

func pairCounts(w http.ResponseWriter, r *http.Request) {
	hist, err := loadHistory(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	hs := hist.summarize()
	io.WriteString(w, `<html>
<head>
<style>
</style>
</head>
<body>
<table cellpadding=2 border=1>
<tr>
<td></td>
`)
	for _, p := range players {
		fmt.Fprintf(w, "<td>%s</td>", p)
	}
	io.WriteString(w, `</tr>`)
	for _, p := range players {
		fmt.Fprintf(w, "<tr><td>%s</td>", p)
		for _, p2 := range players {
			show := ""
			if p == p2 {
				show = "-"
			} else {
				n := hs.count[pairOf(p, p2)]
				if n > 0 {
					show = strconv.Itoa(n)
				}
			}
			fmt.Fprintf(w, "<td align=center>%s</td>", show)
		}
		fmt.Fprintf(w, "</tr>\n")
	}

	io.WriteString(w, `</table></body></html>`)

}

func showHistory(w http.ResponseWriter, r *http.Request) {
	hist, err := loadHistory(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	io.WriteString(w, `<html>
<head>
</head>
<body>
`)
	for _, gp := range hist {
		fmt.Fprintf(w, "<p><b>%s</b>: %v</p>", gp.When, gp.Players)
	}
}

func init() {
	http.HandleFunc("/", listPlayers)
	http.HandleFunc("/makegroups", postMakeGroups)
	http.HandleFunc("/addhistory", postAddHistory)
	http.HandleFunc("/paircounts", pairCounts)
	http.HandleFunc("/history", showHistory)
}
