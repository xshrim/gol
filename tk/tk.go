package tk

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	rd "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type V = struct{}
type I = interface{}
type M = map[string]interface{}

const hexs = "0123456789abcdef"

// var noEscapeTable = [256]bool{}

// func init() {
// 	for i := 0; i <= 0x7e; i++ {
// 		noEscapeTable[i] = i >= 0x20 && i != '\\' && i != '"'
// 	}
// }

// util functions

// // Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding
// func itoa(i int, wid int) []byte {
// 	// Assemble decimal in reverse order
// 	var b [20]byte
// 	bp := len(b) - 1
// 	for i >= 10 || wid > 1 {
// 		wid--
// 		q := i / 10
// 		b[bp] = byte('0' + i - q*10)
// 		bp--
// 		i = q
// 	}
// 	// i < 10
// 	b[bp] = byte('0' + i)
// 	return b[bp:]
// }

func int2chinese(num int) string {
	//1、数字为0
	if num == 0 {
		return "零"
	}
	var ans string
	//数字
	szdw := []string{"十", "百", "千", "万", "十万", "百万", "千万", "亿"}
	//数字单位
	sz := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	res := make([]string, 0)

	//数字单位角标
	idx := -1
	for num > 0 {
		//当前位数的值
		x := num % 10
		//2、数字大于等于10
		// 插入数字单位，只有当数字单位角标在范围内，且当前数字不为0 时才有效
		if idx >= 0 && idx < len(szdw) && x != 0 {
			res = append([]string{szdw[idx]}, res...)
		}
		//3、数字中间有多个0
		// 当前数字为0，且后一位也为0 时，为避免重复删除一个零文字
		if x == 0 && len(res) != 0 && res[0] == "零" {
			res = res[1:]
		}
		// 插入数字文字
		res = append([]string{sz[x]}, res...)
		num /= 10
		idx++
	}
	//4、个位数为0
	if len(res) > 1 && res[len(res)-1] == "零" {
		res = res[:len(res)-1]
	}
	//合并字符串
	for i := 0; i < len(res); i++ {
		ans = ans + res[i]
	}
	return ans
}

func int2roman(num int) string {
	//创建映射列表
	numsmap := map[int]string{
		1:    "I",
		4:    "IV",
		5:    "V",
		9:    "IX",
		10:   "X",
		40:   "XL",
		50:   "L",
		90:   "XC",
		100:  "C",
		400:  "CD",
		500:  "D",
		900:  "CM",
		1000: "M",
	}
	//创建整数数组
	numsint := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	results := []string{}
	count := 0
	//进入循环
	for i := 0; i < len(numsint) && num != 0; i++ {
		//判断当前数字是否比map中数值大
		count = num / numsint[i]
		//如果大，则减去当前值
		num = num - count*numsint[i]
		//并记录字符，注意这里用的是for循环
		for count != 0 {
			results = append(results, numsmap[numsint[i]])
			//更新count值
			count--
		}
	}
	return strings.Join(results, "")
}

// intToCircledNumber 将整数转换为带圈的 Unicode 字符。
// 支持 1 到 20 的整数，超出范围返回错误。
func int2cnum(num int) string {
	if num < 1 || num > 20 {
		return ""
	}

	// Unicode 编码从 ① (U+2460) 到 ⑳ (U+2473) 是连续的
	circledNum := rune(0x245F + num)
	return string(circledNum)
}

// intToParenthesizedNumber 将整数转换为带括号的 Unicode 字符
// 支持 1 到 20 的整数，超出范围返回错误。
func int2pnum(num int) string {
	if num < 1 || num > 20 {
		return ""
	}

	// Unicode 编码从 ⑴ (U+2474) 到 ⑳ (U+2487) 是连续的
	parenthesizedNum := rune(0x2473 + num)
	return string(parenthesizedNum)
}

func toUpper(str string) string {
	var dst []rune
	for _, v := range str {
		if v >= 97 && v <= 122 {
			v -= 32
		}
		dst = append(dst, v)
	}
	return string(dst)
}

func toLower(str string) string {
	var dst []rune
	for _, v := range str {
		if v >= 65 && v <= 90 {
			v += 32
		}
		dst = append(dst, v)
	}
	return string(dst)
}

func toCapitalize(str string) string {
	var dst []rune
	for idx, v := range str {
		if idx == 0 {
			if v >= 97 && v <= 122 {
				v -= 32
			}
		} else {
			if v >= 65 && v <= 90 {
				v += 32
			}
		}
		dst = append(dst, v)
	}
	return string(dst)
}

func replaceEscapePeriod(str string, flag bool) string {
	var buf []rune
	for _, c := range str {
		if flag {
			if c == '.' {
				l := len(buf)
				if l > 0 && buf[l-1] == '\\' {
					buf = buf[:l-1]
					buf = append(buf, '`')
					continue
				}
			}
		} else {
			if c == '`' {
				buf = append(buf, '.')
				continue
			}
		}
		buf = append(buf, c)
	}
	return string(buf)
}

func stringEscapeSep(str string, sep rune) string {
	var buf []rune
	for _, c := range str {
		if c == sep {
			buf = append(buf, '\\')
		}
		buf = append(buf, c)
	}
	return string(buf)
}

func stringRepeat(str string, times int) string {
	out := ""
	for i := 0; i < times; i++ {
		out += str
	}

	return out
}

func stringJoin(strs []string, r rune) string {
	out := ""
	for idx, str := range strs {
		out += str
		if idx != len(strs)-1 {
			out += string(r)
		}
	}

	return out
}

func stringContainRune(str string, r rune) bool {
	for _, c := range str {
		if c == r {
			return true
		}
	}
	return false
}

func stringIndex(str, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	if len(str) < len(sub) {
		return -1
	}
	for i := 0; i <= len(str)-len(sub); i++ {
		if string(str[i:i+len(sub)]) == sub {
			return i
		}
	}
	return -1
}

func stringContainStr(str, sub string) bool {
	return stringIndex(str, sub) >= 0
	// sr := []rune(s)
	// tr := []rune(t)
	// for i := 0; i <= len(sr)-len(tr); i++ {
	// 	if sr[i] == tr[0] {
	// 		j := 1
	// 		for ; j < len(tr); j++ {
	// 			if sr[i+j] != tr[j] {
	// 				break
	// 			}
	// 		}
	// 		if j == len(tr) {
	// 			return true
	// 		}
	// 	}
	// }
	// return false
}

func stringPrefixStr(str, sub string) bool {
	return stringIndex(str, sub) == 0
}

func stringSuffixStr(str, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(str) < len(sub) {
		return false
	}
	return string(str[len(str)-len(sub):]) == sub
}

func stringTrimPrefix(s, prefix string) string {
	if stringPrefixStr(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

func stringSplit(str string, r rune) []string {
	var strs []string
	var runes []rune
	for i, c := range str {
		if c != r {
			runes = append(runes, c)
			if i == len(str)-1 {
				strs = append(strs, string(runes))
				break
			}
		} else {
			if runes != nil {
				strs = append(strs, string(runes))
			}
			runes = nil
		}
	}
	return strs
}

func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", fmt.Errorf("time: bad [0-9]*")
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", fmt.Errorf("time: bad [0-9]*")
		}
	}
	return x, s[i:], nil
}

func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}

func mapi2maps(i interface{}) interface{} {
	// var body interface{}
	// _ = yaml.Unmarshal([]byte(yamlstr), &body)
	// body = mapi2maps(body)
	// jsb,  _:= json.Marshal(body);
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[fmt.Sprintf("%v", k)] = mapi2maps(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = mapi2maps(v)
		}
	}
	return i
}

// func map2str(fds F) string {
// 	if fds == nil {
// 		return ""
// 	}

// 	res := ""
// 	for k, v := range fds {
// 		val := fmt.Sprintf("%v", v)
// 		if bytes.ContainsRune([]byte(val), ' ') {
// 			val = "'" + val + "'"
// 		}
// 		res += k + "=" + val + " "
// 	}
// 	return res[:len(res)-1]
// }

// convert data to json-like []byte
func tojson(dst []byte, v interface{}) []byte {
	switch val := v.(type) {
	case []byte:
		dst = appendBytes(dst, val)
	case error:
		dst = appendStr(dst, val.Error())
	case bool:
		dst = appendBool(dst, val)
	case []bool:
		dst = appendBools(dst, val)
	case int:
		dst = appendInt(dst, val)
	case []int:
		dst = appendInts(dst, val)
	case int8:
		dst = appendInt8(dst, val)
	case []int8:
		dst = appendInts8(dst, val)
	case int16:
		dst = appendInt16(dst, val)
	case []int16:
		dst = appendInts16(dst, val)
	case int32:
		dst = appendInt32(dst, val)
	case []int32:
		dst = appendInts32(dst, val)
	case int64:
		dst = appendInt64(dst, val)
	case []int64:
		dst = appendInts64(dst, val)
	case uint:
		dst = appendUint(dst, val)
	case []uint:
		dst = appendUints(dst, val)
	case uint8:
		dst = appendUint8(dst, val)
	case uint16:
		dst = appendUint16(dst, val)
	case []uint16:
		dst = appendUints16(dst, val)
	case uint32:
		dst = appendUint32(dst, val)
	case []uint32:
		dst = appendUints32(dst, val)
	case uint64:
		dst = appendUint64(dst, val)
	case []uint64:
		dst = appendUints64(dst, val)
	case float32:
		dst = appendFloat32(dst, val)
	case []float32:
		dst = appendFloats32(dst, val)
	case float64:
		dst = appendFloat64(dst, val)
	case []float64:
		dst = appendFloats64(dst, val)
	case string:
		dst = appendStr(dst, val)
	case []string:
		dst = appendStrs(dst, val)
	case time.Time:
		dst = appendTime(dst, val, time.RFC3339)
	case []time.Time:
		dst = appendTimes(dst, val, time.RFC3339)
	case time.Duration:
		dst = appendDuration(dst, val, time.Millisecond)
	case []time.Duration:
		dst = appendDurations(dst, val, time.Millisecond)
	case map[string]interface{}:
		dst = append(dst, '{')
		for k, v := range val {
			dst = appendKey(dst, k)
			dst = tojson(dst, v)
		}
		dst = append(dst, '}')
	case map[interface{}]interface{}:
		// m2 := map[string]interface{}{}
		// for k, v := range val {
		// 	m2[fmt.Sprintf("%v", k)] = mapi2maps(v) // convert key to string
		// }
		return tojson(dst, mapi2maps(v))
	case []map[string]interface{}:
		dst = append(dst, '[')
		for _, m := range val {
			if dst[len(dst)-1] == '[' || dst[len(dst)-1] == ',' {
				dst = append(dst, '{')
			}
			for k, v := range m {
				dst = appendKey(dst, k)
				dst = tojson(dst, v)
			}
			dst = append(dst, '}')
			dst = append(dst, ',')
		}
		if len(val) > 0 {
			dst = dst[:len(dst)-1]
		}
		dst = append(dst, ']')
	case []interface{}:
		dst = append(dst, '[')
		for _, s := range val {
			dst = tojson(dst, s)
			dst = append(dst, ',')
		}
		if len(val) > 0 {
			dst = dst[:len(dst)-1]
		}
		dst = append(dst, ']')
	case net.IP:
		dst = appendIP(dst, val)
	case []net.IP:
		dst = appendIPs(dst, val)
	case net.IPNet:
		dst = appendIPNet(dst, val)
	case []net.IPNet:
		dst = appendIPNets(dst, val)
	case net.HardwareAddr:
		dst = appendMac(dst, val)
	case []net.HardwareAddr:
		dst = appendMacs(dst, val)
	case nil:
		dst = append(dst, []byte("null")...)
	case interface{}:
		dst = appendInterface(dst, val)
	default:
		dst = appendObject(dst, val)
	}
	return dst
}

// marshal json with skipping func fields
func Marshal(v interface{}) ([]byte, error) {
	value := reflect.Indirect(reflect.ValueOf(v))
	typ := value.Type()
	if typ.Kind() == reflect.Struct {
		sf := make([]reflect.StructField, 0)
		for i := 0; i < typ.NumField(); i++ {
			sf = append(sf, typ.Field(i))
			if typ.Field(i).Type.Kind() == reflect.Func {
				sf[i].Tag = `json:"-"`
			}
		}
		newType := reflect.StructOf(sf)
		newValue := value.Convert(newType)
		return json.Marshal(newValue.Interface())
	} else {
		return json.Marshal(v)
	}
}

// convert data to map[string]interface{}
func Imapify(data interface{}) map[string]interface{} {
	var err error
	m := make(map[string]interface{})

	switch v := data.(type) {
	case string:
		err = json.Unmarshal([]byte(v), &m)
	case []byte:
		err = json.Unmarshal(v, &m)
	default:
		if d, er := json.Marshal(v); er != nil {
			return nil
		} else {
			err = json.Unmarshal(d, &m)
		}
	}

	if err != nil {
		return nil
	}

	return m
}

// modify map[string]interface{}
func Imapset(data map[string]interface{}, keyPath string, val interface{}) error {
	ok := false
	keys := stringSplit(replaceEscapePeriod(keyPath, true), '.')

	for idx, key := range keys {
		kidx := -1
		key = replaceEscapePeriod(key, false)
		if key == "" {
			continue
		}
		if matched, _ := regexp.MatchString(`^\w+\[\d+\]$`, key); matched {
			res := regexp.MustCompile(`^(\w+)\[(\d+)\]$`).FindStringSubmatch(key)
			key = res[1]
			var err error
			if kidx, err = strconv.Atoi(res[2]); err != nil {
				return err
			}
		}

		if idx == len(keys)-1 {
			if val == nil {
				if kidx < 0 {
					delete(data, key)
				} else {
					if reflect.TypeOf(data[key]).Kind() != reflect.Slice {
						return fmt.Errorf("value of keyPath %s is not type []interface{}", keyPath)
					}
					var temp []interface{}
					v := reflect.ValueOf(data[key])
					if kidx >= v.Len() {
						return fmt.Errorf("keyPath %s index out of range", keyPath)
					}
					for i := 0; i < v.Len(); i++ {
						if i != kidx {
							temp = append(temp, v.Index(i).Interface())
						}
					}
					data[key] = temp
				}
			} else {
				if kidx < 0 {
					data[key] = val
				} else {
					if reflect.TypeOf(data[key]).Kind() != reflect.Slice {
						return fmt.Errorf("value of keyPath %s is not type []interface{}", keyPath)
					}
					var temp []interface{}
					v := reflect.ValueOf(data[key])
					if kidx >= v.Len() {
						return fmt.Errorf("keyPath %s index out of range", keyPath)
					}
					for i := 0; i < v.Len(); i++ {
						if i != kidx {
							temp = append(temp, v.Index(i).Interface())
						} else {
							temp = append(temp, val)
						}
					}
					data[key] = temp
				}
			}
		} else {
			if kidx < 0 {
				data, ok = data[key].(map[string]interface{})
				if !ok {
					return fmt.Errorf("value of keyPath %s is not type map[string]interface{}", keyPath)
				}
			} else {
				if reflect.TypeOf(data[key]).Kind() != reflect.Slice {
					return fmt.Errorf("value of keyPath %s is not type []interface{}", keyPath)
				}
				v := reflect.ValueOf(data[key])
				if kidx >= v.Len() {
					return fmt.Errorf("keyPath %s index out of range", keyPath)
				}
				data = v.Index(kidx).Interface().(map[string]interface{})
			}
		}
	}

	return nil
}

// convert data to json-like string
func Jsonify(v interface{}) string {
	return string(tojson(nil, v))
}

// get all leaf key paths of the json string
func Jsleaf(jsonData string, separator ...rune) []string {
	out := []string{}
	paths := Jsdig(jsonData, separator...)
	for idx, path := range paths {
		if idx < len(paths)-1 {
			if !stringContainStr(paths[idx+1], paths[idx]+"[") && !stringContainStr(paths[idx+1], paths[idx]+".") {
				out = append(out, path)
			}
		}
	}

	return out
}

// get all key paths of the json string
func Jsdig(jsonData string, separator ...rune) []string {
	sep := '.'
	if len(separator) > 0 {
		sep = separator[0]
	}

	out := []string{}
	mapDig(&out, "", Imapify(jsonData), sep)

	return out
}

// get differences between a and b json strings and return the corresponding key paths
func Jsdiff(a, b string, separator ...rune) []string {
	out := []string{}
	sep := '.'
	if len(separator) > 0 {
		sep = separator[0]
	}
	paths := Jsleaf(a, sep)
	for _, path := range paths {
		if compareInterface(Jsquery(a, path), Jsquery(b, path)) != 0 {
			out = append(out, path)
		}
	}

	return out
}

// set value of the path key from json string
func Jsmodify(jsonData string, keyPath string, val interface{}) string {
	data := Imapify(jsonData)
	_ = Imapset(data, keyPath, val)
	return Jsonify(data)
}

// get value of the path key from json string
func Jsquery(jsonData string, keyPath string) interface{} {
	var val interface{}

	val = Imapify(jsonData)
	keyPath = replaceEscapePeriod(keyPath, true)
	for _, p := range stringSplit(keyPath, '.') {
		p = replaceEscapePeriod(p, false)
		if matched, _ := regexp.MatchString(`^\[\d+\]$`, p); matched {
			if data, ok := val.([]interface{}); ok {
				if len(data) == 0 {
					val = data
					continue
				}
				sp := regexp.MustCompile(`^\[(\d+)\]$`).FindStringSubmatch(p)
				val = getJsonItem(data, sp[1])
			} else {
				return nil
			}
		} else if matched, _ := regexp.MatchString(`^\[\w+\s?-?\w?\]$`, p); matched {
			if data, ok := val.([]interface{}); ok {
				if len(data) == 0 {
					val = data
					continue
				}
				sp := regexp.MustCompile(`^\[(\w+\s?-?\w?)\]$`).FindStringSubmatch(p)
				val = getJsonItem(data, sp[1])
			} else {
				return nil
			}
		} else if matched, _ := regexp.MatchString(`^\*?\w+\*?#?\[\d+\]$`, p); matched {
			if data, ok := val.(map[string]interface{}); ok {
				sp := regexp.MustCompile(`^(\*?\w+\*?#?)\[(\d+)\]$`).FindStringSubmatch(p)
				val = getJsonVal(data, sp[1])
				if data, ok := val.([]interface{}); ok {
					val = getJsonItem(data, sp[2])
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else if matched, _ := regexp.MatchString(`^\*?\w+\*?#?\[\w+\s?-?\w?\]$`, p); matched {
			if data, ok := val.(map[string]interface{}); ok {
				sp := regexp.MustCompile(`^(\*?\w+\*?#?)\[(\w+\s?-?\w?)\]$`).FindStringSubmatch(p)
				val = getJsonVal(data, sp[1])
				if data, ok := val.([]interface{}); ok {
					if len(data) == 0 {
						val = data
						continue
					}
					val = getJsonItem(data, sp[2])
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			if data, ok := val.(map[string]interface{}); ok {
				val = getJsonVal(data, p)
			} else {
				return nil
			}
		}
	}
	return val
}

func getJsonVal(data map[string]interface{}, p string) interface{} {
	var key string
	var i int
	ki := stringSplit(p, '#')
	key = ki[0]
	if len(ki) > 1 {
		i, _ = strconv.Atoi(ki[1])
	}
	var keys []string
	if key[0] == '*' && key[len(key)-1] == '*' {
		for k := range data {
			if stringContainStr(k, key) {
				keys = append(keys, k)
			}
		}
	} else if key[0] == '*' {
		for k := range data {
			if stringSuffixStr(k, key[1:]) {
				keys = append(keys, k)
			}
		}
	} else if key[len(key)-1] == '*' {
		for k := range data {
			if stringPrefixStr(k, key[:len(key)-1]) {
				keys = append(keys, k)
			}
		}
	} else {
		keys = append(keys, key)
	}
	if i >= len(keys) {
		return nil
	}
	return data[keys[i]]
}

func getJsonItem(data []interface{}, p string) interface{} {
	if p == "first" {
		return data[0]
	} else if p == "last" {
		return data[len(data)-1]
	} else if p == "odd" {
		var tdata []interface{}
		for idx, d := range data {
			if idx%2 == 0 {
				tdata = append(tdata, d)
			}
		}
		return tdata
	} else if p == "even" {
		var tdata []interface{}
		for idx, d := range data {
			if idx%2 == 1 {
				tdata = append(tdata, d)
			}
		}
		return tdata
	} else if p == "#" || p == "len" {
		return len(data)
	} else if stringContainRune(p, ' ') {
		var tdata []interface{}
		for _, i := range stringSplit(p, ' ') {
			if idx, err := strconv.Atoi(i); err == nil {
				if idx < len(data) {
					tdata = append(tdata, data[idx])
				}
			}
		}
		return tdata
	} else if matched, _ := regexp.MatchString(`^\d+$`, p); matched {
		idx, _ := strconv.Atoi(p)
		if idx >= len(data) {
			return nil
		}
		return data[idx]
	} else if matched, _ := regexp.MatchString(`^\d+-\d+$`, p); matched {
		ssp := regexp.MustCompile(`^(\d+)-(\d+)$`).FindStringSubmatch(p)
		start, _ := strconv.Atoi(ssp[1])
		end, _ := strconv.Atoi(ssp[2])
		if start >= len(data) {
			start = len(data) - 1
		}
		if end >= len(data) {
			end = len(data) - 1
		}
		if start == end {
			return data[start]
		} else {
			if start < end {
				return data[start:end]
			} else {
				return data[end:start]
			}
		}
	} else {
		return nil
	}
}

func mapDig(result *[]string, root string, mp map[string]interface{}, sep rune) {
	for k, v := range mp {
		nroot := fmt.Sprintf("%s%c%s", root, sep, stringEscapeSep(k, sep))
		*result = append(*result, nroot)
		switch val := v.(type) {
		case map[string]interface{}:
			mapDig(result, nroot, val, sep)
		case []interface{}:
			for idx, obj := range val {
				sroot := fmt.Sprintf("%s[%d]", nroot, idx)
				*result = append(*result, sroot)
				if nval, ok := obj.(map[string]interface{}); ok {
					mapDig(result, sroot, nval, sep)
				}
			}
		}
	}
}

// func writeFile(logfile string, mode int, data []string) {
// 	if logfile == "" {
// 		fmt.Println("Can not get log file")
// 		return
// 	}
// 	file, err := os.OpenFile(logfile, mode, 0666)
// 	if err != nil {
// 		fmt.Println("Can not open log file: ", err)
// 		return
// 	}
// 	for _, data := range data {
// 		if _, err = file.WriteString(data); err != nil {
// 			fmt.Println("Write log file failed: ", err)
// 			return
// 		}
// 	}
// 	file.Close()
// }

func appendKey(dst []byte, key string) []byte {
	if len(dst) == 0 || dst[len(dst)-1] == '[' || dst[len(dst)-1] == ':' {
		dst = append(dst, '{')
	}

	// if c.buf[len(c.buf)-1] == '}' {
	// 	c.buf = c.buf[:len(c.buf)-1]
	// }

	if dst[len(dst)-1] != '{' {
		dst = append(dst, ',')
	}
	dst = appendStr(dst, key)
	return append(dst, ':')
}

func appendStrComplex(dst []byte, s string, i int) []byte {
	start := 0
	for i < len(s) {
		b := s[i]
		if b >= utf8.RuneSelf {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size == 1 {
				if start < i {
					dst = append(dst, s[start:i]...)
				}
				dst = append(dst, `\ufffd`...)
				i += size
				start = i
				continue
			}
			i += size
			continue
		}
		if b >= 0x20 && b != '\\' && b != '"' {
			i++
			continue
		}
		if start < i {
			dst = append(dst, s[start:i]...)
		}
		switch b {
		case '"', '\\':
			dst = append(dst, '\\', b)
		case '\b':
			dst = append(dst, '\\', 'b')
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, '\\', 'u', '0', '0', hexs[b>>4], hexs[b&0xF])
		}
		i++
		start = i
	}
	if start < len(s) {
		dst = append(dst, s[start:]...)
	}
	return dst
}

func appendBytesComplex(dst []byte, s []byte, i int) []byte {
	start := 0
	for i < len(s) {
		b := s[i]
		if b >= utf8.RuneSelf {
			r, size := utf8.DecodeRune(s[i:])
			if r == utf8.RuneError && size == 1 {
				if start < i {
					dst = append(dst, s[start:i]...)
				}
				dst = append(dst, `\ufffd`...)
				i += size
				start = i
				continue
			}
			i += size
			continue
		}
		if b >= 0x20 && b != '\\' && b != '"' {
			i++
			continue
		}
		if start < i {
			dst = append(dst, s[start:i]...)
		}
		switch b {
		case '"', '\\':
			dst = append(dst, '\\', b)
		case '\b':
			dst = append(dst, '\\', 'b')
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, '\\', 'u', '0', '0', hexs[b>>4], hexs[b&0xF])
		}
		i++
		start = i
	}
	if start < len(s) {
		dst = append(dst, s[start:]...)
	}
	return dst
}

func appendStr(dst []byte, str string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(str); i++ {
		if !(str[i] >= 0x20 && str[i] != '\\' && str[i] != '"') {
			dst = appendStrComplex(dst, str, i)
			return append(dst, '"')
		}
	}
	dst = append(dst, str...)
	return append(dst, '"')
}

func appendStrs(dst []byte, vals []string) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendStr(dst, vals[0])
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendStr(append(dst, ','), val)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendBytes(dst []byte, bs []byte) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(bs); i++ {
		if !(bs[i] >= 0x20 && bs[i] != '\\' && bs[i] != '"') {
			dst = appendBytesComplex(dst, bs, i)
			return append(dst, '"')
		}
	}
	dst = append(dst, bs...)
	return append(dst, '"')
}

func appendHex(dst []byte, s []byte) []byte {
	dst = append(dst, '"')
	for _, v := range s {
		dst = append(dst, hexs[v>>4], hexs[v&0x0f])
	}
	return append(dst, '"')
}

func appendJson(dst []byte, j []byte) []byte {
	return append(dst, j...)
}

func appendBool(dst []byte, b bool) []byte {
	if b {
		return append(dst, "true"...)
	} else {
		return append(dst, "false"...)
	}
}

func appendBools(dst []byte, vals []bool) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendBool(dst, vals[0])
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendBool(append(dst, ','), val)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendInt(dst []byte, val int) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

func appendInts(dst []byte, vals []int) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendInt8(dst []byte, val int8) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

func appendInts8(dst []byte, vals []int8) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendInt16(dst []byte, val int16) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

func appendInts16(dst []byte, vals []int16) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendInt32(dst []byte, val int32) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

func appendInts32(dst []byte, vals []int32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendInt64(dst []byte, val int64) []byte {
	return strconv.AppendInt(dst, val, 10)
}

func appendInts64(dst []byte, vals []int64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0], 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), val, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUint(dst []byte, val uint) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

func appendUints(dst []byte, vals []uint) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUint8(dst []byte, val uint8) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

func appendUints8(dst []byte, vals []uint8) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUint16(dst []byte, val uint16) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

func appendUints16(dst []byte, vals []uint16) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUint32(dst []byte, val uint32) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

func appendUints32(dst []byte, vals []uint32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUint64(dst []byte, val uint64) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

func appendUints64(dst []byte, vals []uint64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, vals[0], 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), val, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendFloat(dst []byte, val float64, bitSize int) []byte {
	switch {
	case math.IsNaN(val):
		return append(dst, `"NaN"`...)
	case math.IsInf(val, 1):
		return append(dst, `"+Inf"`...)
	case math.IsInf(val, -1):
		return append(dst, `"-Inf"`...)
	default:
		return strconv.AppendFloat(dst, val, 'f', -1, bitSize)
	}
}

func appendFloat32(dst []byte, val float32) []byte {
	return appendFloat(dst, float64(val), 32)
}

func appendFloats32(dst []byte, vals []float32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, float64(vals[0]), 32)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), float64(val), 32)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendFloat64(dst []byte, val float64) []byte {
	return appendFloat(dst, val, 64)
}

func appendFloats64(dst []byte, vals []float64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, vals[0], 32)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), val, 64)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendTime(dst []byte, t time.Time, format string) []byte {
	switch format {
	case "":
		return appendInt64(dst, t.Unix())
	case "UNIXMS":
		return appendInt64(dst, t.UnixNano()/1000000)
	case "UNIXMICRO":
		return appendInt64(dst, t.UnixNano()/1000)
	default:
		return append(t.AppendFormat(append(dst, '"'), format), '"')
	}
}

func appendTimes(dst []byte, vals []time.Time, format string) []byte {
	switch format {
	case "":
		return appendUnixTimes(dst, vals)
	case "UNIXMS":
		return appendUnixMsTimes(dst, vals)
	}
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = append(vals[0].AppendFormat(append(dst, '"'), format), '"')
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = append(t.AppendFormat(append(dst, ',', '"'), format), '"')
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUnixTimes(dst []byte, vals []time.Time) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0].Unix(), 10)
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), t.Unix(), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUnixMsTimes(dst []byte, vals []time.Time) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0].UnixNano()/1000000, 10)
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), t.UnixNano()/1000000, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendDuration(dst []byte, d time.Duration, unit time.Duration) []byte {
	return appendFloat64(dst, float64(d)/float64(unit))
}

func appendDurations(dst []byte, vals []time.Duration, unit time.Duration) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendDuration(dst, vals[0], unit)
	if len(vals) > 1 {
		for _, d := range vals[1:] {
			dst = appendDuration(append(dst, ','), d, unit)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendInterface(dst []byte, i interface{}) []byte {
	marshaled, err := json.Marshal(i)
	if err != nil {
		marshaled, err = Marshal(i)
		if err != nil {
			return appendStr(dst, fmt.Sprintf("marshaling error: %v", err))
		}
	}

	return append(dst, marshaled...)
}

func appendObject(dst []byte, o interface{}) []byte {
	return appendStr(dst, fmt.Sprintf("%v", o))
}

func appendIP(dst []byte, ip net.IP) []byte {
	return appendStr(dst, ip.String())
}

func appendIPs(dst []byte, vals []net.IP) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendStr(dst, vals[0].String())
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendStr(dst, val.String())
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendIPNet(dst []byte, ipn net.IPNet) []byte {
	return appendStr(dst, ipn.String())
}

func appendIPNets(dst []byte, vals []net.IPNet) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendStr(dst, vals[0].String())
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendStr(dst, val.String())
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendMac(dst []byte, mac net.HardwareAddr) []byte {
	return appendStr(dst, mac.String())
}

func appendMacs(dst []byte, vals []net.HardwareAddr) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendStr(dst, vals[0].String())
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendStr(dst, val.String())
		}
	}
	dst = append(dst, ']')
	return dst
}

func checkSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)

	return uint16(^sum)
}

func compareInterface(a, b interface{}) int {
	if a == nil && b != nil {
		return -1
	} else if a != nil && b == nil {
		return 1
	} else if a == nil && b == nil {
		return 0
	}

	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch {
		case aVal.Int() < bVal.Int():
			return -1
		case aVal.Int() == bVal.Int():
			return 0
		case aVal.Int() > bVal.Int():
			return 1
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch {
		case aVal.Uint() < bVal.Uint():
			return -1
		case aVal.Uint() == bVal.Uint():
			return 0
		case aVal.Uint() > bVal.Uint():
			return 1
		}
	case reflect.Float32, reflect.Float64:
		switch {
		case aVal.Float() < bVal.Float():
			return -1
		case aVal.Float() == bVal.Float():
			return 0
		case aVal.Float() > bVal.Float():
			return 1
		}
	case reflect.String:
		switch {
		case aVal.String() < bVal.String():
			return -1
		case aVal.String() == bVal.String():
			return 0
		case aVal.String() > bVal.String():
			return 1
		}
	default:
		ma, _ := json.Marshal(a)
		mb, _ := json.Marshal(b)
		return bytes.Compare(ma, mb)
	}

	return 0
}

func quickSortBool(list []bool, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && (!pvt || list[chigh]) {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && (pvt || !list[clow]) {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortBool(list, low, pivot-1)
		quickSortBool(list, pivot+1, high)
	}
}

func quickSortRune(list []rune, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortRune(list, low, pivot-1)
		quickSortRune(list, pivot+1, high)
	}
}

func quickSortByte(list []byte, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortByte(list, low, pivot-1)
		quickSortByte(list, pivot+1, high)
	}
}

func quickSortBytes(list [][]byte, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && bytes.Compare(pvt, list[chigh]) <= 0 {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && bytes.Compare(pvt, list[clow]) >= 0 {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortBytes(list, low, pivot-1)
		quickSortBytes(list, pivot+1, high)
	}
}

func quickSortInt(list []int, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortInt(list, low, pivot-1)
		quickSortInt(list, pivot+1, high)
	}
}

func quickSortInt8(list []int8, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortInt8(list, low, pivot-1)
		quickSortInt8(list, pivot+1, high)
	}
}

func quickSortInt16(list []int16, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortInt16(list, low, pivot-1)
		quickSortInt16(list, pivot+1, high)
	}
}

func quickSortInt32(list []int32, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortInt32(list, low, pivot-1)
		quickSortInt32(list, pivot+1, high)
	}
}

func quickSortInt64(list []int64, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortInt64(list, low, pivot-1)
		quickSortInt64(list, pivot+1, high)
	}
}

func quickSortUint(list []uint, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortUint(list, low, pivot-1)
		quickSortUint(list, pivot+1, high)
	}
}

func quickSortUint8(list []uint8, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortUint8(list, low, pivot-1)
		quickSortUint8(list, pivot+1, high)
	}
}

func quickSortUint16(list []uint16, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortUint16(list, low, pivot-1)
		quickSortUint16(list, pivot+1, high)
	}
}

func quickSortUint32(list []uint32, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortUint32(list, low, pivot-1)
		quickSortUint32(list, pivot+1, high)
	}
}

func quickSortUint64(list []uint64, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortUint64(list, low, pivot-1)
		quickSortUint64(list, pivot+1, high)
	}
}

func quickSortFloat32(list []float32, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortFloat32(list, low, pivot-1)
		quickSortFloat32(list, pivot+1, high)
	}
}

func quickSortFloat64(list []float64, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortFloat64(list, low, pivot-1)
		quickSortFloat64(list, pivot+1, high)
	}
}

func quickSortTime(list []time.Time, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && !pvt.After(list[chigh]) {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && !pvt.Before(list[clow]) {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortTime(list, low, pivot-1)
		quickSortTime(list, pivot+1, high)
	}
}

func quickSortDuration(list []time.Duration, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortDuration(list, low, pivot-1)
		quickSortDuration(list, pivot+1, high)
	}
}

func quickSortIP(list []net.IP, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && bytes.Compare(pvt, list[chigh]) <= 0 {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && bytes.Compare(pvt, list[clow]) >= 0 {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortIP(list, low, pivot-1)
		quickSortIP(list, pivot+1, high)
	}
}

func quickSortMac(list []net.HardwareAddr, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && bytes.Compare(pvt, list[chigh]) <= 0 {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && bytes.Compare(pvt, list[clow]) >= 0 {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortMac(list, low, pivot-1)
		quickSortMac(list, pivot+1, high)
	}
}

func quickSortString(list []string, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && pvt <= list[chigh] {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && pvt >= list[clow] {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortString(list, low, pivot-1)
		quickSortString(list, pivot+1, high)
	}
}

func quickSortInterface(list []interface{}, low, high int) {
	if high > low {
		clow := low
		chigh := high
		pvt := list[clow]
		for clow < chigh {
			for clow < chigh && compareInterface(pvt, list[chigh]) <= 0 {
				chigh--
			}
			list[clow] = list[chigh]
			for clow < chigh && compareInterface(pvt, list[clow]) >= 0 {
				clow++
			}
			list[chigh] = list[clow]
		}
		list[clow] = pvt
		pivot := clow

		quickSortInterface(list, low, pivot-1)
		quickSortInterface(list, pivot+1, high)
	}
}

// sort various types of slices with quick sort algorithm
func QuickSort(list interface{}) {
	switch v := list.(type) {
	case []byte:
		quickSortByte(v, 0, len(v)-1)
	case [][]byte:
		quickSortBytes(v, 0, len(v)-1)
	case []bool:
		quickSortBool(v, 0, len(v)-1)
	case []int:
		quickSortInt(v, 0, len(v)-1)
	case []int8:
		quickSortInt8(v, 0, len(v)-1)
	case []int16:
		quickSortInt16(v, 0, len(v)-1)
	case []int32:
		quickSortInt32(v, 0, len(v)-1)
	case []int64:
		quickSortInt64(v, 0, len(v)-1)
	case []uint:
		quickSortUint(v, 0, len(v)-1)
	case []uint16:
		quickSortUint16(v, 0, len(v)-1)
	case []uint32:
		quickSortUint32(v, 0, len(v)-1)
	case []uint64:
		quickSortUint64(v, 0, len(v)-1)
	case []float32:
		quickSortFloat32(v, 0, len(v)-1)
	case []float64:
		quickSortFloat64(v, 0, len(v)-1)
	case []string:
		quickSortString(v, 0, len(v)-1)
	case []time.Time:
		quickSortTime(v, 0, len(v)-1)
	case []time.Duration:
		quickSortDuration(v, 0, len(v)-1)
	case []net.IP:
		quickSortIP(v, 0, len(v)-1)
	case []net.HardwareAddr:
		quickSortMac(v, 0, len(v)-1)
	case []interface{}:
		quickSortInterface(v, 0, len(v)-1)
	}
}

// remove duplicate elements in various types of slices
func Uniq(list interface{}) []interface{} {
	out := []interface{}{}
	switch v := list.(type) {
	case []byte:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(byte) != val {
				out = append(out, val)
			}
		}
	case [][]byte:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || string(out[len(out)-1].([]byte)) != string(val) {
				out = append(out, val)
			}
		}
	case []bool:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(bool) != val {
				out = append(out, val)
			}
		}
	case []int:
		sort.Ints(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(int) != val {
				out = append(out, val)
			}
		}
	case []int8:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(int8) != val {
				out = append(out, val)
			}
		}
	case []int16:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(int16) != val {
				out = append(out, val)
			}
		}
	case []int32:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(int32) != val {
				out = append(out, val)
			}
		}
	case []int64:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(int64) != val {
				out = append(out, val)
			}
		}
	case []uint:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(uint) != val {
				out = append(out, val)
			}
		}
	case []uint16:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(uint16) != val {
				out = append(out, val)
			}
		}
	case []uint32:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(uint32) != val {
				out = append(out, val)
			}
		}
	case []uint64:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(uint64) != val {
				out = append(out, val)
			}
		}
	case []float32:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(float32) != val {
				out = append(out, val)
			}
		}
	case []float64:
		sort.Float64s(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(float64) != val {
				out = append(out, val)
			}
		}
	case []string:
		sort.Strings(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(string) != val {
				out = append(out, val)
			}
		}
	case []time.Time:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(time.Time) != val {
				out = append(out, val)
			}
		}
	case []time.Duration:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(time.Duration) != val {
				out = append(out, val)
			}
		}
	case []net.IP:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(net.IP).String() != val.String() {
				out = append(out, val)
			}
		}
	case []net.HardwareAddr:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1].(net.HardwareAddr).String() != val.String() {
				out = append(out, val)
			}
		}
	case []interface{}:
		QuickSort(v)
		for _, val := range v {
			if len(out) == 0 || out[len(out)-1] != val {
				out = append(out, val)
			}
		}
	}

	return out
}

// return the largest element in various types of slices
func Max(list interface{}) interface{} {
	var out interface{}
	switch v := list.(type) {
	case []byte:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(byte) {
				out = val
			}
		}
	case [][]byte:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if string(val) > string(out.([]byte)) {
				var tmp []byte
				copy(tmp, val)
				out = tmp
			}
		}
	case []bool:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val && !out.(bool) {
				out = val
			}
		}
	case []int:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(int) {
				out = val
			}
		}
	case []int8:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(int8) {
				out = val
			}
		}
	case []int16:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(int16) {
				out = val
			}
		}
	case []int32:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(int32) {
				out = val
			}
		}
	case []int64:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(int64) {
				out = val
			}
		}
	case []uint:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(uint) {
				out = val
			}
		}
	case []uint16:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(uint16) {
				out = val
			}
		}
	case []uint32:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(uint32) {
				out = val
			}
		}
	case []uint64:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(uint64) {
				out = val
			}
		}
	case []float32:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(float32) {
				out = val
			}
		}
	case []float64:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(float64) {
				out = val
			}
		}
	case []string:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(string) {
				out = val
			}
		}
	case []time.Time:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val.After(out.(time.Time)) {
				out = val
			}
		}
	case []time.Duration:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val > out.(time.Duration) {
				out = val
			}
		}
	case []net.IP:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val.String() > out.(net.IP).String() {
				out = val
			}
		}
	case []net.HardwareAddr:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val.String() > out.(net.HardwareAddr).String() {
				out = val
			}
		}
	case []interface{}:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if compareInterface(val, out) > 0 {
				out = val
			}
		}
	}

	return out
}

// return minimum element in various types of slices
func Min(list interface{}) interface{} {
	var out interface{}
	switch v := list.(type) {
	case []byte:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(byte) {
				out = val
			}
		}
	case [][]byte:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if string(val) < string(out.([]byte)) {
				var tmp []byte
				copy(tmp, val)
				out = tmp
			}
		}
	case []bool:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if !val && out.(bool) {
				out = val
			}
		}
	case []int:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(int) {
				out = val
			}
		}
	case []int8:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(int8) {
				out = val
			}
		}
	case []int16:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(int16) {
				out = val
			}
		}
	case []int32:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(int32) {
				out = val
			}
		}
	case []int64:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(int64) {
				out = val
			}
		}
	case []uint:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(uint) {
				out = val
			}
		}
	case []uint16:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(uint16) {
				out = val
			}
		}
	case []uint32:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(uint32) {
				out = val
			}
		}
	case []uint64:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(uint64) {
				out = val
			}
		}
	case []float32:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(float32) {
				out = val
			}
		}
	case []float64:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(float64) {
				out = val
			}
		}
	case []string:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(string) {
				out = val
			}
		}
	case []time.Time:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val.Before(out.(time.Time)) {
				out = val
			}
		}
	case []time.Duration:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val < out.(time.Duration) {
				out = val
			}
		}
	case []net.IP:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val.String() < out.(net.IP).String() {
				out = val
			}
		}
	case []net.HardwareAddr:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if val.String() < out.(net.HardwareAddr).String() {
				out = val
			}
		}
	case []interface{}:
		for idx, val := range v {
			if idx == 0 {
				out = val
			} else if compareInterface(val, out) < 0 {
				out = val
			}
		}
	}

	return out
}

// traverse from start to end with step by iterator
func Iter(v ...int) <-chan int {
	start := 0
	end := start
	step := 1
	c := make(chan int)
	if len(v) == 1 {
		end = v[0]
	} else if len(v) == 2 {
		start = v[0]
		end = v[1]
	} else if len(v) > 2 {
		start = v[0]
		end = v[1]
		step = v[2]
	}

	go func() {
		for start < end {
			c <- start
			start += step
		}
		close(c)
	}()
	return c
}

// return integer slice range start to end with step
func Slice(v ...int) []int {
	start := 0
	end := start
	step := 1
	s := []int{}
	if len(v) == 1 {
		end = v[0]
	} else if len(v) == 2 {
		start = v[0]
		end = v[1]
	} else if len(v) > 2 {
		start = v[0]
		end = v[1]
		step = v[2]
	}

	for start < end {
		s = append(s, start)
		start += step
	}

	return s
}

// reverse elements in the slice or string
func Reverse(list interface{}) interface{} {
	val := reflect.ValueOf(list)
	length := val.Len()
	out := make([]interface{}, length)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < length; i++ {
			out[length-1-i] = val.Index(i).Interface()
		}
	} else if val.Kind() == reflect.String {
		r := []rune(list.(string))
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r)
	}
	return out
}

// return all indices of special element in the slice or string
func Index(list interface{}, v interface{}) []int {
	out := []int{}

	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			if val.Index(i).Interface() == v {
				out = append(out, i)
			}
		}
	} else if val.Kind() == reflect.String && reflect.ValueOf(v).Kind() == reflect.String {
		src := list.(string)
		sub := v.(string)
		if len(sub) == 0 {
			out = append(out, 0)
		} else {
			bf := 0
			for {
				if len(src) >= len(sub) {
					if i := stringIndex(src, sub); i >= 0 {
						out = append(out, i+bf)
						bf += len(src[:i]) + len(sub)
						src = src[i+len(sub):]
					} else {
						break
					}
				} else {
					break
				}
			}
		}
	}

	return out
}

// split the slice or string
func Split(list interface{}, v interface{}) []interface{} {
	out := []interface{}{}
	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		tmps := []interface{}{}
		for i := 0; i < val.Len(); i++ {
			tmpv := val.Index(i).Interface()
			if tmpv == v {
				var dst []interface{}
				dst = append(dst, tmps...)
				if len(dst) > 0 {
					out = append(out, dst)
				}
				tmps = []interface{}{}
			} else {
				tmps = append(tmps, tmpv)
			}
		}
		if len(tmps) > 0 {
			out = append(out, tmps)
		}
	} else if val.Kind() == reflect.String && reflect.ValueOf(v).Kind() == reflect.String {
		src := list.(string)
		sub := v.(string)
		if len(sub) == 0 {
			for _, r := range src {
				out = append(out, string(r))
			}
		} else {
			for {
				if len(src) >= len(sub) {
					if i := stringIndex(src, sub); i >= 0 {
						if len(src[:i]) > 0 {
							out = append(out, src[:i])
						}
						src = src[i+len(sub):]
					} else {
						break
					}
				} else {
					break
				}
			}
			if len(src) > 0 {
				out = append(out, src)
			}
		}
	}

	return out
}

// return if special element is in the slice or string
func Contain(list interface{}, v interface{}) bool {
	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			if val.Index(i).Interface() == v {
				return true
			}
		}
	} else if val.Kind() == reflect.String && reflect.ValueOf(v).Kind() == reflect.String {
		return stringContainStr(list.(string), v.(string))
	}

	return false
}

// remove special element from the slice or string
func Remove(list interface{}, v interface{}) interface{} {
	out := []interface{}{}

	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			if val.Index(i).Interface() != v {
				out = append(out, val.Index(i).Interface())
			}
		}
	} else if val.Kind() == reflect.String && reflect.ValueOf(v).Kind() == reflect.String {
		src := list.(string)
		sub := v.(string)
		if len(sub) == 0 {
			return list
		}
		for {
			if len(src) >= len(sub) {
				if i := stringIndex(src, sub); i >= 0 {
					src = src[:i] + src[i+len(sub):]
				} else {
					break
				}
			} else {
				break
			}
		}
		return src
	}
	return out
}

// remove the elements at special index from the slice or string
func RemoveAt(list interface{}, idx ...int) interface{} {
	offset := 1
	out := []interface{}{}
	index := 0

	if len(idx) > 0 {
		index = idx[0]
	}
	if len(idx) > 1 {
		offset = idx[1]
	}

	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			if i < index || i >= index+offset {
				out = append(out, val.Index(i).Interface())
			}
		}
	} else if val.Kind() == reflect.String {
		src := list.(string)
		if len(src) >= index+offset {
			src = src[:index] + src[index+offset:]
		}
		return src
	}
	return out
}

// return the elements meet the filter function in the slice or string
func Filter(list interface{}, fn func(interface{}) bool) interface{} {
	var out []interface{}
	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			v := val.Index(i).Interface()
			if fn(v) {
				out = append(out, v)
			}
		}
	} else if val.Kind() == reflect.String {
		for _, r := range list.(string) {
			if fn(r) {
				out = append(out, string(r))
			}
		}
	}

	return out
}

// return rune length of the string
//
//	func Strlen(str string) int {
//		return len([]rune(str))
//	}
func Strlen(str string) int {
	length := 0
	for _, r := range []rune(str) {
		length += 1
		if unicode.Is(unicode.Han, r) {
			length += 1
		}
	}
	return length
}

// return substring by rune
func Substr(str string, pos, length int) string {
	runes := []rune(str)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

// convert to Upper
func Upper(str string) string {
	return toUpper(str)
}

// convert to Lower
func Lower(str string) string {
	return toLower(str)
}

// convert to Capitalize
func Capitalize(str string) string {
	return toCapitalize(str)
}

func Ch2Eng(str string) string {
	var runes []rune
	chars := map[rune]rune{'，': ',', '。': '.', '：': ':', '！': '!', '？': '?', '·': '`', '’': '\'', '”': '"', '（': ')', '）': ')', '《': '<', '》': '>', '【': ']', '】': ']'}
	for _, r := range str {
		if v, ok := chars[r]; ok {
			runes = append(runes, v)
		} else {
			runes = append(runes, r)
		}
	}
	return string(runes)
}

func StrOmit(str string, length int) string {
	r := []rune(str)
	if len(r) <= length {
		return str
	} else {
		return string(r[:length-1]) + "..."
	}
}

func StrAlign(str, placeholder, align string, length int) string {
	str = strings.TrimSpace(strings.ReplaceAll(str, "\n", "\\n"))
	if Strlen(str) >= length {
		return str
	}
	phnum := length - Strlen(str)
	left := strings.Repeat(placeholder, phnum/2)
	right := strings.Repeat(placeholder, phnum-phnum/2)
	switch align {
	case "left":
		return str + left + right
	case "right":
		return left + right + str
	default:
		return left + str + right
	}
}

// convert byte size to human readable format
func HumanSize(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f%ciB", float64(b)/float64(div), "KM1PEZY"[exp])
}

// convert human readable string to byte size
func ByteSize(str string) (uint64, error) {
	i := strings.IndexFunc(str, func(r rune) bool {
		return r != '.' && !unicode.IsDigit(r)
	})
	var multiplier float64 = 1
	var sizeSuffixes = "BKM1PEZY"
	if i > 0 {
		suffix := str[i:]
		multiplier = 0
		for j := 0; j < len(sizeSuffixes); j++ {
			base := string(sizeSuffixes[j])
			// M, MB, or MiB are all valid.
			if suffix == base || suffix == base+"B" || suffix == base+"iB" {
				sz := 1 << uint(j*10)
				multiplier = float64(sz)
				break
			}
		}
		if multiplier == 0 {
			return 0, fmt.Errorf("invalid multiplier suffix %q, expected one of %s", suffix, []byte(sizeSuffixes))
		}
		str = str[:i]
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil || val < 0 {
		return 0, fmt.Errorf("expected a non-negative number, got %q", str)
	}
	val *= multiplier
	return uint64(math.Ceil(val)), nil
}

func Seq(sample string, i int) string {
	if i < 1 {
		return ""
	}

	suffix := ""
	if strings.HasSuffix(sample, ".") {
		suffix = "."
		sample = strings.TrimSuffix(sample, ".")
	}

	switch sample {
	case "1":
		return fmt.Sprintf("%d", i) + suffix
	case "01":
		return fmt.Sprintf("%02d", i) + suffix
	case "001":
		return fmt.Sprintf("%03d", i) + suffix
	case "0001":
		return fmt.Sprintf("%04d", i) + suffix
	case "00001":
		return fmt.Sprintf("%05d", i) + suffix
	case "000001":
		return fmt.Sprintf("%06d", i) + suffix
	case "a":
		return fmt.Sprintf("%c", i+96) + suffix
	case "A":
		return fmt.Sprintf("%c", i+64) + suffix
	case "I":
		return int2roman(i) + suffix
	case "i":
		return strings.ToLower(int2roman(i)) + suffix
	case "一":
		return int2chinese(i) + suffix
	case "①":
		return int2cnum(i) + suffix
	case "⑴":
		return int2pnum(i) + suffix
	case "(1)":
		return fmt.Sprintf("(%d)", i) + suffix
	case "(01)":
		return fmt.Sprintf("(%02d)", i) + suffix
	case "(001)":
		return fmt.Sprintf("(%03d)", i) + suffix
	case "(0001)":
		return fmt.Sprintf("(%04d)", i) + suffix
	case "(00001)":
		return fmt.Sprintf("(%05d)", i) + suffix
	case "(000001)":
		return fmt.Sprintf("(%06d)", i) + suffix
	case "(a)":
		return fmt.Sprintf("(%c)", i+96) + suffix
	case "(A)":
		return fmt.Sprintf("(%c)", i+64) + suffix
	case "(I)":
		return fmt.Sprintf("(%s)", int2roman(i)) + suffix
	case "(i)":
		return fmt.Sprintf("(%s)", strings.ToLower(int2roman(i))) + suffix
	case "(一)":
		return fmt.Sprintf("(%s)", int2chinese(i)) + suffix
	default:
		return fmt.Sprintf("%d", i)
	}
}

func MemEst(num int) int {
	sizes := []int{1, 2, 4, 6, 8, 12, 16, 24, 32, 48, 64, 80, 96, 128, 256, 512, 1024, 2048, 4096}

	best := sizes[0]
	mindist := num - best
	if mindist < 0 {
		mindist = 0 - mindist
	}

	for _, size := range sizes {
		curdist := num - size
		if curdist < 0 {
			curdist = 0 - curdist
		}

		if curdist < mindist {
			mindist = curdist
			best = size
		}
	}

	return best
}

func DiskEst(num int) int {
	sizes := []int{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1500, 2000, 2500, 3000, 3500, 4000, 4500, 5000, 5500, 6000, 6500, 7000, 7500, 8000, 8500, 9000, 9500, 10000}

	num -= 90

	best := sizes[0]
	mindist := num - best
	if mindist < 0 {
		mindist = 0 - mindist
	}

	for _, size := range sizes {
		curdist := num - size
		if curdist < 0 {
			curdist = 0 - curdist
		}

		if curdist < mindist {
			mindist = curdist
			best = size
		}
	}

	return best
}

// func Index(str, sub string) []int {
// 	var positions []int
// 	if sub == "" {
// 		return positions // 如果子串为空，返回空数组
// 	}

// 	strs := []rune(str)
// 	subs := []rune(sub)

// 	offset := 0
// 	for i := range strs {
// 		if i < offset {
// 			continue
// 		}

// 		found := true
// 		for j := range subs {
// 			if strs[i+j] != subs[j] {
// 				found = false
// 				break
// 			}
// 		}
// 		if found {
// 			positions = append(positions, i)
// 			offset = i + 1
// 		}
// 	}

//		return positions
//	}

// split a slice into multiple smaller slices
func GroupSlice(slc []interface{}, num int64) [][]interface{} {
	max := int64(len(slc))

	if max <= num {
		return [][]interface{}{slc}
	}

	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}

	var segments = make([][]interface{}, 0)

	var start, end, i int64
	for i = 1; i <= quantity; i++ {
		end = i * num
		if i != quantity {
			segments = append(segments, slc[start:end])
		} else {
			segments = append(segments, slc[start:])
		}
		start = i * num
	}
	return segments
}

// func SortedKeys[T interface{}](m map[string]T) []string {
// 	keys := make([]string, 0, len(m))

// 	for k := range m {
// 		keys = append(keys, k)
// 	}

// 	sort.Sort(sort.StringSlice(keys))
// 	return keys
// }

// get executable file path
func ExePath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exePath := filepath.Dir(ex)
	return exePath
}

// get working path
func WorkPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// check if target is exist
func IsExist(target string) bool {
	if _, err := os.Stat(target); err == nil {
		return true

	} else if os.IsNotExist(err) {
		return false

	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false
	}
}

// check if target is directory
func IsDir(target string) bool {
	info, err := os.Stat(target)
	if os.IsNotExist(err) {
		return false
	}
	if info.IsDir() {
		return true
	} else {
		return false
	}
}

// list all files and directories in root folder
func ListAll(root string) []string {
	var files []string

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	return files
}

// read file line by line
func IterFile(fpath string) <-chan string {
	c := make(chan string)

	f, err := os.Open(fpath)
	if err != nil {
		// panic(fmt.Sprintf("read file %s fail: %s", fpath, err.Error()))
		return c
	}
	//defer f.Close()
	go func(fl *os.File) {
		buf := bufio.NewScanner(fl)
		defer fl.Close()

		for {
			if !buf.Scan() {
				break
			}
			c <- buf.Text()
		}
		close(c)
	}(f)

	return c
}

// read the entire contents of the file
func ReadFile(fpath string) []byte {
	f, err := os.Open(fpath)
	if err != nil {
		// panic(fmt.Sprintf("open file %s fail: %s", fpath, err.Error()))
		return nil
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		// panic(fmt.Sprintf("read file %s fail: %s", fpath, err.Error()))
		return nil
	}

	return bytes
}

// write to file
func WriteFile(fpath string, data []byte, append ...bool) error {
	mode := os.O_RDWR | os.O_CREATE
	if len(append) > 0 && append[0] {
		mode = mode | os.O_APPEND
	} else {
		mode = mode | os.O_TRUNC
	}
	file, err := os.OpenFile(fpath, mode, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	writer.Flush()
	return nil
}

// Strip comments in data, [commentSingle, commentMultiStart, commentMultiEnd] can be set
func StripComment(data []byte, commentSymbols ...string) []byte {
	reader := NewReader(bytes.NewReader(data))
	switch len(commentSymbols) {
	case 0:
	case 1:
		reader.commentSingle = commentSymbols[0]
		reader.commentMultiStart = ""
		reader.commentMultiEnd = ""
	case 2:
		reader.commentSingle = commentSymbols[0]
		reader.commentMultiStart = commentSymbols[1]
		reader.commentMultiEnd = commentSymbols[1]
	default:
		reader.commentSingle = commentSymbols[0]
		reader.commentMultiStart = commentSymbols[1]
		reader.commentMultiEnd = commentSymbols[2]
	}

	out, _ := ioutil.ReadAll(reader)
	return out
}

// count word frequency
func WordFrequency(fpath string, analysis func(string) []string, order ...bool) [][2]interface{} {
	var wordFrequencyMap = make(map[string]int)

	for line := range IterFile(fpath) {
		var arr []string
		if analysis != nil {
			arr = analysis(line)
		} else {
			arr = strings.FieldsFunc(line, func(c rune) bool {
				if !unicode.IsLetter(c) && !unicode.IsNumber(c) {
					return true
				}
				return false
			})
		}

		for _, v := range arr {
			if _, ok := wordFrequencyMap[v]; ok {
				wordFrequencyMap[v] = wordFrequencyMap[v] + 1
			} else {
				wordFrequencyMap[v] = 1
			}
		}
	}

	var wordFrequency [][2]interface{}
	for k, v := range wordFrequencyMap {
		wordFrequency = append(wordFrequency, [2]interface{}{k, v})
	}

	if len(order) > 0 && order[0] {
		sort.Slice(wordFrequency, func(i, j int) bool {
			if wordFrequency[i][1].(int) > wordFrequency[j][1].(int) {
				return true
			} else if wordFrequency[i][1].(int) == wordFrequency[j][1].(int) {
				if wordFrequency[i][0].(string) < wordFrequency[j][0].(string) {
					return true
				}
			}
			return false
		})
	}

	return wordFrequency
}

// generate random intergers
func RandInt(count int, v ...int64) []int64 {
	var min int64 = 0
	var max int64 = 100

	if len(v) == 1 {
		max = v[0]
	} else if len(v) > 1 {
		min = v[0]
		max = v[1]
	}

	out := []int64{}
	if min > max {
		return out
	}

	allCount := make(map[int64]struct{})
	maxBigInt := big.NewInt(max)
	for {
		i, _ := rd.Int(rd.Reader, maxBigInt)
		number := i.Int64()
		if i.Int64() >= min {
			_, ok := allCount[number]
			if !ok {
				out = append(out, number)
				allCount[number] = struct{}{}
			}
		}
		if len(out) >= count {
			return out
		}
	}
}

// generate random strings
func RandString(count int, src ...byte) string {
	rand.Seed(time.Now().UnixNano())
	if len(src) == 0 {
		src = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	idxBits := 6
	idxMask := 1<<idxBits - 1
	idxMax := 63 / idxBits
	b := make([]byte, count)

	for i, cache, remain := count-1, rand.Int63(), idxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		if idx := int(cache) & idxMask; idx < len(src) {
			b[i] = src[idx]
			i--
		}
		cache >>= idxBits
		remain--
	}

	return string(b)
}

// generate uuid
func Uuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// float64 fix
func ToFixed(num float64, precisions ...int) float64 {
	precision := 3
	if len(precisions) > 0 {
		precision = precisions[0]
	}

	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

// trim space
func TrimSpace(s string) string {
	s1 := strings.TrimSpace(strings.Replace(s, "	", " ", -1))
	regstr := "\\s{2,}"
	reg, _ := regexp.Compile(regstr)
	s2 := make([]byte, len(s1))
	copy(s2, s1)
	spc_index := reg.FindStringIndex(string(s2))
	for len(spc_index) > 0 {
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...)
		spc_index = reg.FindStringIndex(string(s2))
	}
	return string(s2)
}

// get current time
func TimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// get current timestamp
func StampNow() int64 {
	return time.Now().Unix()
}

// convert time to timestamp
func Time2Stamp(t string) int64 {
	stamp, _ := time.ParseInLocation("2006-01-02 15:04:05", t, time.Local)
	return stamp.Unix()
}

// convert timestamp to time
func Stamp2Time(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

// format duration to string
func FormatDuration(d time.Duration) string {
	return (time.Duration(d.Milliseconds()) * time.Millisecond).String()
}

// parse duration string
func ParseDuration(s string) (time.Duration, error) {
	var unitMap = map[string]int64{
		"ns": int64(time.Nanosecond),
		"us": int64(time.Microsecond),
		"µs": int64(time.Microsecond), // U+00B5 = micro symbol
		"μs": int64(time.Microsecond), // U+03BC = Greek letter mu
		"ms": int64(time.Millisecond),
		"s":  int64(time.Second),
		"m":  int64(time.Minute),
		"h":  int64(time.Hour),
		"d":  int64(time.Hour) * 24,
		"w":  int64(time.Hour) * 168,
	}

	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d int64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, fmt.Errorf("time: invalid duration %s", orig)
	}
	for s != "" {
		var (
			v, f  int64       // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, fmt.Errorf("time: invalid duration %s", orig)
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, fmt.Errorf("time: invalid duration %s", orig)
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, fmt.Errorf("time: invalid duration %s", orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, fmt.Errorf("time: missing unit in duration %s", orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, fmt.Errorf("time: unknown unit %s in duration %s", u, orig)
		}
		if v > (1<<63-1)/unit {
			// overflow
			return 0, fmt.Errorf("time: invalid duration %s", orig)
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += int64(float64(f) * (float64(unit) / scale))
			if v < 0 {
				// overflow
				return 0, fmt.Errorf("time: invalid duration %s", orig)
			}
		}
		d += v
		if d < 0 {
			// overflow
			return 0, fmt.Errorf("time: invalid duration %s", orig)
		}
	}

	if neg {
		d = -d
	}
	return time.Duration(d), nil
}

// get timestamp range
func TsRange(ranger string) (int64, int64) {
	var st, et time.Time

	if _, err := strconv.Atoi(ranger); err == nil {
		ranger = ranger + "d"
	}

	if dur, err := ParseDuration(ranger); err == nil {
		et = time.Now().Local()
		st = et.Add(-1 * dur)
	} else {
		est := strings.Split(ranger, "--")
		ststr := strings.TrimSpace(est[0])
		if !strings.Contains(ststr, " ") {
			st, _ = time.ParseInLocation("2006-01-02", ststr, time.Local)
		} else {
			st, _ = time.ParseInLocation("2006-01-02 15:04:05", ststr, time.Local)
		}
		et = time.Now().Local()
		if len(est) > 1 {
			etstr := strings.TrimSpace(est[1])
			if !strings.Contains(etstr, " ") {
				et, _ = time.ParseInLocation("2006-01-02", etstr, time.Local)
			} else {
				et, _ = time.ParseInLocation("2006-01-02 15:04:05", etstr, time.Local)
			}
		}
	}

	return st.UnixMilli(), et.UnixMilli()
}

// gzip compresses the given data
func Gzip(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		// panic(err)
		return nil
	}
	if err := w.Close(); err != nil {
		// panic(err)
		return nil
	}
	return buf.Bytes()
}

// gunzip uncompresses the given data
func Gunzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

// archive target folder
func Zip(source, target string, filter ...string) (err error) {
	if isAbs := filepath.IsAbs(source); !isAbs {
		source, err = filepath.Abs(source)
		if err != nil {
			return err
		}
	}

	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}

	defer func() {
		if err := zipfile.Close(); err != nil {
			return
			// Errorf("file close error: %s, file: %s", err.Error(), zipfile.Name())
		}
	}()

	zw := zip.NewWriter(zipfile)

	defer func() {
		if err := zw.Close(); err != nil {
			return
			// Errorf("zipwriter close error: %s", err.Error())
		}
	}()

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if len(filter) > 0 {
			ism, err := filepath.Match(filter[0], info.Name())

			if err != nil {
				return err
			}

			if ism {
				return nil
			}
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, stringTrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer func() {
			if err := file.Close(); err != nil {
				return
				// Errorf("file close error: %s, file: %s", err.Error(), file.Name())
			}
		}()
		_, err = io.Copy(writer, file)

		return err
	})

	if err != nil {
		return err
	}

	return nil
}

// unzip target archived file
func Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		unzippath := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(unzippath, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(unzippath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

// retry
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if err.Error() == "retry-stop" {
			return err
		}

		if attempts--; attempts > 0 {
			// Warnf("retry func error: %s. attemps #%d after %s.", err.Error(), attempts, sleep)
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

// base64 encode
func Encode(src string) string {
	return base64.StdEncoding.EncodeToString([]byte(src))
}

// base64 decode
func Decode(src string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// encode url
func UrlEncodeBase64(str string) string {
	str = strings.ReplaceAll(str, "+", ".")
	str = strings.ReplaceAll(str, "/", "_")
	str = strings.ReplaceAll(str, "=", "-")
	return str
}

// decode url
func UrlDecodeBase64(str string) string {
	str = strings.ReplaceAll(str, ".", "+")
	str = strings.ReplaceAll(str, "_", "/")
	str = strings.ReplaceAll(str, "-", "=")
	return str
}

// encrypt src data with aes algorithm
func Encrypt(src []byte, keyStr string) ([]byte, error) {
	if len(keyStr) < 16 {
		keyStr = fmt.Sprintf("%16s", keyStr)
	} else if len(keyStr) < 24 {
		keyStr = fmt.Sprintf("%24s", keyStr)
	} else if len(keyStr) < 32 {
		keyStr = fmt.Sprintf("%32s", keyStr)
	} else {
		keyStr = keyStr[:32]
	}

	key := []byte(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padnum := block.BlockSize() - len(src)%block.BlockSize()
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	src = append(src, pad...)
	blockmode := cipher.NewCBCEncrypter(block, key)
	blockmode.CryptBlocks(src, src)
	return src, nil
}

// decrypt src data with aes algorithm
func Decrypt(src []byte, keyStr string) ([]byte, error) {
	if len(keyStr) < 16 {
		keyStr = fmt.Sprintf("%16s", keyStr)
	} else if len(keyStr) < 24 {
		keyStr = fmt.Sprintf("%24s", keyStr)
	} else if len(keyStr) < 32 {
		keyStr = fmt.Sprintf("%32s", keyStr)
	} else {
		keyStr = keyStr[:32]
	}

	key := []byte(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockmode := cipher.NewCBCDecrypter(block, key)
	blockmode.CryptBlocks(src, src)
	n := len(src)
	unpadnum := int(src[n-1])
	return src[:n-unpadnum], nil
}

// encrypt string with aes algorithm
func EncryptStr(str string, keyStr string) (string, error) {
	enc, err := Encrypt([]byte(str), keyStr)
	if err != nil {
		return "", err
	}

	return Encode(string(enc)), nil
}

// decrypt string with aes algorithm
func DecryptStr(str string, keyStr string) (string, error) {
	dec, err := Decode(str)
	if err != nil {
		return "", err
	}

	decb, err := Decrypt([]byte(dec), keyStr)
	if err != nil {
		return "", err
	}
	return string(decb), nil
}

// generate asymmetric key pair
func GenKeyPair() (privateKey string, publicKey string, e error) {
	priKey, err := ecdsa.GenerateKey(elliptic.P256(), rd.Reader)
	if err != nil {
		return "", "", err
	}
	ecPrivateKey, err := x509.MarshalECPrivateKey(priKey)
	if err != nil {
		return "", "", err
	}
	privateKey = base64.StdEncoding.EncodeToString(ecPrivateKey)

	X := priKey.X
	Y := priKey.Y
	xStr, err := X.MarshalText()
	if err != nil {
		return "", "", err
	}
	yStr, err := Y.MarshalText()
	if err != nil {
		return "", "", err
	}
	public := string(xStr) + "+" + string(yStr)
	publicKey = base64.StdEncoding.EncodeToString([]byte(public))
	return
}

// build asymmetric private key from private key string
func BuildPrivateKey(privateKeyStr string) (priKey *ecdsa.PrivateKey, e error) {
	bytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	priKey, err = x509.ParseECPrivateKey(bytes)
	if err != nil {
		return nil, err
	}
	return
}

// build asymmetric public key from public key string
func BuildPublicKey(publicKeyStr string) (pubKey *ecdsa.PublicKey, e error) {
	bytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}
	split := stringSplit(string(bytes), '+')
	xStr := split[0]
	yStr := split[1]
	x := new(big.Int)
	y := new(big.Int)
	err = x.UnmarshalText([]byte(xStr))
	if err != nil {
		return nil, err
	}
	err = y.UnmarshalText([]byte(yStr))
	if err != nil {
		return nil, err
	}
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	pubKey = &pub
	return
}

// sign content by private key string
func Sign(content []byte, privateKeyStr string) (signature string, e error) {
	priKey, err := BuildPrivateKey(privateKeyStr)
	if err != nil {
		return "", err
	}
	r, s, err := ecdsa.Sign(rd.Reader, priKey, []byte(Hash(content)))
	if err != nil {
		return "", err
	}
	rt, _ := r.MarshalText()
	st, _ := s.MarshalText()
	signStr := string(rt) + "+" + string(st)
	signature = hex.EncodeToString([]byte(signStr))
	return
}

// verify sign by public key string
func VerifySign(content []byte, signature string, publicKeyStr string) bool {
	decodeSign, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	split := stringSplit(string(decodeSign), '+')
	rStr := split[0]
	sStr := split[1]
	rr := new(big.Int)
	ss := new(big.Int)
	_ = rr.UnmarshalText([]byte(rStr))
	_ = ss.UnmarshalText([]byte(sStr))
	pubKey, err := BuildPublicKey(publicKeyStr)
	if err != nil {
		return false
	}
	return ecdsa.Verify(pubKey, []byte(Hash(content)), rr, ss)
}

// generate sha256 code for data
func Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}

// generate md5 code for data
func Md5(data []byte) string {
	sum := md5.Sum(data)
	return fmt.Sprintf("%x", sum)
}

func IsBase64Encoded(str string) bool {
	bs64pattern := regexp.MustCompile(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`)
	if !bs64pattern.MatchString(str) {
		return false
	}
	// if len(str)%4 != 0 {
	// 	return false
	// }
	if data, err := base64.StdEncoding.DecodeString(str); err != nil {
		return false
	} else {
		return utf8.Valid(data)
	}
}

// execute command with realtime output
func Exec(command string, args ...string) <-chan string {
	out := make(chan string, 1000)

	if len(args) == 0 && stringIndex(command, " ") >= 0 {
		args = []string{"-c", command}
		command = "bash"
	}

	go func(c string, a ...string) {
		defer close(out)

		cmd := exec.Command(c, a...)

		stdout, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout
		_ = cmd.Start()

		scanner := bufio.NewScanner(stdout)
		// scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			m := scanner.Text()
			out <- m
			// fmt.Println(m)
		}
		_ = cmd.Wait()
	}(command, args...)

	return out
}

// wait until the time
func WaitUntil(t string) {
	datetime := strings.Split(t, " ")
	var datestr, timestr string
	if len(datetime) < 2 {
		timestr = datetime[0]
	} else {
		datestr = datetime[0]
		timestr = datetime[1]
	}

	ts := strings.Split(timestr, ":")
	hour, _ := strconv.Atoi(ts[0])
	minute := 0
	second := 0
	if len(ts) > 1 {
		minute, _ = strconv.Atoi(ts[1])
	}
	if len(ts) > 2 {
		second, _ = strconv.Atoi(ts[2])
	}
	t = fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)

	ct := time.Now().Local()

	year := ct.Year()
	month := int(ct.Month())
	day := ct.Day()
	if datestr != "" {
		ymd := strings.Split(datestr, "-")
		year, _ = strconv.Atoi(ymd[0])
		month = 1
		day = 1

		if len(ymd) > 1 {
			month, _ = strconv.Atoi(ymd[1])
		}
		if len(ymd) > 2 {
			day, _ = strconv.Atoi(ymd[2])
		}
	}
	d := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	loc, _ := time.LoadLocation("Local")

	dt, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", d, t), loc)

	for dt.Before(ct) {
		dt = dt.Add(time.Hour * 24)
	}

	select {
	case <-time.After(dt.Sub(ct)):
		return
	}
}

func TimeConvert(kind, unit string, val ...interface{}) (interface{}, error) {
	var t time.Time
	if len(val) > 0 && val[0] != nil {
		var ts int64
		v := val[0]
		if val, ok := v.(float64); ok {
			ts = int64(val)
		} else if iv, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64); err == nil {
			ts = iv
		}
		if ts > 0 {
			length := len(strconv.FormatInt(ts, 10))
			switch {
			case length <= 10:
				t = time.Unix(ts, 0)
			case length > 10 && length <= 13:
				t = time.UnixMilli(ts)
			case length > 13 && length <= 16:
				t = time.UnixMicro(ts)
			case length > 16:
				t = time.Unix(0, ts)
			}
		} else {
			if ft, err := time.ParseInLocation("15:04:05", fmt.Sprintf("%v", v), time.Local); err != nil {
				return "", err
			} else {
				t = ft
			}
		}
	} else {
		t = time.Now().Local()
	}

	switch kind {
	case "time":
		switch unit {
		case "s", "sec", "second":
			return t.Format("15:04:05"), nil
		case "ms", "msec", "millisecond":
			return t.Format("15:04:05.000"), nil
		case "us", "usec", "microsecond":
			return t.Format("15:04:05.000000"), nil
		case "ns", "nsec", "nanosecond":
			return t.Format("15:04:05.000000000"), nil
		default:
			return t.Format("15:04:05"), nil
		}
	case "date":
		return t.Format("2006-01-02"), nil
	case "datetime":
		switch unit {
		case "s", "sec", "second":
			return t.Format("2006-01-02 15:04:05"), nil
		case "ms", "msec", "millisecond":
			return t.Format("2006-01-02 15:04:05.000"), nil
		case "us", "usec", "microsecond":
			return t.Format("2006-01-02 15:04:05.000000"), nil
		case "ns", "nsec", "nanosecond":
			return t.Format("2006-01-02 15:04:05.000000000"), nil
		default:
			return t.Format("2006-01-02 15:04:05"), nil
		}
	case "ts", "timestamp":
		switch unit {
		case "s", "sec", "second":
			return t.Unix(), nil
		case "ms", "msec", "millisecond":
			return t.UnixMilli(), nil
		case "us", "usec", "microsecond":
			return t.UnixMicro(), nil
		case "ns", "nsec", "nanosecond":
			return t.UnixNano(), nil
		default:
			return t.Unix(), nil
		}
	default:
		return "", fmt.Errorf("unsupported time kind: %s", kind)
	}
}

// if the string is an ip address
func IsIP(str string) bool {
	matched, _ := regexp.MatchString(`^^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}$`, str)
	return matched
}

// switch length to subnetmask
func LenToSubNetMask(subnet int) string {
	var buff bytes.Buffer
	for i := 0; i < subnet; i++ {
		buff.WriteString("1")
	}
	for i := subnet; i < 32; i++ {
		buff.WriteString("0")
	}
	masker := buff.String()
	a, _ := strconv.ParseUint(masker[:8], 2, 64)
	b, _ := strconv.ParseUint(masker[8:16], 2, 64)
	c, _ := strconv.ParseUint(masker[16:24], 2, 64)
	d, _ := strconv.ParseUint(masker[24:32], 2, 64)
	resultMask := fmt.Sprintf("%v.%v.%v.%v", a, b, c, d)
	return resultMask
}

// switch subnetmask to length
func SubNetMaskToLen(netmask string) (int, error) {
	ipSplitArr := strings.Split(netmask, ".")
	if len(ipSplitArr) != 4 {
		return 0, fmt.Errorf("netmask:%v is not valid, pattern should like: 255.255.255.0", netmask)
	}
	ipv4MaskArr := make([]byte, 4)
	for i, value := range ipSplitArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("ipMaskToInt call strconv.Atoi error:[%v] string value is: [%s]", err, value)
		}
		if intValue > 255 {
			return 0, fmt.Errorf("netmask cannot greater than 255, current value is: [%s]", value)
		}
		ipv4MaskArr[i] = byte(intValue)
	}

	ones, _ := net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
	return ones, nil
}

// get ips from subnet
func GetIPsInSubNet(ipAddr string, subMask ...string) ([]string, error) {
	subnetMask := subMask[0]
	if subnetMask == "" && strings.Contains(ipAddr, "/") {
		ipmask := strings.Split(ipAddr, "/")
		ipAddr = ipmask[0]
		subnetMask = ipmask[1]
	}
	if subnet, err := strconv.Atoi(subnetMask); err == nil {
		subnetMask = LenToSubNetMask(subnet)
	}
	ip := net.ParseIP(ipAddr)
	mask := net.IPMask(net.ParseIP(subnetMask).To4())
	// 获取 IP 地址所在的网络
	_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip, countBits(mask)))
	if err != nil {
		return nil, err
	}
	// 遍历网络中的 IP 地址
	ips := []string{}
	for {
		ip = ip_inc(ip)
		if ipNet.Contains(ip) == false {
			break
		}
		ips = append(ips, ip.String())
	}
	return ips, nil
}

// 计算子网掩码中的位数
func countBits(mask net.IPMask) int {
	count := 0
	for _, b := range mask {
		for b > 0 {
			b &= (b - 1)
			count++
		}
	}
	return count
}

// IP地址加1的操作
func ip_inc(ip net.IP) net.IP {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	return ip
}

func ParseStruct(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)

	// Ensure we're dealing with a struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference the pointer if it's a pointer to a struct
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct or a pointer to a struct")
	}

	type StructInfo struct {
		Name  string      `json:"name"`
		Type  string      `json:"type"`
		Tag   string      `json:"tag,omitempty"`
		Value interface{} `json:"value"`
	}

	typeOfS := val.Type()
	var fieldsInfo []StructInfo

	for i := 0; i < val.NumField(); i++ {
		field := typeOfS.Field(i)
		fieldValue := val.Field(i)

		info := StructInfo{
			Name:  field.Name,
			Type:  field.Type.String(),
			Tag:   string(field.Tag), // Convert StructTag to string
			Value: fieldValue.Interface(),
		}
		fieldsInfo = append(fieldsInfo, info)
	}

	return json.MarshalIndent(fieldsInfo, "", "  ")
}

// get ips from range
func GetIPsInRange(ipstr string) []string {
	var ips []string
	for _, iprstr := range strings.Split(ipstr, ",") {
		iprstr = strings.TrimSpace(iprstr)
		ipr := strings.Split(iprstr, "-")
		if len(ipr) > 1 {
			start := ipr[0]
			end := ipr[1]
			if !strings.Contains(end, ".") {
				ss := strings.Split(start, ".")
				ss[3] = end
				end = strings.Join(ss, ".")
			}

			startIp := net.ParseIP(start)
			ips = append(ips, start)
			for {
				ip := ip_inc(startIp)
				if ip.String() == end || len(ips) > 1024 {
					break
				}
				ips = append(ips, ip.String())
			}
			ips = append(ips, end)
		} else {
			if strings.Contains(iprstr, "/") {
				ips, _ = GetIPsInSubNet(iprstr)
			} else {
				ips = append(ips, iprstr)
			}
		}
	}

	return ips
}

// get ports from range
func GetPortsInRange(portstr string) []int {
	var ports []int
	for _, portrstr := range strings.Split(portstr, ",") {
		portstr = strings.TrimSpace(portrstr)
		portrs := strings.Split(portstr, "-")
		if len(portrs) > 1 {
			start, err := strconv.Atoi(portrs[0])
			if err != nil {
				continue
			}
			end, err := strconv.Atoi(portrs[1])
			if err != nil {
				continue
			}

			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			if tmp, err := strconv.Atoi(portstr); err == nil {
				ports = append(ports, tmp)
			}
		}
	}

	return ports
}

// get function name
func FuncName(i interface{}) string {
	if i == nil {
		return "nil"
	}

	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// string anonymization
func Desensitize(str string) string {
	dstr := str

	// 定义所有需要脱敏的正则表达式及其替换占位符
	// 注意：匹配顺序很重要，通常从更具体的模式到更通用的模式
	// 例如，身份证号（更具体）应在通用数字串（更通用）之前匹配

	// 1. 身份证号 (中国大陆18位身份证号)
	// 15位: [1-9]\d{7}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}
	// 18位: [1-9]\d{5}[1-9]\d{3}((0\d)|(1[0-2]))(([0|1|2]\d)|3[0-1])\d{3}[\dxX]
	reIDCard := regexp.MustCompile(`\b[1-9]\d{5}(18|19|([23]\d))\d{2}((0\d)|(10|11|12))(([0|1|2]\d)|3[0-1])\d{3}[\dxX]\b`)
	dstr = reIDCard.ReplaceAllString(dstr, "[ID_CARD_NUM]")

	// 2. 银行卡号 (13-19位数字，且通常是连续的数字串)
	// 通常，银行卡号没有固定的前缀或后缀，所以匹配连续的数字串
	reBankCard := regexp.MustCompile(`\b\d{13,19}\b`) // 匹配13到19位数字
	dstr = reBankCard.ReplaceAllStringFunc(dstr, func(s string) string {
		// 避免误伤其他长数字串，可以增加一些启发式规则
		// 例如：如果这个数字串附近没有“卡号”、“账户”等关键词，则不替换
		// 或者，如果它同时匹配到日期、时间等其他更具体的正则，则让更具体的正则优先处理
		// 这里简化处理，直接替换
		return "[BANK_CARD_NUM]"
	})

	// 3. IP 地址 (IPv4)
	reIP := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	dstr = reIP.ReplaceAllString(dstr, "[IP_ADDR]")

	// reIPv6 := regexp.MustCompile(`\b((([0-9a-fA-F]{1,4}:){7}([0-9a-fA-F]{1,4}|:))|(([0-9a-fA-F]{1,4}:){6}(:[0-9a-fA-F]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9a-fA-F]{1,4}:){5}(((:[0-9a-fA-F]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9a-fA-F]{1,4}:){4}(((:[0-9a-fA-F]{1,4}){1,3})|((:[0-9a-fA-F]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9a-fA-F]{1,4}:){3}(((:[0-9a-fA-F]{1,4}){1,4})|((:[0-9a-fA-F]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9a-fA-F]{1,4}:){2}(((:[0-9a-fA-F]{1,4}){1,5})|((:[0-9a-fA-F]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9a-fA-F]{1,4}:){1}(((:[0-9a-fA-F]{1,4}){1,6})|((:[0-9a-fA-F]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9a-fA-F]{1,4}){1,7})|((:[0-9a-fA-F]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\b`)
	// dstr = reIPv6.ReplaceAllString(dstr, "[IPv6_ADDR]")

	// 4. Mac 地址
	reMac := regexp.MustCompile(`\b([0-9a-fA-F]{2}([.:-])){5}[0-9a-fA-F]{2}\b`)
	dstr = reMac.ReplaceAllString(dstr, "[MAC_ADDR]")

	// 4. 日期 (多种常见格式，例如 YYYY-MM-DD, YYYY/MM/DD)
	reDate := regexp.MustCompile(`\b\d{4}[-/]\d{1,2}[-/]\d{1,2}\b`)
	dstr = reDate.ReplaceAllString(dstr, "[DATE]")

	// 5. 时间 (H:M:S 或 H:M:S-ms，带或不带时区偏移)
	reTime := regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:[-,+][\d]{3,4})?(?:[\d]{2})?\b`)
	dstr = reTime.ReplaceAllString(dstr, "[TIME]")

	reDateTime := regexp.MustCompile(`\b\d{4}[-/]\d{1,2}[-/]\d{1,2}T\d{2}:\d{2}:\d{2}(?:[-,+][\d]{3,4})?(?:[\d]{2})?\b`)
	dstr = reDateTime.ReplaceAllString(dstr, "[DATETIME]")

	// 6. 时间戳 (Unix timestamp，秒级或毫秒级) ---
	// 匹配10位数字（秒级）或13位数字（毫秒级）的时间戳
	// 注意：这可能与某些长ID号冲突，所以如果ID号有固定前缀，最好使用更精确的ID正则
	reTimestamp := regexp.MustCompile(`\b\d{10}(?:\d{3})?\b`) // 匹配10位或13位数字
	dstr = reTimestamp.ReplaceAllStringFunc(dstr, func(s string) string {
		// 结合上下文判断，例如前面有 "createTime": 或 "timestamp": 等关键词
		// 这是一个简单的例子，实际可能需要更复杂的逻辑
		// if strings.Contains(log, `"createTime":`+s) || strings.Contains(log, `"timestamp":`+s) {
		//     return "[TIMESTAMP]"
		// }
		// 为了简单，这里直接替换符合长度的数字串
		return "[TIMESTAMP]"
	})

	// 7. 电话号码 (常见的手机号码格式)
	rePhone := regexp.MustCompile(`\b(1[3-9]\d{9}|0\d{2,3}-?\d{7,8})\b`)
	dstr = rePhone.ReplaceAllString(dstr, "[PHONE_NUM]")

	// 8. 邮箱地址
	reEmail := regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	dstr = reEmail.ReplaceAllString(dstr, "[EMAIL_ADDR]")

	// 9. Token / API Key (非常通用，需要根据实际Token格式调整)
	// 示例：Base64编码的字符串 (由大小写字母、数字、+、/ 组成，末尾可能有=)
	// 或 UUID 格式（UUID已在通用ID中处理）
	// 通常Token会非常长，且可能包含特殊字符
	reToken := regexp.MustCompile(`\b[a-zA-Z0-9+/=_-]{32,128}\b`) // 匹配32到128位Base64或类似Token
	// 还可以匹配特定的前缀：例如 "Bearer\s+[a-zA-Z0-9+/=_-]+"
	// reAPIKey := regexp.MustCompile(`\b(?:API_KEY|TOKEN|SECRET)[=:]\s*([a-zA-Z0-9+/=_-]{16,})\b`)
	dstr = reToken.ReplaceAllString(dstr, "[TOKEN]")

	// 10. 姓名 (简单的中文姓名，2-4个汉字)
	// 这是非常基础的匹配，实际情况中姓名识别非常复杂
	// 放弃识别人名
	// reName := regexp.MustCompile(`\p{Han}{2,4}[0-9]{0,2}`)
	// dstr = reName.ReplaceAllStringFunc(dstr, func(s string) string {
	// 	commonWords := map[string]bool{
	// 		"失败": true, "成功": true, "错误": true, "信息": true, "用户": true,
	// 		"创建": true, "更新": true, "删除": true, "查询": true, "服务": true,
	// 		"测试": true, "系统": true, "接口": true, "任务": true,
	// 	}
	// 	if commonWords[s] {
	// 		return s
	// 	}
	// 	return "[NAME]"
	// })

	// 11. 编号/ID (通用数字序列或UUID) - 放到较后，避免与银行卡号等冲突
	// UUID 格式
	reUUID := regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)
	dstr = reUUID.ReplaceAllString(dstr, "[UUID]")

	// 查找方括号里的数字（例如线程ID，行号等）
	// 不替换线程ID, 行号等，避免误伤
	// reNumInBrackets := regexp.MustCompile(`\[\d+\]`)
	// dstr = reNumInBrackets.ReplaceAllString(dstr, "[NUM_ID]")

	// 查找引号内的数字串 (例如 "logisticsNumber":"SF0269595341663")
	// 注意：这可能与银行卡号重叠，因此银行卡号的正则必须优先匹配
	reNumInQuotes := regexp.MustCompile(`"[a-zA-Z]*\d{5,}[a-zA-Z]*"`) // 匹配引号内包含5位以上数字的字符串
	dstr = reNumInQuotes.ReplaceAllString(dstr, `"[GENERIC_NUM_ID]"`)

	// 查找通用长数字串（例如订单号、客户ID等），避免与日期时间、电话、卡号重叠
	// 放在最后，作为兜底，且避免误伤。
	// 这里更倾向于替换独立存在的长数字串
	reLongNum := regexp.MustCompile(`\b\d{5,}\b`) // 匹配5位或更长的连续数字
	dstr = reLongNum.ReplaceAllStringFunc(dstr, func(s string) string {
		// 再次检查是否可能是IP、日期、时间的一部分，或者已经被其他更具体的正则处理过
		// 由于正则匹配是从左到右进行的，并且我们是按顺序替换的，
		// 理论上前面更具体的正则会优先处理。
		// 这里可以增加一个判断，例如，如果这个数字串是已知的常见数字（如端口号），则不替换。
		// 简单起见，这里直接替换
		return "[LONG_NUM]"
	})

	return dstr
}

// parse url params
func UrlParams(rawUrl string) (map[string][]string, error) {
	stUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	m := stUrl.Query()
	return m, nil
}

func request(method, url string, headers map[string]interface{}, payload []byte, timeout time.Duration) (resp *http.Response, err error) {
	client := &http.Client{
		Timeout: timeout * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(payload)))
	if err != nil {
		return
	}

	for k, v := range headers {
		vstr, _ := v.(string)
		req.Header.Add(k, vstr)
	}

	if method == "POST" {
		if _, ok := headers["Content-Type"]; !ok {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	resp, err = client.Do(req)

	return
}

// get http request header
func HttpHeader(method, url string, headers map[string]interface{}, payload []byte, timeout time.Duration) (header http.Header, err error) {
	resp, err := request(method, url, headers, payload, timeout)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	header = resp.Header

	return
}

// send http request
func HttpRequest(method, url string, headers map[string]interface{}, payload []byte, timeout time.Duration) (latency int64, statuscode int, response []byte, err error) {
	defer func(t time.Time) {
		latency = time.Since(t).Milliseconds()
	}(time.Now())

	resp, err := request(method, url, headers, payload, timeout)
	if err != nil {
		return 0, 0, nil, err
	}
	defer resp.Body.Close()

	response, err = ioutil.ReadAll(resp.Body)
	statuscode = resp.StatusCode

	return
}

// send http post request
func HttpPost(url string, headers map[string]interface{}, payload []byte, timeout int64) (int, []byte) {
	_, sc, resp, _ := HttpRequest("POST", url, headers, payload, time.Duration(timeout)*time.Second)

	return sc, resp
}

// send http get request
func HttpGet(url string, headers map[string]interface{}, timeout int64) (int, []byte) {
	_, sc, resp, _ := HttpRequest("GET", url, headers, nil, time.Duration(timeout)*time.Second)

	return sc, resp
}

// Http2Curl returns a CurlCommand corresponding to an http.Request
func Http2Curl(req *http.Request) (string, error) {
	var command []string
	if req == nil || req.URL == nil {
		return "", fmt.Errorf("getCurlCommand: invalid request, req or req.URL is nil")
	}

	command = append(command, "curl")

	schema := req.URL.Scheme
	requestURL := req.URL.String()
	if schema == "" {
		schema = "http"
		if req.TLS != nil {
			schema = "https"
		}
		requestURL = schema + "://" + req.Host + req.URL.Path
	}

	if schema == "https" {
		command = append(command, "-k")
	}

	command = append(command, "-X", bashEscape(req.Method))

	if req.Body != nil {
		var buff bytes.Buffer
		_, err := buff.ReadFrom(req.Body)
		if err != nil {
			return "", fmt.Errorf("getCurlCommand: buffer read from body error: %w", err)
		}
		// reset body for potential re-reads
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buff.Bytes()))
		if len(buff.String()) > 0 {
			bodyEscaped := bashEscape(buff.String())
			command = append(command, "-d", bodyEscaped)
		}
	}

	var keys []string

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		command = append(command, "-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))))
	}

	command = append(command, bashEscape(requestURL))

	command = append(command, "--compressed")

	return strings.Join(command, " "), nil
}

func TlsCheck(domain string) (string, string, string, string, float64, error) {
	info := strings.Split(domain, ":")
	domain = info[0]
	port := "443"
	if len(info) > 1 {
		port = info[1]
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", domain, port), &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10})
	if err != nil {
		return "", "", "", "", 0, err
		// return "", "", "", "", 0, fmt.Errorf("not support ssl certificate: %s", err.Error())
	}

	// err = conn.VerifyHostname(domain)
	// if err != nil {
	// 	return "", "", "", "", 0, fmt.Errorf("hostname doesn't match with certificate: %s", err.Error())
	// }

	for _, cert := range conn.ConnectionState().PeerCertificates {
		// 检测服务器证书是否已经过期(CA证书过期时间会比服务器证书长)
		if !cert.IsCA {
			sn := fmt.Sprintf("%x", cert.SerialNumber)
			issuer := cert.Issuer.CommonName
			dnss := strings.Join(cert.DNSNames, ",")
			expire := cert.NotAfter.Local().Format("2006-01-02 15:04:05")
			remain := cert.NotAfter.Sub(time.Now()).Hours()
			//version := conn.ConnectionState().Version
			return sn, issuer, dnss, expire, remain, cert.VerifyHostname(domain)
		}
	}

	return "", "", "", "", 0, nil

}
