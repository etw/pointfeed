package main

import (
	"github.com/etw/pointapi"
)

func findNl(str []rune) int {
	for i, c := range str {
		if c == '\n' {
			return i
		}
	}
	return -1
}

func isElem(v []string, e string) bool {
	for _, c := range v {
		if c == e {
			return true
		}
	}
	return false
}

func haveIntersec(a []string, b []string) bool {
	var smap = make(map[string]bool)

	for _, val := range a {
		smap[val] = true
	}
	for _, val := range b {
		if smap[val] {
			return true
		}
	}
	return false
}

func filterPosts(l []pointapi.PostMeta, f *Filter) []pointapi.PostMeta {
	var res []pointapi.PostMeta
	if f == nil {
		return l
	}
	for _, p := range l {
		if isElem(f.Users, p.Post.Author.Login) {
			continue
		} else if haveIntersec(f.Tags, p.Post.Tags) {
			continue
		} else {
			res = append(res, p)
		}
	}
	return res
}
