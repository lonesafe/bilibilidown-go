package parser

import (
	"strconv"

	"bilibilidown-go/internal/util"
)

func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func av2Bv(avid int64) string {
	return util.Av2Bv(avid)
}
