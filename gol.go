package gol

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xshrim/gol/color"
	"github.com/xshrim/gol/tk"
)

type LogHook func(int, *[]byte) bool

// convert map[string]interface{} or F object to string
// func (f M) Jsonify() string {
// 	return Jsonify(f)
// }

// get value of the path key from the map[string]interface{} or F object
// func (f M) Jsquery(keyPath string) interface{} {
// 	return Jsquery(Jsonify(f), keyPath)
// }

// log outpu interface
type Printer interface {
	Error(...interface{}) Printer
	Errorf(string, ...interface{}) Printer
	Warn(...interface{}) Printer
	Warnf(string, ...interface{}) Printer
	Notic(...interface{}) Printer
	Noticf(string, ...interface{}) Printer
	Info(...interface{}) Printer
	Infof(string, ...interface{}) Printer
	Debug(...interface{}) Printer
	Debugf(string, ...interface{}) Printer
	Trace(...interface{}) Printer
	Tracef(string, ...interface{}) Printer
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	Log(interface{}, ...interface{}) Printer
	Logf(interface{}, string, ...interface{}) Printer
	Flush()
}

// key names for json-format log
type KeyName struct {
	prefixKey string
	dateKey   string
	timeKey   string
	stackKey  string
	fileKey   string
	levelKey  string
	ctxKey    string
	msgKey    string
}

// logger
type Logger struct {
	mu        sync.Mutex  // ensures atomic writes; protects the following fields
	prefix    string      // prefix to write at beginning of each line
	watchfile string      // the file to watch for reload configurations dynamically
	level     int         // log level(ERROR, WARN, INFO, DEBUG, TRACE)
	dtfmt     string      // date time formatter(2006-01-02 15:04:05.000)
	flag      int         // properties
	keys      *KeyName    // filed keys
	writer    io.Writer   // destination for output synchronously
	saver     LogSaver    // write log to file asynchronously
	hook      LogHook     // hook function
	buf       []byte      // for accumulating text to write
	bufchan   chan string // for logsaver to read from
	done      chan bool   // for savelog goroutine to exit
}

var levels = []string{
	"OFF",
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"NOTIC",
	"INFO",
	"DEBUG",
	"TRACE",
	"ALL",
}

var flags = []string{
	"",
	"Ldate",
	"Ltime",
	"Ldefault",
	"Lmsec",
	"Lstack",
	"Lnolvl",
	"Lnobrkt",
	"Lfile",
	"Llfile",
	"Ljson",
	"Lcolor",
	"Lfcolor",
	"Lutc",
}

func index(slice []string, str string) int {
	for idx, val := range slice {
		if val == str {
			return idx
		}
	}
	return -1
}

// new logger with default configurations
func New() *Logger {
	level := INFO
	flag := Ldefault

	var prefix, watchfile string

	if val, ok := os.LookupEnv("GOL_PREFIX"); ok {
		prefix = val
	}
	if val, ok := os.LookupEnv("GOL_WATCHFILE"); ok {
		watchfile = val
	}
	if val, ok := os.LookupEnv("GOL_LEVEL"); ok && val != "" {
		if tmp := index(levels, strings.ToUpper(val)); tmp != -1 {
			level = tmp
		}
	}
	if val, ok := os.LookupEnv("GOL_FLAG"); ok && val != "" {
		if tmp := index(flags, strings.ToUpper(val[:1])+strings.ToLower(val[1:])); tmp != -1 {
			flag = tmp
		}
	}

	return NewLogger(os.Stderr, prefix, watchfile, level, flag)
}

// create logger with customized configurations
func NewLogger(writer io.Writer, prefix, watchfile string, level, flag int) *Logger {
	if level < OFF {
		level = OFF
	} else if level > ALL {
		level = ALL
	}

	if writer == nil {
		writer = ioutil.Discard
	}

	return &Logger{
		writer:    writer,
		prefix:    prefix,
		level:     level,
		flag:      flag,
		watchfile: watchfile,
		keys: &KeyName{
			prefixKey: "\"prefix\": \"",
			dateKey:   "\"date\": \"",
			timeKey:   "\"time\": \"",
			stackKey:  "\"stack\": \"",
			fileKey:   "\"file\": \"",
			levelKey:  "\"level\": \"",
			ctxKey:    "\"ctx\": ",
			msgKey:    "\"msg\": ",
		}}
}

// get writer config of default logger
func GetWriter() io.Writer {
	return std.GetWriter()
}

// get saver config of default logger
func GetSaver() LogSaver {
	return std.GetSaver()
}

// get prefix config of default logger
func GetPrefix() string {
	return std.GetPrefix()
}

// get level config of default logger
func GetLevel() int {
	return std.GetLevel()
}

// get time formatter config of default logger
func GetTimeFormatter() string {
	return std.GetTimeFormatter()
}

// get flag config of default logger
func GetFlag() int {
	return std.GetFlag()
}

// set hot reload for default logger
func HotReload(watchfile ...string) *Logger {
	return std.HotReload(watchfile...)
}

// ensure all logs are written to file
func Flush() {
	std.Flush()
}

// create a context with fields using default logger
func With(ctx map[string]interface{}) *Context {
	return std.With(ctx)
}

// create a thread-safe context with fields using default logger
func WithSafe(ctx map[string]interface{}) *SafeContext {
	return std.WithSafe(ctx)
}

// set the output destinations for default logger
func Writer(w ...io.Writer) *Logger {
	return std.Writer(w...)
}

// clean the output destinations for default logger
func UnWriter() *Logger {
	return std.UnWriter()
}

// set log hook for default logger
func Hook(hook LogHook) *Logger {
	return std.Hook(hook)
}

// set log saver for default logger
func Saver(ls LogSaver) *Logger {
	return std.Saver(ls)
}

// set log prefix for default logger
func Prefix(p string) *Logger {
	return std.Prefix(p)
}

// set log level for default logger
func Level(v int) *Logger {
	return std.Level(v)
}

// set time formatter (2006-01-02 15:04:05.000) for default logger
func TimeFormatter(f string) *Logger {
	return std.TimeFormatter(f)
}

// set log flags for default logger
func Flag(flag int, mode ...int) *Logger {
	return std.Flag(flag, mode...)
}

// add flag for default logger
func AddFlag(flag int) *Logger {
	return std.AddFlag(flag)
}

// delete flag for default logger
func DelFlag(flag int) *Logger {
	return std.DelFlag(flag)
}

func HasFlag(flag int) bool {
	return std.HasFlag(flag)
}

// set prefix key name of json log for default logger
func PrefixKey(key string) *Logger {
	return std.PrefixKey(key)
}

// set date key name of json log for default logger
func DateKey(key string) *Logger {
	return std.DateKey(key)
}

// set time key name of json log for default logger
func TimeKey(key string) *Logger {
	return std.TimeKey(key)
}

// set stack key name of json log for default logger
func StackKey(key string) *Logger {
	return std.StackKey(key)
}

// set file key name of json log for default logger
func FileKey(key string) *Logger {
	return std.FileKey(key)
}

// set level key name of json log for default logger
func LevelKey(key string) *Logger {
	return std.LevelKey(key)
}

// set context key name of json log for default logger
func CtxKey(key string) *Logger {
	return std.CtxKey(key)
}

// set message key name of json log for default logger
func MsgKey(key string) *Logger {
	return std.MsgKey(key)
}

// output error log using default logger
func Error(v ...interface{}) Printer {
	if std.lvcheck(ERROR) {
		_ = std.Output(ERROR, 2, nil, fmt.Sprint(v...), true)
	}
	return std
}

// output format error log using default logger
func Errorf(format string, v ...interface{}) Printer {
	if std.lvcheck(ERROR) {
		_ = std.Output(ERROR, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return std
}

// output warn log using default logger
func Warn(v ...interface{}) Printer {
	if std.lvcheck(WARN) {
		_ = std.Output(WARN, 2, nil, fmt.Sprint(v...), true)
	}
	return std
}

// output format warn log using default logger
func Warnf(format string, v ...interface{}) Printer {
	if std.lvcheck(WARN) {
		_ = std.Output(WARN, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return std
}

// output notice log using default logger
func Notic(v ...interface{}) Printer {
	if std.lvcheck(NOTIC) {
		_ = std.Output(NOTIC, 2, nil, fmt.Sprint(v...), true)
	}
	return std
}

// output format notice log using default logger
func Noticf(format string, v ...interface{}) Printer {
	if std.lvcheck(NOTIC) {
		_ = std.Output(NOTIC, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return std
}

// output info log using default logger
func Info(v ...interface{}) Printer {
	if std.lvcheck(INFO) {
		_ = std.Output(INFO, 2, nil, fmt.Sprint(v...), true)
	}
	return std
}

// output format info log using default logger
func Infof(format string, v ...interface{}) Printer {
	if std.lvcheck(INFO) {
		_ = std.Output(INFO, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return std
}

// output debug log using default logger
func Debug(v ...interface{}) Printer {
	if std.lvcheck(DEBUG) {
		_ = std.Output(DEBUG, 2, nil, fmt.Sprint(v...), true)
	}
	return std
}

// output format debug log using default logger
func Debugf(format string, v ...interface{}) Printer {
	if std.lvcheck(DEBUG) {
		_ = std.Output(DEBUG, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return std
}

// output trace log using default logger
func Trace(v ...interface{}) Printer {
	if std.lvcheck(TRACE) {
		_ = std.Output(TRACE, 2, nil, fmt.Sprint(v...), true)
	}
	return std
}

// output format trace log using default logger
func Tracef(format string, v ...interface{}) Printer {
	if std.lvcheck(TRACE) {
		_ = std.Output(TRACE, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return std
}

// exit with code 1 after output fatal log using default logger
func Fatal(v ...interface{}) {
	if std.lvcheck(FATAL) {
		_ = std.Output(FATAL, 2, nil, fmt.Sprint(v...), true)
	}
	os.Exit(1)
}

// exit with code 1 after output format fatal log using default logger
func Fatalf(format string, v ...interface{}) {
	if std.lvcheck(FATAL) {
		_ = std.Output(FATAL, 2, nil, fmt.Sprintf(format, v...), true)
	}
	os.Exit(1)
}

// panic after output panic log using default logger
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	if std.lvcheck(PANIC) {
		_ = std.Output(PANIC, 2, nil, s, true)
	}
	panic(s)
}

// panic after output format panic log using default logger
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	if std.lvcheck(PANIC) {
		_ = std.Output(PANIC, 2, nil, s, true)
	}
	panic(s)
}

// output log without newline using default logger
func Log(level interface{}, v ...interface{}) Printer {
	s := fmt.Sprint(v...)
	lv := parseLevel(level)
	if std.lvcheck(lv) {
		_ = std.Output(lv, 2, nil, s, false)
	}
	switch lv {
	case FATAL:
		os.Exit(1)
	case PANIC:
		panic(s)
	}
	return std
}

// output format log without newline using default logger
func Logf(level interface{}, format string, v ...interface{}) Printer {
	s := fmt.Sprintf(format, v...)
	lv := parseLevel(level)
	if std.lvcheck(lv) {
		_ = std.Output(lv, 2, nil, s, false)
	}
	switch lv {
	case FATAL:
		os.Exit(1)
	case PANIC:
		panic(s)
	}
	return std
}

// get writer config
func (l *Logger) GetWriter() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.writer
}

// get saver config
func (l *Logger) GetSaver() LogSaver {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.saver
}

// get prefix config
func (l *Logger) GetPrefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// get level config
func (l *Logger) GetLevel() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// get time formatter config
func (l *Logger) GetTimeFormatter() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.dtfmt
}

// get flag config
func (l *Logger) GetFlag() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

// set hot reload
func (l *Logger) HotReload(watchfile ...string) *Logger {
	if l.watchfile == "" && len(watchfile) > 0 { // hot reload can only be set once
		l.watchfile = watchfile[0]
	}
	go l.notisfy()
	return l
}

// ensure all logs are written to file
func (l *Logger) Flush() {
	if l.done != nil {
		l.done <- true
	}
}

// create a context with fields
func (l *Logger) With(fd map[string]interface{}) *Context {
	// 	if len(fd) == 0 {
	// 		return &Context{loggers: []*Logger{l}, buf: nil}
	// 	} else {
	return &Context{loggers: []*Logger{l}, buf: map2json(nil, fd)}
	// 	}
}

// create a thread-safe context with fields
func (l *Logger) WithSafe(fd map[string]interface{}) *SafeContext {
	// if len(fd) == 0 {
	// 	return &SafeContext{loggers: []*Logger{l}, buf: nil}
	// } else {
	return &SafeContext{loggers: []*Logger{l}, buf: map2json(nil, fd)}
	// }
}

// sets the output destinations
func (l *Logger) Writer(writer ...io.Writer) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(writer) == 0 {
		l.writer = nil
	} else {
		l.writer = io.MultiWriter(writer...)
	}
	return l
}

// clean the output destinations
func (l *Logger) UnWriter() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = nil
	return l
}

// set log hook
func (l *Logger) Hook(hook LogHook) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hook = hook
	return l
}

// set log saver
func (l *Logger) Saver(ls LogSaver) *Logger {
	if l.saver != nil || ls == nil { // log saver can only be set once
		return l
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.saver = ls
	l.done = make(chan bool)
	l.bufchan = make(chan string, 3000)
	go l.saveLog()
	return l
}

// set log prefix
func (l *Logger) Prefix(prefix string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
	return l
}

// set log level
func (l *Logger) Level(level int) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	return l
}

// set time formatter (2006-01-02 15:04:05.000)
func (l *Logger) TimeFormatter(f string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.dtfmt = f
	return l
}

// set log flags
func (l *Logger) Flag(flag int, mode ...int) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(mode) == 0 || mode[0] == OVERRIDE {
		l.flag = flag
	} else if mode[0] == APPEND {
		l.flag = l.flag | flag
	} else {
		if l.flag&flag != 0 {
			l.flag = l.flag ^ flag
		}
	}
	return l
}

// add flag for the logger
func (l *Logger) AddFlag(flag int) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = l.flag | flag
	return l
}

// delete this flag for the logger
func (l *Logger) DelFlag(flag int) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&flag != 0 {
		l.flag = l.flag ^ flag
	}
	return l
}

// if the logger has this flag
func (l *Logger) HasFlag(flag int) bool {
	return l.flag&flag != 0
}

// set prefix key name of json log
func (l *Logger) PrefixKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "prefix"
	}
	l.keys.prefixKey = "\"" + key + "\": \""
	return l
}

// set date key name of json log
func (l *Logger) DateKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "date"
	}
	l.keys.dateKey = "\"" + key + "\": \""
	return l
}

// set time key name of json log
func (l *Logger) TimeKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "time"
	}
	l.keys.timeKey = "\"" + key + "\": \""
	return l
}

// set stack key name of json log
func (l *Logger) StackKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "stack"
	}
	l.keys.stackKey = "\"" + key + "\": \""
	return l
}

// set file key name of json log
func (l *Logger) FileKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "file"
	}
	l.keys.fileKey = "\"" + key + "\": \""
	return l
}

// set level key name of json log
func (l *Logger) LevelKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "level"
	}
	l.keys.levelKey = "\"" + key + "\": \""
	return l
}

// set context key name of json log
func (l *Logger) CtxKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "ctx"
	}
	l.keys.ctxKey = "\"" + key + "\": "
	return l
}

// set message key name of json log
func (l *Logger) MsgKey(key string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if key == "" {
		key = "msg"
	}
	l.keys.msgKey = "\"" + key + "\": "
	return l
}

// output error log
func (l *Logger) Error(v ...interface{}) Printer {
	if l.lvcheck(ERROR) {
		_ = l.Output(ERROR, 2, nil, fmt.Sprint(v...), true)
	}
	return l
}

// output format error log
func (l *Logger) Errorf(format string, v ...interface{}) Printer {
	if l.lvcheck(ERROR) {
		_ = l.Output(ERROR, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return l
}

// output warn log
func (l *Logger) Warn(v ...interface{}) Printer {
	if l.lvcheck(WARN) {
		_ = l.Output(WARN, 2, nil, fmt.Sprint(v...), true)
	}
	return l
}

// output format warn log
func (l *Logger) Warnf(format string, v ...interface{}) Printer {
	if l.lvcheck(WARN) {
		_ = l.Output(WARN, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return l
}

// output notice log
func (l *Logger) Notic(v ...interface{}) Printer {
	if l.lvcheck(NOTIC) {
		_ = l.Output(NOTIC, 2, nil, fmt.Sprint(v...), true)
	}
	return l
}

// output format notic log
func (l *Logger) Noticf(format string, v ...interface{}) Printer {
	if l.lvcheck(NOTIC) {
		_ = l.Output(NOTIC, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return l
}

// output info log
func (l *Logger) Info(v ...interface{}) Printer {
	if l.lvcheck(INFO) {
		_ = l.Output(INFO, 2, nil, fmt.Sprint(v...), true)
	}
	return l
}

// output format info log
func (l *Logger) Infof(format string, v ...interface{}) Printer {
	if l.lvcheck(INFO) {
		_ = l.Output(INFO, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return l
}

// output debug log
func (l *Logger) Debug(v ...interface{}) Printer {
	if l.lvcheck(DEBUG) {
		_ = l.Output(DEBUG, 2, nil, fmt.Sprint(v...), true)
	}
	return l
}

// output format debug log
func (l *Logger) Debugf(format string, v ...interface{}) Printer {
	if l.lvcheck(DEBUG) {
		_ = l.Output(DEBUG, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return l
}

// output trace log
func (l *Logger) Trace(v ...interface{}) Printer {
	if l.lvcheck(TRACE) {
		_ = l.Output(TRACE, 2, nil, fmt.Sprint(v...), true)
	}
	return l
}

// output format trace log
func (l *Logger) Tracef(format string, v ...interface{}) Printer {
	if l.lvcheck(TRACE) {
		_ = l.Output(TRACE, 2, nil, fmt.Sprintf(format, v...), true)
	}
	return l
}

// exit with code 1 after output fatal log
func (l *Logger) Fatal(v ...interface{}) {
	if l.lvcheck(FATAL) {
		_ = l.Output(FATAL, 2, nil, fmt.Sprint(v...), true)
	}
	os.Exit(1)
}

// exit with code 1 after output format fatal log
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.lvcheck(FATAL) {
		_ = l.Output(FATAL, 2, nil, fmt.Sprintf(format, v...), true)
	}
	os.Exit(1)
}

// panic after output panic log
func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	if l.lvcheck(PANIC) {
		_ = l.Output(PANIC, 2, nil, s, true)
	}
	panic(s)
}

// panic after output format panic log
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	if l.lvcheck(PANIC) {
		_ = l.Output(PANIC, 2, nil, s, true)
	}
	panic(s)
}

// output log without newline
func (l *Logger) Log(level interface{}, v ...interface{}) Printer {
	s := fmt.Sprint(v...)
	lv := parseLevel(level)
	if l.lvcheck(lv) {
		_ = l.Output(lv, 2, nil, s, false)
	}
	switch lv {
	case FATAL:
		os.Exit(1)
	case PANIC:
		panic(s)
	}
	return l
}

// output format log without newline
func (l *Logger) Logf(level interface{}, format string, v ...interface{}) Printer {
	s := fmt.Sprintf(format, v...)
	lv := parseLevel(level)
	if l.lvcheck(lv) {
		_ = l.Output(lv, 2, nil, s, false)
	}
	switch lv {
	case FATAL:
		os.Exit(1)
	case PANIC:
		panic(s)
	}
	return l
}

// func (l *Logger) appendPrefix(buf *[]byte, str ...string) {
// 	if l.flag&Ljson != 0 {
// 		*buf = append(*buf, "\""+str[0]+"\": \""...)
// 	} else if l.flag&Lcolor != 0 || l.flag&Lfullcolor != 0 {
// 		if str[0] == "level" {
// 			switch str[1] {
// 			case "ERROR":
// 				*buf = append(*buf, "\033[1;31m"...)
// 			case "WARN":
// 				*buf = append(*buf, "\033[1;33m"...)
// 			case "INFO":
// 				*buf = append(*buf, "\033[1;32m"...)
// 			case "DEBUG":
// 				*buf = append(*buf, "\033[1;35m"...)
// 			case "TRACE":
// 				*buf = append(*buf, "\033[1;36m"...)
// 			case "FATAL":
// 				*buf = append(*buf, "\033[1;34m"...)
// 			case "PANIC":
// 				*buf = append(*buf, "\033[1;37m"...)
// 			}
// 		} else if l.flag&Lfullcolor != 0 {
// 			switch str[0] {
// 			case "prefix":
// 				*buf = append(*buf, "\033[0;31m"...)
// 			case "date", "time":
// 				*buf = append(*buf, "\033[0;33m"...)
// 			case "stack":
// 				*buf = append(*buf, "\033[0;35m"...)
// 			case "file":
// 				*buf = append(*buf, "\033[0;36m"...)
// 			case "ctx":
// 				*buf = append(*buf, "\033[0;34m"...)
// 			case "msg":
// 				*buf = append(*buf, "\033[0;32m"...)
// 			}
// 		}
// 	}
// }

// func (l *Logger) appendSuffix(buf *[]byte, str ...string) {
// 	if l.flag&Ljson != 0 {
// 		*buf = append(*buf, "\","...)
// 	} else if l.flag&Lfullcolor != 0 || (len(str) > 0 && str[0] == "level" && l.flag&Lcolor != 0) {
// 		*buf = append(*buf, "\033[0m"...)
// 	}
// 	*buf = append(*buf, ' ')
// }

// format log header data
func (l *Logger) formatHeader(buf *[]byte, t time.Time, fn, file string, fd []byte, lv, line int) {
	if l.prefix != "" {
		if l.flag&Ljson != 0 {
			*buf = append(*buf, l.keys.prefixKey...)
		} else if l.flag&Lfcolor != 0 {
			*buf = append(*buf, color.Red...)
		}
		*buf = append(*buf, l.prefix...)
		if l.flag&Ljson != 0 {
			*buf = append(*buf, "\","...)
		} else if l.flag&Lfcolor != 0 {
			*buf = append(*buf, color.ColorOff...)
		}
		*buf = append(*buf, ' ')
	}

	if l.flag&(Ldate|Ltime|Lmsec) != 0 {
		if l.flag&Lutc != 0 {
			t = t.UTC()
		}

		if l.dtfmt != "" {
			dt := t.Format(l.dtfmt)
			if l.flag&Ljson != 0 {
				*buf = append(*buf, l.keys.timeKey...)
			} else if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.Yellow...)
			}

			*buf = append(*buf, dt...)

			if l.flag&Ljson != 0 {
				*buf = append(*buf, "\","...)
			} else if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.ColorOff...)
			}
			*buf = append(*buf, ' ')
		} else {
			if l.flag&Ldate != 0 {
				if l.flag&Ljson != 0 {
					*buf = append(*buf, l.keys.dateKey...)
				} else if l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.Yellow...)
				}

				year, month, day := t.Date()
				itoa(buf, year, 4)
				*buf = append(*buf, '/')
				itoa(buf, int(month), 2)
				*buf = append(*buf, '/')
				itoa(buf, day, 2)

				if l.flag&Ljson != 0 {
					*buf = append(*buf, "\","...)
				} else if l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.ColorOff...)
				}
				*buf = append(*buf, ' ')
			}

			if l.flag&(Ltime|Lmsec) != 0 {
				if l.flag&Ljson != 0 {
					*buf = append(*buf, l.keys.timeKey...)
				} else if l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.Yellow...)
				}

				hour, min, sec := t.Clock()
				itoa(buf, hour, 2)
				*buf = append(*buf, ':')
				itoa(buf, min, 2)
				*buf = append(*buf, ':')
				itoa(buf, sec, 2)
				if l.flag&Lmsec != 0 {
					*buf = append(*buf, '.')
					itoa(buf, t.Nanosecond()/1e3, 6)
				}

				if l.flag&Ljson != 0 {
					*buf = append(*buf, "\","...)
				} else if l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.ColorOff...)
				}
				*buf = append(*buf, ' ')
			}
		}
	}

	if l.flag&Lnolvl == 0 {
		if l.flag&Ljson != 0 {
			*buf = append(*buf, l.keys.levelKey...)
		}
		switch lv {
		case PANIC:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BWhite...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "PANIC"...)
		case FATAL:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BBlue...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "FATAL"...)
		case ERROR:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BRed...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "ERROR"...)
		case WARN:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BYellow...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "WARN"...)
		case NOTIC:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BYellow...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "NOTIC"...)
		case INFO:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BGreen...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "INFO"...)
		case DEBUG:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BPurple...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "DEBUG"...)
		case TRACE:
			if l.flag&Ljson == 0 {
				if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
					*buf = append(*buf, color.BCyan...)
				}
				if l.flag&Lnobrkt == 0 {
					*buf = append(*buf, '[')
				}
			}
			*buf = append(*buf, "TRACE"...)
		}
		if l.flag&Ljson != 0 {
			*buf = append(*buf, "\","...)
		} else {
			if l.flag&Lnobrkt == 0 {
				*buf = append(*buf, ']')
			}
			if l.flag&Lcolor != 0 || l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.ColorOff...)
			}
		}
		*buf = append(*buf, ' ')
	}

	if l.flag&Lstack != 0 {
		if l.flag&Ljson != 0 {
			*buf = append(*buf, l.keys.stackKey...)
		} else {
			if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.Purple...)
			}
			if l.flag&Lnobrkt == 0 {
				*buf = append(*buf, '<')
			}
		}
		*buf = append(*buf, fn...)
		if l.flag&Ljson != 0 {
			*buf = append(*buf, "\","...)
		} else {
			if l.flag&Lnobrkt == 0 {
				*buf = append(*buf, '>')
			}
			if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.ColorOff...)
			}
		}
		*buf = append(*buf, ' ')
	}

	if l.flag&(Lfile|Llfile) != 0 {
		if l.flag&Lfile != 0 {
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					file = file[i+1:]
					break
				}
			}
		}

		if l.flag&Ljson != 0 {
			*buf = append(*buf, l.keys.fileKey...)
		} else if l.flag&Lfcolor != 0 {
			*buf = append(*buf, color.Cyan...)
		}
		l.buf = append(l.buf, file...)
		l.buf = append(l.buf, ':')
		itoa(buf, line, -1)
		if l.flag&Ljson != 0 {
			*buf = append(*buf, "\","...)
		} else {
			// *buf = append(*buf, ':')
			if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.ColorOff...)
			}
		}
		*buf = append(*buf, ' ')
	}

	if len(fd) > 0 {
		if l.flag&Ljson != 0 {
			*buf = append(*buf, l.keys.ctxKey...)
			//replaceDoubleQuote(buf, fd)
		} else {
			if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.Blue...)
			}
		}
		*buf = append(*buf, fd...)
		// *buf = append(*buf, '}')
		if l.flag&Ljson != 0 {
			*buf = append(*buf, ',')
		} else {
			if l.flag&Lfcolor != 0 {
				*buf = append(*buf, color.ColorOff...)
			}
		}
		*buf = append(*buf, ' ')
	}
}

// output log data
func (l *Logger) Output(lv, calldepth int, fd []byte, s string, feed bool) error {
	now := time.Now()
	var pc uintptr
	var fn string
	var file string
	var line int
	var err error
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lfile|Llfile|Lstack) != 0 {
		// Release lock while getting caller info - it's expensive
		l.mu.Unlock()
		var ok bool
		pc, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
			fn = "???"
		} else {
			if l.flag&Lstack != 0 {
				fn = runtime.FuncForPC(pc).Name()
			}
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	if l.flag&Ljson != 0 {
		l.buf = append(l.buf, '{')
	}

	l.formatHeader(&l.buf, now, fn, file, fd, lv, line)

	if s != "" {
		if len(s) > 0 && s[len(s)-1] == '\n' {
			s = s[:len(s)-1]
		}
		if l.flag&Ljson != 0 {
			if (s[0] == '{' && s[len(s)-1] == '}') || (s[0] == '[' && s[len(s)-1] == ']') {
				line = -1 // reuse this variable
			}

			l.buf = append(l.buf, l.keys.msgKey...)
			if line == -1 {
				l.buf = append(l.buf, s...)
			} else {
				l.buf = append(l.buf, '"')
				replaceDoubleQuote(&l.buf, s)
			}
		} else {
			if l.flag&Lfcolor != 0 {
				l.buf = append(l.buf, color.Green...)
			}
			l.buf = append(l.buf, s...)
		}
		if l.flag&Ljson != 0 {
			if line != -1 {
				l.buf = append(l.buf, '"')
			}
		} else if l.flag&Lfcolor != 0 {
			l.buf = append(l.buf, color.ColorOff...)
		}
	}
	if l.flag&Ljson != 0 {
		l.buf = append(l.buf, '}')
	}

	if feed {
		l.buf = append(l.buf, '\n')
	}

	if l.hook != nil {
		if !l.hook(lv, &l.buf) {
			return nil
		}
	}
	if l.bufchan != nil {
		l.bufchan <- string(l.buf)
	}

	if l.writer != nil {
		_, err = l.writer.Write(l.buf)
	}

	return err
}

// watch the special file to reload log configuratins dynamically
func (l *Logger) notisfy() {
	var lastModifyTime int64
	defaultLevel := l.level
	defaultFlag := l.flag
	cpath := l.watchfile
	if cpath == "" {
		cpath = filepath.Join(tk.WorkPath(), ".gol")
		// cpath = "/tmp/.gol"
		// if runtime.GOOS == "windows" {
		// 	cpath = "C:\\.gol"
		// }
	}
	for {
		file, err := os.Open(cpath)
		if err == nil {
			fileInfo, err := file.Stat()
			if err == nil {
				curModifyTime := fileInfo.ModTime().Unix()
				if curModifyTime > lastModifyTime {
					lastModifyTime = curModifyTime
					l.level = defaultLevel
					l.flag = defaultFlag
					var line []byte
					for {
						b := make([]byte, 1)
						n, err := file.Read(b)
						if err != nil || err == io.EOF {
							break
						}
						if n > 0 {
							c := b[0]
							if c == '\r' || c == '\n' {
								break
							}
							line = append(line, c)
						}
					}
					if bytes.Contains(line, []byte("json")) {
						l.flag = l.flag | Ljson
					}
					if bytes.Contains(line, []byte("format")) {
						if l.flag&Ljson != 0 {
							l.flag = l.flag ^ Ljson
						}
					}
					if bytes.Contains(line, []byte("stack")) {
						l.flag = l.flag | Lstack | Lfile
					}
					if bytes.Contains(line, []byte("nostack")) {
						if l.flag&Lstack != 0 {
							l.flag = l.flag ^ Lstack
						}
						if l.flag&Lfile != 0 {
							l.flag = l.flag ^ Lfile
						}
					}
					if bytes.Contains(line, []byte("color")) {
						if l.flag&Lfcolor != 0 {
							l.flag = l.flag ^ Lfcolor
						}
						l.flag = l.flag | Lcolor
					}
					if bytes.Contains(line, []byte("fcolor")) {
						l.flag = l.flag | Lfcolor
					}
					if bytes.Contains(line, []byte("nocolor")) {
						if l.flag&Lcolor != 0 {
							l.flag = l.flag ^ Lcolor
						}
						if l.flag&Lfcolor != 0 {
							l.flag = l.flag ^ Lfcolor
						}
					}
					line = bytes.ReplaceAll(line, []byte("json"), []byte(""))
					line = bytes.ReplaceAll(line, []byte("format"), []byte(""))
					line = bytes.ReplaceAll(line, []byte("color"), []byte(""))
					line = bytes.ReplaceAll(line, []byte("fcolor"), []byte(""))
					line = bytes.ReplaceAll(line, []byte("nocolor"), []byte(""))
					line = bytes.ReplaceAll(line, []byte("stack"), []byte(""))
					line = bytes.ReplaceAll(line, []byte("nostack"), []byte(""))
					line = bytes.ReplaceAll(line, []byte(" "), []byte(""))
					lv := parseLevel(string(line))
					if lv >= 0 {
						l.level = lv
					} else {
						l.level = defaultLevel
					}
				}
			}
		} else {
			l.level = defaultLevel
			l.flag = defaultFlag
		}
		file.Close()

		time.Sleep(3 * time.Second)
	}
}

// verify whather the log should be output by compare the level config
func (l *Logger) lvcheck(lv int) bool {
	return l.level >= lv
}

// save log to disk
func (l *Logger) saveLog() {
	for {
		end := false
		select {
		case data := <-l.bufchan:
			l.saver.save([]byte(data))
		case <-l.done:
			end = true
			close(l.bufchan)
		}

		if end {
			break
		}
	}
}
