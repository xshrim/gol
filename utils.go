package gol

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/xshrim/gol/colors"
)

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

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func toLower(src string) string {
	var dst []rune
	for _, v := range src {
		if v >= 65 && v <= 90 {
			v += 32
		}
		dst = append(dst, v)
	}
	return string(dst)
}

func toUpper(src string) string {
	var dst []rune
	for _, v := range src {
		if v >= 97 && v <= 122 {
			v -= 32
		}
		dst = append(dst, v)
	}
	return string(dst)
}

func isFormatString(s string) bool {
	for idx, c := range s {
		if c == '%' && idx != len(s)-1 {
			if idx == 0 || (idx > 0 && s[idx-1] != '\\') {
				return true
			}
		}
	}
	return false
}

func replaceDoubleQuote(buf *[]byte, s string) {
	last := false
	for _, c := range []byte(s) {
		if c != '"' {
			if c == '\\' {
				last = true
			} else {
				last = false
			}
		} else {
			if !last {
				*buf = append(*buf, '\\')
			}
			last = false
		}
		*buf = append(*buf, c)
	}
}

func replaceEscapePeriod(s string, flag bool) string {
	var buf []rune
	for _, c := range s {
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

func stringContainRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func stringIndex(s, t string) int {
	if len(t) == 0 {
		return 0
	}
	if len(s) < len(t) {
		return -1
	}
	for i := 0; i <= len(s)-len(t); i++ {
		if string(s[i:i+len(t)]) == t {
			return i
		}
	}
	return -1
}

func stringContainStr(s, t string) bool {
	return stringIndex(s, t) >= 0
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

func stringPrefixStr(s, t string) bool {
	return stringIndex(s, t) == 0
}

func stringSuffixStr(s, t string) bool {
	if len(t) == 0 {
		return true
	}
	if len(s) < len(t) {
		return false
	}
	return string(s[len(s)-len(t):]) == t
}

func stringSplit(s string, r rune) []string {
	var strs []string
	var runes []rune
	for i, c := range s {
		if c != r {
			runes = append(runes, c)
			if i == len(s)-1 {
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

func mapi2maps(i interface{}) interface{} {
	// var body interface{}
	// _ = yaml.Unmarshal([]byte(yamlstr), &body)
	// body = mapi2maps(body)
	// jsb,  _:= json.Marshal(body);
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = mapi2maps(v)
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

func colorStatusCode(statusCode int) string {
	var buff []byte

	switch {
	case statusCode < 200:
		buff = append(buff, colors.WYellow...)
	case statusCode < 300:
		buff = append(buff, colors.WGreen...)
	case statusCode < 400:
		buff = append(buff, colors.WBlue...)
	case statusCode < 500:
		buff = append(buff, colors.WRed...)
	case statusCode < 600:
		buff = append(buff, colors.WPurple...)
	default:
		buff = append(buff, colors.WCyan...)
	}
	itoa(&buff, statusCode, 3)
	buff = append(buff, colors.ColorOff...)
	return string(buff)
}

func colorRequestMethod(mtd string) string {
	var buff []byte
	switch mtd {
	case "GET":
		buff = append(buff, colors.WGreen...)
	case "POST":
		buff = append(buff, colors.WBlue...)
	case "DELETE":
		buff = append(buff, colors.WRed...)
	case "PUT":
		buff = append(buff, colors.WPurple...)
	case "PATCH":
		buff = append(buff, colors.WYellow...)
	default:
		buff = append(buff, colors.WCyan...)
	}
	buff = append(buff, mtd...)
	for i := 5 - len(mtd); i > 0; i-- {
		buff = append(buff, ' ')
	}
	buff = append(buff, colors.ColorOff...)
	return string(buff)
}

func map2json(dst []byte, fds F) []byte {
	for k, v := range fds {
		// append key
		dst = appendKey(dst, k)

		// append value
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
	}
	return dst
}

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
	case F:
		dst = append(dst, '{')
		for k, v := range val {
			dst = appendKey(dst, k)
			dst = tojson(dst, v)
		}
		dst = append(dst, '}')
	case []F:
		dst = append(dst, '[')
		for _, f := range val {
			if dst[len(dst)-1] == '[' || dst[len(dst)-1] == ',' {
				dst = append(dst, '{')
			}
			for k, v := range f {
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

// convert data to json-like string
func Jsonify(v interface{}) string {
	return string(tojson(nil, v))
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
		d, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		err = json.Unmarshal(d, &m)
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

func parseLevel(level interface{}) int {
	var lv int
	switch res := level.(type) {
	case int:
		lv = res
		if lv < 0 {
			lv = -1
		} else if lv > ALL {
			lv = OFF
		}
	case string:
		str := toUpper(res)
		switch str {
		case "OFF", "0":
			lv = OFF
		case "PANIC", "1":
			lv = PANIC
		case "FATAL", "2":
			lv = PANIC
		case "ERROR", "3":
			lv = ERROR
		case "WARN", "4":
			lv = WARN
		case "NOTIC", "5":
			lv = NOTIC
		case "INFO", "6":
			lv = INFO
		case "DEBUG", "7":
			lv = DEBUG
		case "TRACE", "8":
			lv = TRACE
		case "ALL", "9":
			lv = ALL
		default:
			lv = -1
		}
	default:
		lv = INFO
	}
	return lv
}

func writeFile(logfile string, mode int, data []string) {
	if logfile == "" {
		fmt.Println("Can not get log file")
		return
	}
	file, err := os.OpenFile(logfile, mode, 0666)
	if err != nil {
		fmt.Println("Can not open log file: ", err)
		return
	}
	for _, data := range data {
		if _, err = file.WriteString(data); err != nil {
			fmt.Println("Write log file failed: ", err)
			return
		}
	}
	file.Close()
}

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
		if noEscapeTable[b] {
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
			dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
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
		if noEscapeTable[b] {
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
			dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
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
		if !noEscapeTable[str[i]] {
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
		if !noEscapeTable[bs[i]] {
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
		dst = append(dst, hex[v>>4], hex[v&0x0f])
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
	ma, _ := json.Marshal(a)
	mb, _ := json.Marshal(b)
	return bytes.Compare(ma, mb)
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

// traverse from start to end with step
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

func IterS(v ...int) []int {
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

func Reverse(list interface{}) []interface{} {
	val := reflect.ValueOf(list)
	len := val.Len()
	out := make([]interface{}, len)
	if val.Kind() == reflect.Slice {
		for i := 0; i < len; i++ {
			out[len-1-i] = val.Index(i).Interface()
		}
	}
	return out
}

func Index(list interface{}, v interface{}) []int {
	out := []int{}

	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if val.Index(i).Interface() == v {
				out = append(out, i)
			}
		}
	}

	return out
}

func Remove(list interface{}, v interface{}) []interface{} {
	out := []interface{}{}

	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if val.Index(i).Interface() != v {
				out = append(out, val.Index(i).Interface())
			}
		}
	}
	return out
}

func RemoveAt(list interface{}, idx ...int) []interface{} {
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
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if i < index || i >= index+offset {
				out = append(out, val.Index(i).Interface())
			}
		}
	}
	return out
}

func Filter(list interface{}, fn func(interface{}) bool) []interface{} {
	var out []interface{}
	val := reflect.ValueOf(list)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			v := val.Index(i).Interface()
			if fn(v) {
				out = append(out, v)
			}
		}
	}

	return out
}

func Strlen(str string) int {
	return len([]rune(str))
}

func Substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

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
	return fmt.Sprintf("%.2f%ciB", float64(b)/float64(div), "KMGTPEZY"[exp])
}

func ByteSize(str string) (uint64, error) {
	i := strings.IndexFunc(str, func(r rune) bool {
		return r != '.' && !unicode.IsDigit(r)
	})
	var multiplier float64 = 1
	var sizeSuffixes = "BKMGTPEZY"
	if i > 0 {
		suffix := str[i:]
		multiplier = 0
		for j := 0; j < len(sizeSuffixes); j++ {
			base := string(sizeSuffixes[j])
			// M, MB, or MiB are all valid.
			switch suffix {
			case base, base + "B", base + "iB":
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

func ReadFile(fpath string) <-chan string {
	f, err := os.Open(fpath)
	if err != nil {
		panic(fmt.Sprintf("read file %s fail: %s", fpath, err.Error()))
	}
	//defer f.Close()

	c := make(chan string)
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

func ReadFileAll(fpath string) []byte {
	f, err := os.Open(fpath)
	if err != nil {
		panic(fmt.Sprintf("open file %s fail: %s", fpath, err.Error()))
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(fmt.Sprintf("read file %s fail: %s", fpath, err.Error()))
	}

	return bytes
}

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

func WordFrequency(fpath string, order bool, analysis func(string) []string) [][2]interface{} {
	var wordFrequencyMap = make(map[string]int)

	for line := range ReadFile(fpath) {
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

	if order {
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

func RandomString(n int, src ...byte) string {
	rand.Seed(time.Now().UnixNano())
	if len(src) == 0 {
		src = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	idxBits := 6
	idxMask := 1<<idxBits - 1
	idxMax := 63 / idxBits
	b := make([]byte, n)

	for i, cache, remain := n-1, rand.Int63(), idxMax; i >= 0; {
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

func Ping(ip string) bool {
	type ICMP struct {
		Type        uint8
		Code        uint8
		Checksum    uint16
		Identifier  uint16
		SequenceNum uint16
	}

	icmp := ICMP{
		Type: 8,
	}

	recvBuf := make([]byte, 32)
	var buffer bytes.Buffer

	_ = binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = checkSum(buffer.Bytes())
	buffer.Reset()
	_ = binary.Write(&buffer, binary.BigEndian, icmp)

	Time, _ := time.ParseDuration("2s")
	conn, err := net.DialTimeout("ip4:icmp", ip, Time)
	if err != nil {
		return exec.Command("ping", ip, "-c", "2", "-i", "1", "-W", "3").Run() == nil
	}
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		return false
	}
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	num, err := conn.Read(recvBuf)
	if err != nil {
		return false
	}

	_ = conn.SetReadDeadline(time.Time{})

	return string(recvBuf[0:num]) != ""
}

func IPv4() []string {
	out := []string{"127.0.0.1"}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				out = append(out, ipnet.IP.String())
			}
		}
	}
	return out
}

func Hosts(cidr string) []string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); {
		ips = append(ips, ip.String())

		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	if len(ips) < 2 {
		return []string{}
	}
	return ips[1 : len(ips)-1]
}

func TimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func StampNow() int64 {
	return time.Now().Unix()
}

func Time2Stamp(t string) int64 {
	stamp, _ := time.ParseInLocation("2006-01-02 15:04:05", t, time.Local)
	return stamp.Unix()
}

func Stamp2Time(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

// Gzip compresses the given data
func Gzip(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Gunzip uncompresses the given data
func Gunzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func Request(url string, creds ...string) (int, []byte) {
	client := http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Del("Cookie")
	req.Header.Del("Authorization")
	if len(creds) > 1 {
		req.SetBasicAuth(creds[0], creds[1])
	} else if len(creds) == 1 {
		req.Header.Add("Authorization", "Bearer "+creds[0])
	}

	resp, err := client.Do(req)
	if err != nil {
		return 600, []byte(fmt.Sprintf("request %s failed: %s", url, err.Error()))
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 600, []byte(fmt.Sprintf("read response %s failed: %s", url, err.Error()))
	}

	return resp.StatusCode, bodyBytes
}
