package utils

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"unsafe"
)

func HexInStringToInt(hexNum string) int64 {
	numberInt, _ := strconv.ParseInt(hexNum, 16, 64)
	return numberInt
}

// byte array to uint64
func ByteArrayToInt64(byteArr []byte) uint64 {
	return getInt64FromHex(getBytesInStr(byteArr))
}
func getInt64FromHex(hexStr string) uint64 {
	numberInt, _ := strconv.ParseUint(hexStr, 16, 64)
	return numberInt
}
func getBytesInStr(byteArr []byte) string {
	return fmt.Sprintf("%X", byteArr)
}

// get reverse version of array
func ReverseArray(arr []byte) []byte {
	var newArray []byte = make([]byte, len(arr))
	copy(newArray, arr)
	for i, j := 0, len(newArray)-1; i < j; i, j = i+1, j-1 {
		newArray[i], newArray[j] = newArray[j], newArray[i]
	}
	return newArray
}

// parse uint64 number to byte array
func IntToByteArray(num uint64, fixed int) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}
	arr = ReverseArray(arr)
	if fixed == 0 {
		for j := 0; j < len(arr); j++ {
			if arr[j] != 0x00 {
				return arr[j:]
			}
		}
	} else {
		if len(arr) >= fixed {
			return arr[len(arr)-fixed:]
		}
	}
	return []byte{0x00}
}
func ToFloat32IEEE754(bytesNum []byte) (float32, error) {
	reverseArr := ReverseArray(bytesNum)
	base64EncodedStr := base64.StdEncoding.EncodeToString(reverseArr)
	decodedToBytes, err := base64.StdEncoding.DecodeString(base64EncodedStr)
	if err != nil {
		return 0, err
	}
	return Float32frombytes(decodedToBytes)
}
func ToFloat64IEEE754(bytesNum []byte) (float64, error) {
	reverseArr := ReverseArray(bytesNum)
	base64EncodedStr := base64.StdEncoding.EncodeToString(reverseArr)
	decodedToBytes, err := base64.StdEncoding.DecodeString(base64EncodedStr)
	if err != nil {
		return 0, err
	}
	return Float64frombytes(decodedToBytes)
}
func Float32frombytes(bytes []byte) (num float32, err error) {
	defer func() {
		reco := recover()
		if reco != nil {
			err = errors.New("panic recover")
			return
		}
	}()
	bits := binary.LittleEndian.Uint32(bytes)
	num = math.Float32frombits(bits)
	if math.IsNaN(float64(num)) {
		err = fmt.Errorf("NaN")
	}
	return
}
func Float64frombytes(bytes []byte) (num float64, err error) {
	defer func() {
		reco := recover()
		if reco != nil {
			err = errors.New("panic recover")
			return
		}
	}()
	bits := binary.LittleEndian.Uint64(bytes)
	num = math.Float64frombits(bits)
	if math.IsNaN(num) {
		err = fmt.Errorf("NaN")
	}
	return
}
func PadRightByteArray(inputArry []byte, maxSize int) []byte {
	rsp := make([]byte, maxSize)
	if len(inputArry) >= maxSize {
		return inputArry
	}
	copy(rsp, inputArry)
	return rsp
}
