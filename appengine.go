// +build appengine

package main

import (
	"bytes"
	"net/http"
	"sort"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
)

const playKind = "GroupPlay"

type playEnt struct {
	When    time.Time `datastore:"When,noindex"`
	Players string    `datastore:"Players,noindex"`
}

func addGroupPlayHistory(r *http.Request, gp groupPlay) {
	ctx := appengine.NewContext(r)
	var buf bytes.Buffer
	for i, p := range gp.Players {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(string(p))
	}
	commaPlayers := buf.String()
	key := datastore.NewKey(ctx, playKind, "", 0, nil)
	_, err := datastore.Put(ctx, key, &playEnt{
		When:    gp.When,
		Players: commaPlayers,
	})
	if err != nil {
		ctx.Errorf("Error writing context: %v", err)
		panic(err)
	}
	ctx.Infof("Wrote history item for play at %v with: %v", gp.When, commaPlayers)
}

func loadHistory(r *http.Request) (h history, err error) {
	ctx := appengine.NewContext(r)

	historyMu.Lock() // we never mutate memHistory, but for consistency.
	h = append(h, memHistory...)
	historyMu.Unlock()

	q := datastore.NewQuery(playKind).Limit(5000)
	it := q.Run(ctx)
	var play playEnt
	for {
		_, err := it.Next(&play)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var players []playerID
		for _, v := range strings.Split(play.Players, ",") {
			p := playerID(v)
			if _, ok := playerName[p]; ok {
				players = append(players, p)
			} else {
				ctx.Warningf("Skipping unknown player %q in history", v)
			}
		}
		h = append(h, groupPlay{
			Players: players,
			When:    play.When,
		})
	}
	sort.Sort(byPlayTime(h))
	return h, nil
}
