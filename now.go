package main

import "time"

type Nower func() int

func Now() int {
	return int(time.Now().Unix())
}
