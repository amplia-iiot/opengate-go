package utils

import (
	"fmt"
	"math"
	"strconv"
)

// input: binnary number in a2 complement
// output: value in int64
// AC --> 10101100 (input) --> -84 (output)
func ReverseBits(binaryNumber string) int64 {
	var count int64
	max := len(binaryNumber) - 1
	for index, digit := range binaryNumber {
		number, _ := strconv.ParseInt(string(digit), 16, 64)
		if index == 0 {
			math.Pow(2, float64(max-index))
			count = count - number*int64(math.Pow(2, float64(max-index)))
		} else {
			count = count + number*int64(math.Pow(2, float64(max-index)))
		}
	}
	return count
}

// inputValue :=  "3FB60100" --> resultTarget := "0001B63F"
func ReverseIn2Bytes(inputStr string) string {
	strInBytes := ParseStringToBytes(inputStr)
	arrayInverso := ReverseArray(strInBytes)
	return fmt.Sprintf("%X", arrayInverso)
}

// 3FB60100 --> []byte{0x3F,0xB6,0x01,0x00}
func ParseStringToBytes(strFrame string) (frame []byte) {
	var counter int
	var char string
	for _, x := range strFrame {
		if counter == 0 {
			char = ""
		}
		if counter != 2 {
			char = char + string(x)
		}
		counter = counter + 1
		if counter == 2 {
			i, _ := strconv.ParseInt(char, 16, 64)
			frame = append(frame, byte(i))
			counter = 0
		}
	}
	return
}
