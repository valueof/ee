package main

import "bytes"

type Entry struct {
	IndexStatus    byte
	WorktreeStatus byte
	Path           string
	RenamedFrom    string
}

var editableCodes = map[byte]bool{
	'M': true, 'A': true, 'D': true, 'R': true,
	'C': true, 'T': true, 'U': true, '?': true,
}

func filterEditable(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if editableCodes[e.IndexStatus] || editableCodes[e.WorktreeStatus] {
			out = append(out, e)
		}
	}
	return out
}

func parsePorcelain(data []byte) []Entry {
	if len(data) == 0 {
		return nil
	}
	records := bytes.Split(data, []byte{0})
	if len(records) > 0 && len(records[len(records)-1]) == 0 {
		records = records[:len(records)-1]
	}
	var out []Entry
	for i := 0; i < len(records); {
		rec := records[i]
		if len(rec) < 3 {
			i++
			continue
		}
		idx := rec[0]
		wt := rec[1]
		path := string(rec[3:])
		renameOrCopy := idx == 'R' || idx == 'C' || wt == 'R' || wt == 'C'
		if renameOrCopy && i+1 < len(records) {
			out = append(out, Entry{
				IndexStatus:    idx,
				WorktreeStatus: wt,
				Path:           path,
				RenamedFrom:    string(records[i+1]),
			})
			i += 2
		} else {
			out = append(out, Entry{
				IndexStatus:    idx,
				WorktreeStatus: wt,
				Path:           path,
			})
			i++
		}
	}
	return out
}
