package main

func findNl(str []rune) int {
	for i, c := range str {
		if c == '\n' {
			return i
		}
	}
	return -1
}
