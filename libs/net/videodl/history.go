package videodl

import (
	"github.com/sqp/godock/libs/files/history"

	"encoding/json"
	"errors"
	"io/ioutil"
	"time"
)

// HistoryVideo stores and saves videodl done and queued lists.
//
type HistoryVideo struct {
	history.History
	Queue []*Video // still to do.
	List  []*Video // already done.
}

// NewHistoryVideo creates a videodl history manager with the given file location.
//
func NewHistoryVideo(app history.AppletLike, filename string) *HistoryVideo {
	h := &HistoryVideo{History: *history.New(app, filename)}
	h.SetFuncs(h.load, h.save, h.trim)
	return h
}

// Add adds a video to the download queue.
//
func (h *HistoryVideo) Add(q *Video) error {
	h.Queue = append(h.Queue, q)
	h.trim()
	return h.Save()
}

// Next returns the first video in the download queue.
//
func (h *HistoryVideo) Next() (*Video, bool) {
	if len(h.Queue) == 0 {
		return nil, false
	}
	return h.Queue[0], true
}

// Queued returns the size of the download queue.
//
func (h *HistoryVideo) Queued() int {
	return len(h.Queue)
}

// Done moves the first item in the queue to the done list.
//
func (h *HistoryVideo) Done() error {
	now := time.Now()
	h.Queue[0].DateDone = &now
	h.List = append(h.List, h.Queue[0])
	h.Queue = h.Queue[1:]
	return h.Save()
}

// Remove removes the given video from the given list.
//
func (h *HistoryVideo) Remove(vid *Video, list *[]*Video) error {
	for i, test := range *list {
		if vid == test {
			*list = append((*list)[:i], (*list)[i+1:]...)
			return nil
		}
	}
	return errors.New("HistoryVideo Remove: item not found in list")
}

// Find finds the video patching the given source url.
//
func (h *HistoryVideo) Find(url string) *Video {
	for i, vid := range h.List {
		if vid.URL == url {
			return h.List[i]
		}
	}
	for i, vid := range h.Queue {
		if vid.URL == url {
			return h.Queue[i]
		}
	}
	return nil
}

func (h *HistoryVideo) load() error {
	content, e := ioutil.ReadFile(h.File)
	if e != nil {
		return e
	}

	var relist []*Video
	e = json.Unmarshal(content, &relist)
	for _, inf := range relist {
		if inf.DateDone == nil {
			h.Queue = append(h.Queue, inf)
		} else {
			h.List = append(h.List, inf)
		}
	}

	return e
}

func (h *HistoryVideo) save() error {
	relist := append(h.List, h.Queue...)

	data, e := json.MarshalIndent(relist, "", "\t")
	if e != nil {
		return e
	}
	return ioutil.WriteFile(h.File, data, 0644)
}

func (h *HistoryVideo) trim() {
	switch h.Max {
	case -1:
		return
	case 0:
		h.List = nil
	default:
		if len(h.List) > h.Max {
			h.List = h.List[len(h.List)-h.Max:]
		}
	}
}
