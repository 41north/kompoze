package internal

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func parseURL(rawUrl string) *url.URL {
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatalf("unable to parse url %s: %s", rawUrl, err)
	}
	return u
}

func isTrue(s string) bool {
	b, err := strconv.ParseBool(strings.ToLower(s))
	if err == nil {
		return b
	}
	return false
}

func isFalse(s string) bool {
	return !isTrue(s)
}

func loop(args ...int) (<-chan int, error) {
	var start, stop, step int
	switch len(args) {
	case 1:
		start, stop, step = 0, args[0], 1
	case 2:
		start, stop, step = args[0], args[1], 1
	case 3:
		start, stop, step = args[0], args[1], args[2]
	default:
		return nil, fmt.Errorf("wrong number of arguments, expected 1-3, but got %d", len(args))
	}

	c := make(chan int)
	go func() {
		for i := start; i < stop; i += step {
			c <- i
		}
		close(c)
	}()
	return c, nil
}
