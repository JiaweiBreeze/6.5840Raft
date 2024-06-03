package raft

import (
	"fmt"
	"strings"
)

type Entry struct {
	Command interface{}
	Term    int
	Index   int
}

type Log struct {
	Entries []Entry
	Index   int
}

func (l *Log) append(entries ...Entry) {
	l.Entries = append(l.Entries, entries...)
}

func makeLog() Log {
	return Log{
		Entries: make([]Entry, 0),
		Index:   0,
	}
}

func (l *Log) at(idx int) *Entry {
	return &l.Entries[idx]
}

func (l *Log) len() int {
	return len(l.Entries)
}

func (l *Log) lastLog() *Entry {
	return l.at(l.len() - 1)
}

func (l *Log) String() string {
	nums := []string{}
	for _, entry := range l.Entries {
		nums = append(nums, fmt.Sprintf("%4d", entry.Term))
	}
	return fmt.Sprint(strings.Join(nums, "|"))
}
func (l *Log) slice(idx int) []Entry {
	return l.Entries[idx:]
}

func (l *Log) truncate(idx int) {
	l.Entries = l.Entries[:idx]
}
