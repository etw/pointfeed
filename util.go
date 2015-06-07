package main

import (
	"log"
	"regexp"

	point "github.com/etw/pointapi"
)

const (
	FATAL = iota
	ERROR
	WARN
	INFO
	DEBUG
)

func logger(l int, s interface{}) {
	if l <= loglvl {
		switch l {
		case FATAL:
			log.Fatalf("[FATAL] %s\n", s)
		case ERROR:
			log.Printf("[ERROR] %s\n", s)
		case WARN:
			log.Printf("[WARN] %s\n", s)
		case INFO:
			log.Printf("[INFO] %s\n", s)
		case DEBUG:
			log.Printf("[DEBUG] %s\n", s)
		}
	}
}

func findNl(str []rune) int {
	for i, c := range str {
		if c == '\n' {
			return i
		}
	}
	return -1
}

func isElem(a []string, e *string) bool {
	for _, c := range a {
		if c == *e {
			return true
		}
	}
	return false
}

func isKey(m map[string]bool, e *string) bool {
	for k, _ := range m {
		if k == *e {
			return true
		}
	}
	return false
}

func isMatching(a []*regexp.Regexp, e *string) bool {
	for _, c := range a {
		if c.MatchString(*e) {
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

func filterPosts(l []point.PostMeta, f *Filter) []point.PostMeta {
	var res []point.PostMeta
	if f == nil {
		return l
	}
	for _, p := range l {
		if isElem(f.Users, &p.Post.Author.Login) {
			continue
		} else if haveIntersec(f.Tags, p.Post.Tags) {
			continue
		} else {
			res = append(res, p)
		}
	}
	return res
}
