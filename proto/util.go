package proto

import (
	"fmt"
	"time"
)

var programStart int64
var debugging bool

func SetLogStart() {
	programStart = time.Now().UnixNano()
}

func SetDebug(value bool) {
	debugging = value
	SetLogStart()
}

func debug(message string) {
	if debugging {
		fmt.Printf("+%.03fms: %s\n", (float64(time.Now().UnixNano()-programStart) / float64(10E6)), message)
	}
}
