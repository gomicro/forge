package scribe

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Scribe struct {
	writer io.Writer
	level  int
	theme  *Theme
}

func NewScribe(writer io.Writer, theme *Theme) (Scriber, error) {
	err := ValidateTheme(theme)
	if err != nil {
		return nil, fmt.Errorf("new scribe: %w", err)
	}

	return &Scribe{
		writer: writer,
		theme:  theme,
	}, nil
}

func (s *Scribe) BeginDescribe(desc string) {
	s.println()
	s.printt(s.theme.Describe(desc))
	s.level++
}

func (s *Scribe) EndDescribe() {
	s.level--
}

func (s *Scribe) Print(str string) {
	s.level++
	s.printt(s.theme.Print(str))
	s.level--
}

func (s *Scribe) PrintLines(buf *bytes.Buffer) {
	scanner := bufio.NewScanner(buf)

	s.level++

	for scanner.Scan() {
		s.printt(s.theme.Print(scanner.Text()))
	}

	s.println()
	s.level--
}

func (s *Scribe) Error(err error) {
	s.level++
	s.printt(s.theme.Error(err))
	s.level--
}

func (s *Scribe) print(str string) {
	fmt.Fprintf(s.writer, "%v\n", str)
}

func (s *Scribe) printt(str string) {
	s.print(fmt.Sprintf("%v%v", s.space(), str))
}

func (s *Scribe) println() {
	s.print("")
}

func (s *Scribe) space() string {
	return strings.Repeat(" ", s.level*2)
}
