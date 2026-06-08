package logging

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	now     func() time.Time
	info    *color.Color
	step    *color.Color
	success *color.Color
	warn    *color.Color
	err     *color.Color
	banner  *color.Color
}

func New(out io.Writer) *Logger {
	return &Logger{
		out:     out,
		now:     time.Now,
		info:    color.New(color.FgCyan),
		step:    color.New(color.FgBlue, color.Bold),
		success: color.New(color.FgGreen, color.Bold),
		warn:    color.New(color.FgYellow, color.Bold),
		err:     color.New(color.FgRed, color.Bold),
		banner:  color.New(color.FgHiWhite, color.BgBlue, color.Bold),
	}
}

func (l *Logger) Banner(title string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintf(l.out, "\n%s\n", l.banner.Sprintf(" %s ", title))
}

func (l *Logger) Step(format string, args ...any) {
	l.print("STEP", l.step, format, args...)
}

func (l *Logger) Info(format string, args ...any) {
	l.print("INFO", l.info, format, args...)
}

func (l *Logger) Success(format string, args ...any) {
	l.print(" OK ", l.success, format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.print("WARN", l.warn, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.print("ERR ", l.err, format, args...)
}

func (l *Logger) print(label string, style *color.Color, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	ts := l.now().Format("15:04:05")
	_, _ = fmt.Fprintf(l.out, "%s %s %s\n", color.New(color.FgHiBlack).Sprintf("[%s]", ts), style.Sprintf("[%s]", label), fmt.Sprintf(format, args...))
}
