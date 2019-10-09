package util

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/astaxie/beego/logs"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/axgle/mahonia"
)

func PanicHandler() {
	MyPrintf("检测到系统崩溃=====================")
	if err := recover(); err != nil {
		CrashFileLogs.Critical(fmt.Sprintf("系统崩溃：%v", err))
		CrashFileLogs.Critical("================================================")
	}

	CrashFileLogs.Critical(string(debug.Stack()))
}

func MySleep_ms(ms int64) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func MySleep_s(s int64) {
	time.Sleep(time.Duration(s) * time.Second)
}

//true :不带短横；false:带短横
func GetNow(bflag bool) string {
	now := time.Now()
	if bflag {
		s := now.Format("20060102150405")
		return s
	} else {
		s := now.Format("2006-01-02 15:04:05")
		return s
	}
}

func GetNowAndMs() (string, string) {
	now := time.Now()
	s1 := now.Format("2006-01-02 15:04:05")
	s2 := fmt.Sprintf("%s%03s", now.Format("20060102150405"), strconv.Itoa(now.Nanosecond()/1000000))

	return s1, s2
}

func GetMs() string {
	now := time.Now()
	s := now.Format("20060102150405")
	ms := strconv.Itoa(now.Nanosecond() / 1000000)

	return fmt.Sprintf("%s%03s", s, ms)
}

func GetMs_() string {
	now := time.Now()
	s := now.Format("2006/01/02 15:04:05")
	ms := fmt.Sprintf("%03s", strconv.Itoa(now.Nanosecond()/1000000))

	return s + "." + ms
}

func Diffms(ms1 string, ms2 string) int {
	i1 := ConvertS2I64(ms1)
	i2 := ConvertS2I64(ms2)

	return int(i2 - i1)
}

func ConvertTime_2(strTm string) string {
	tm, _ := time.Parse("2006-01-02 15:04:05", strTm)
	return tm.Format("20060102150405")
}

func ConvertTime2_(strTm string) string {
	tm, _ := time.Parse("20060102150405", strTm)
	return tm.Format("2006-01-02 15:04:05")
}

func ChkTimeOut(strtm string, secs int) bool {
	strnow := time.Now().Format("2006-01-02 15:04:05")
	nowtm, _ := time.Parse("2006-01-02 15:04:05", strnow)

	tm, _ := time.Parse("2006-01-02 15:04:05", strtm)
	if int(nowtm.Sub(tm).Seconds()) > secs {
		return true
	}

	return false
}

func MyPrintf(format string, a ...interface{}) {
	fmt.Printf("%-25s", GetMs_())
	s := fmt.Sprintf(format, a...)
	logs.Info(s)
}

func ConvertI2S(ival int) string {
	str := strconv.Itoa(ival)
	return str
}

func Convert64I2S(ival uint64) string {
	return strconv.FormatUint(ival, 10)
}

func ConvertS2I(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func ConvertS2I64(str string) int64 {
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}

func ConvertByte2Hexstring(b []byte, flag bool) string {
	s := ""
	for _, value := range b {
		s += fmt.Sprintf("%02X", value)
		if flag {
			s += " "
		}
	}

	return s
}

func Converts2b(str string) byte {
	i := ConvertS2I(str)
	return byte(i)
}

func Convertb2s(b byte) string {
	return ConvertI2S(int(b))
}

func Short2Bytes_L(n int) []byte {
	x := int16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

//16进制字符串 转 byte[]
func ConvertHexstring2Byte(s string) []byte {
	barray, _ := hex.DecodeString(s)
	return barray
}

//字符串 转 byte[]
func ConvertString2Byte(s string) []byte {
	return []byte(s)
}

// getDir 获得目录
func getDir(path string) (string, error) {
	rs := []rune(path)
	length := len(rs)
	end := strings.LastIndex(path, "/")

	if length < 0 {
		return "", errors.New("getDir error")
	}

	if end < 0 || end > length {
		return "", errors.New("getDir error")
	}

	return string(rs[0:end]), nil
}

func GetTimeStamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func GetTimeStampMs() int64 {
	return time.Now().UnixNano() / 1000000
}

func GetTimeStampSec() int64 {
	return time.Now().Unix()
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Convertb2sb(value byte) byte {
	return 0x30 + value
}

func ConvertToString8(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func Utf8str2Gbk(str string) string {
	bs1, _ := Utf8ToGbk([]byte(str))
	return string(bs1)
}

func Gbkstr2Utf8(str string) string {
	return ConvertToString8(str, "gbk", "utf-8")

}

func Unicode2Utf8(str string) string {
	s := ""
	for i := 0; i < len(str); {
		r, n := utf8.DecodeRuneInString(str[i:])
		i += n

		s += string(r)
	}

	return s
}

func Remove0(s []byte) []byte {
	slen := len(s)
	d := make([]byte, slen)
	i := 0
	j := 0
	for i = 0; i < slen; i += 1 {
		if s[i] != 0x00 {
			d[j] = s[i]
			j += 1
		}
	}

	return d[0:j]
}

// Int2Bytes Int转Byte
func Int2Bytes_B(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// Bytes2Int Byte转Int
func Bytes2Int_L(b []byte) int {
	byteBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(byteBuffer, binary.LittleEndian, &x)
	return int(x)
}

func Bytes2Int_B(b []byte) int {
	byteBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(byteBuffer, binary.BigEndian, &x)
	return int(x)
}

func Bytes2Short_L(b []byte) int {
	byteBuffer := bytes.NewBuffer(b)

	var x int16
	binary.Read(byteBuffer, binary.LittleEndian, &x)
	return int(x)
}

func Bytes2Short_B(b []byte) int {
	byteBuffer := bytes.NewBuffer(b)

	var x int16
	binary.Read(byteBuffer, binary.BigEndian, &x)
	return int(x)
}

// Short2Bytes Int转Byte
func Short2Bytes_B(n int) []byte {
	x := int16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// Byte2word Byte转word
func Byte2word(b []byte) int16 {
	byteBuffer := bytes.NewBuffer(b)

	var x int16
	binary.Read(byteBuffer, binary.LittleEndian, &x)
	return int16(x)
}

// Word2Byte Word转byte
func Word2Byte(n int16) []byte {
	x := int16(n)

	byteBuffer := bytes.NewBuffer([]byte{})
	binary.Write(byteBuffer, binary.LittleEndian, x)
	return byteBuffer.Bytes()
}

func ConvertByte2UnixTmstring(b []byte, flag bool) string {
	secs := Bytes2Int_B(b)
	return GetTimeByTimestamp((int64)(secs), flag)
}

func StaByte2StaStr(b []byte) string {
	ival := int(b[0])*10000 + (int((b[1]&0xE0)>>5))*100 + int((b[1] & 0x1F))
	return ConvertI2S(ival)
}

func ConvertTimestmp2Time(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02 15:04:05")
}

func GetTimeByTimestamp(sec int64, flag bool) string {
	if flag {
		return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
	} else {
		return time.Unix(sec, 0).Format("20060102150405")
	}
}

//字符串时间转unix时间
func ConvertTime2Unix(strTm string, flag bool) []byte {
	if flag {
		tm1, _ := time.Parse("2006-01-02 15:04:05", strTm)
		return Int2Bytes_B((int)(tm1.Unix()) - 3600*8)
	} else {
		tm2, _ := time.Parse("20060102150405", strTm)
		return Int2Bytes_B((int)(tm2.Unix()) - 3600*8)
	}
}

func GetDigitsByStr(str string) int {
	index := strings.IndexAny(str, "[0123456789]")

	if index < 0 {
		return 0
	}

	return ConvertS2I(str[index:])
}

func EncodeBase64(enbyte []byte) string {
	encodeString := base64.StdEncoding.EncodeToString(enbyte)
	return encodeString
}

func DncodeBase64(dncodeString string) ([]byte, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(dncodeString)
	return decodeBytes, err
}
