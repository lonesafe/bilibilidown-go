package util

// BV/AV ID conversion (ported from Java ConvertUtil)
const (
	bvTable = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
	xorVal  = 0x1552356C4CDB
	maskVal = 0x7FFFFFFFFFFFF
)

var trTable map[byte]int

func init() {
	trTable = make(map[byte]int, 58)
	for i := 0; i < 58; i++ {
		trTable[bvTable[i]] = i
	}
}

// Bv2Av converts BV id to av number
func Bv2Av(bvid string) int64 {
	arr := []byte(bvid)
	// swap positions
	arr[3], arr[9] = arr[9], arr[3]
	arr[4], arr[7] = arr[7], arr[4]

	var r int64
	for i := 3; i < len(arr); i++ {
		r = r*58 + int64(trTable[arr[i]])
	}
	return (r & maskVal) ^ xorVal
}

// Av2Bv converts av number to BV id
func Av2Bv(avid int64) string {
	arr := []byte("BV1000000000")
	idx := 11
	tmp := (0x8000000000000 | avid) ^ xorVal
	for tmp != 0 {
		n := int(tmp % 58)
		arr[idx] = bvTable[n]
		tmp /= 58
		idx--
	}
	// swap back
	arr[3], arr[9] = arr[9], arr[3]
	arr[4], arr[7] = arr[7], arr[4]
	return string(arr)
}
