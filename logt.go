package logt

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Same as flags in log package, except for LCaller
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LCaller                       // package/file-name.go:file-line func-name()
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

type Output interface {
	// Printf format can be empty string "" which means no formatting
	Printf(format string, vset ...interface{})
	Flags() (flag int)
	Prefix() (prefix string)
	SetFlags(flag int)
	SetPrefix(prefix string)
}

// Logger .
type Logger struct {
	out Output
}

// New creates a new *Logger
func New(out Output, prefix string, flag int) *Logger {
	out.SetFlags(flag)
	out.SetPrefix(prefix)
	return &Logger{
		out: out,
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	l.out.Printf("", v...)
	os.Exit(1)
}
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.out.Printf(format, v...)
	os.Exit(1)
}
func (l *Logger) Fatalln(v ...interface{}) {
	v = append(v, "\n")
	l.out.Printf("", v...)
	os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
	l.out.Printf("", v...)
	panic(fmt.Sprint(v...))
}
func (l *Logger) Panicf(format string, v ...interface{}) {
	l.out.Printf(format, v...)
	panic(fmt.Sprintf(format, v...))
}
func (l *Logger) Panicln(v ...interface{}) {
	v = append(v, "\n")
	l.out.Printf("", v...)
	panic(fmt.Sprint(v...))
}

func (l *Logger) Print(v ...interface{}) {
	l.out.Printf("", v...)
}
func (l *Logger) Printf(format string, v ...interface{}) {
	l.out.Printf(format, v...)
}
func (l *Logger) Println(v ...interface{}) {
	v = append(v, "\n")
	l.out.Printf("", v...)
}

func (l *Logger) SetFlags(flag int)       { l.out.SetFlags(flag) }
func (l *Logger) SetOutput(out Output)    { l.out = out }
func (l *Logger) SetPrefix(prefix string) { l.out.SetPrefix(prefix) }

func (l *Logger) Prefix() string { return l.out.Prefix() }
func (l *Logger) Flags() int     { return l.out.Flags() }

// func (l *Logger) Output(calldepth int, s string) error

var std = New(NewStdLogget(), "", LCaller|Ldate|Ltime)

func Fatal(v ...interface{})                 { std.Fatal(v...) }
func Fatalf(format string, v ...interface{}) { std.Fatalf(format, v...) }
func Fatalln(v ...interface{})               { std.Fatalln(v...) }
func Panic(v ...interface{})                 { std.Panic(v...) }
func Panicf(format string, v ...interface{}) { std.Panicf(format, v...) }
func Panicln(v ...interface{})               { std.Panicln(v...) }
func Print(v ...interface{})                 { std.Print(v...) }
func Printf(format string, v ...interface{}) { std.Printf(format, v...) }
func Println(v ...interface{})               { std.Println(v...) }

func Flags() int              { return std.Flags() }
func Prefix() string          { return std.Prefix() }
func SetFlags(flag int)       { std.SetFlags(flag) }
func SetPrefix(prefix string) { std.SetPrefix(prefix) }

// func SetOutput(w io.Writer)
// func Output(calldepth int, s string) error

//-----------------------------------------------------------------------------
// std output

var (
	errNotAvailable = errors.New("N/A")
)

func here(skip ...int) (funcName, fileName string, fileLine int, callerErr error) {
	sk := 1
	if len(skip) > 0 && skip[0] > 1 {
		sk = skip[0]
	}
	var pc uintptr
	var ok bool
	pc, fileName, fileLine, ok = runtime.Caller(sk)
	if !ok {
		callerErr = errNotAvailable
		return
	}
	fn := runtime.FuncForPC(pc)
	name := fn.Name()
	ix := strings.LastIndex(name, ".")
	if ix > 0 && (ix+1) < len(name) {
		name = name[ix+1:]
	}
	funcName = name
	nd, nf := filepath.Split(fileName)
	fileName = filepath.Join(filepath.Base(nd), nf)
	return
}

// StdLogget do not call SetFlags or SetPrefix concurrently
type StdLogget struct {
	prefix string
	flag   int
}

func NewStdLogget() *StdLogget {
	return &StdLogget{}
}

func (sl *StdLogget) Flags() (flag int)       { return sl.flag }
func (sl *StdLogget) Prefix() (prefix string) { return sl.prefix }
func (sl *StdLogget) SetFlags(flag int)       { sl.flag = flag }
func (sl *StdLogget) SetPrefix(prefix string) { sl.prefix = prefix }

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func getBuffer() *bytes.Buffer {
	buff := bufferPool.Get().(*bytes.Buffer)
	buff.Reset()
	return buff
}

func putBuffer(buff *bytes.Buffer) {
	bufferPool.Put(buff)
}

func anyErr(v ...interface{}) bool {
	for _, av := range v {
		_, ok := av.(error)
		if ok {
			return true
		}
	}
	return false
}

func (sl *StdLogget) Printf(format string, vset ...interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)
	defer func() {
		var ln string
		if format == "" {
			ln = fmt.Sprint(vset...)
		} else {
			ln = fmt.Sprintf(format, vset...)
		}
		fmt.Printf("%s %s", buf.Bytes(), ln)
	}()

	w := bufio.NewWriter(buf)
	defer w.Flush()

	if anyErr(vset...) {
		color.New(color.FgHiRed).Fprintf(w, "[error]")
	} else {
		color.New(color.FgHiBlue).Fprintf(w, "[info ]")
	}

	if sl.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		dtFormat := ""
		dt := time.Now()
		if sl.flag&LUTC != 0 {
			dt = dt.UTC()
		}
		if sl.flag&Ldate != 0 {
			dtFormat += "2006/01/02 "
		}
		if sl.flag&Ltime != 0 {
			dtFormat += "15:04:05"
		} else if sl.flag&Lmicroseconds != 0 {
			dtFormat += "15:04:05.000000"
		}
		fmt.Fprintf(w, " %v", dt.Format(strings.TrimSpace(dtFormat)))
	}

	if sl.flag&(LCaller|Lshortfile|Llongfile) != 0 {
		funcName, fileName, fileLine, fileErr := here(4)
		if fileErr == nil {
			if sl.flag&LCaller != 0 {
				fmt.Fprintf(w, " %s:%02d %s()", fileName, fileLine, funcName)
			} else if sl.flag&Lshortfile != 0 {
				short := fileName
				for i := len(fileName) - 1; i > 0; i-- {
					if fileName[i] == '/' {
						short = fileName[i+1:]
						break
					}
				}
				fileName = " " + short
				fmt.Fprintf(w, fileName)
			}
		}
	}
}
