package tk

import (
	"bufio"
	"io"
)

const (
	stateOther = iota
	stateQuotation
	stateEscape
	stateSComment
	stateMComment
)

type readerState int

type reader struct {
	input         *bufio.Reader
	err           error
	buff          string
	state         readerState
	commentSingle string

	commentMultiStart string
	commentMultiEnd   string

	quoteMarks []string
}

func getFromNext(next, comment string) string {
	if len(comment) == 0 {
		return ""
	}

	if len(next)+1 < len(comment) {
		return next
	}

	return next[:len(comment)-1]
}

func stringInSlice(list []string, str string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}

	return false
}

func (reader *reader) checkForPreviousError(buff []byte) (count int, err error) {
	if reader.err != nil {
		buffLen := len(buff)
		err = reader.err

		if reader.buff != "" {
			if reader.state == stateOther || reader.state == stateQuotation || reader.state == stateEscape {
				readBuffBytes := []byte(reader.buff)
				readBuffLen := len(readBuffBytes)

				if readBuffLen <= buffLen {
					copy(buff, readBuffBytes)
					reader.buff = ""
					count = readBuffLen
				} else {
					err = io.ErrShortBuffer
				}
				return
			}

			reader.buff = ""
		}
	}

	return
}

func writeStringToBuffer(buff []byte, pos int, value string) (count int) {
	for _, b := range []byte(value) {
		buff[pos+count] = b
		count++
	}

	return
}

func (reader *reader) readRuneAsString() (value string, err error) {
	var char rune
	char, _, err = reader.input.ReadRune()

	if err == nil {
		value = string(char)
	}
	return
}

func (reader *reader) readRunesAsString(n int) (string, error) {
	var chars []rune
	for i := 0; i < n; i++ {
		char, _, err := reader.input.ReadRune()
		if err == nil {
			chars = append(chars, char)
		}
	}

	return string(chars), nil
}

func (reader *reader) fillBuff() (err error) {
	if reader.buff == "" {
		var buff string
		buff, err = reader.readRuneAsString()
		if err == nil {
			reader.buff = buff
		}
	}
	return
}

func (reader *reader) processOther(next string, buff []byte, pos int) (count int) {
	if reader.buff+getFromNext(next, reader.commentMultiStart) == reader.commentMultiStart {
		reader.state = stateMComment
		reader.buff = ""
		return
	}

	if reader.buff+getFromNext(next, reader.commentSingle) == reader.commentSingle {
		reader.state = stateSComment
		reader.buff = ""
		return
	}

	if stringInSlice(reader.quoteMarks, reader.buff) {
		reader.state = stateQuotation
	}

	count = writeStringToBuffer(buff, pos, reader.buff)
	reader.buff = next

	return
}

func (reader *reader) processMComment(next string, buff []byte, pos int) (count int) {
	if reader.buff+getFromNext(next, reader.commentMultiEnd) == reader.commentMultiEnd {
		reader.state = stateOther
		reader.buff = ""
		return
	}
	reader.buff = next
	return
}

func (reader *reader) processSComment(next string, buff []byte, pos int) (count int) {
	if reader.buff == "\n" || reader.buff == "\r" && next == "\n" {
		reader.state = stateOther

		// keep line endings
		count += writeStringToBuffer(buff, pos, reader.buff)
	}
	reader.buff = next
	return
}

func (reader *reader) processQuotation(next string, buff []byte, pos int) (count int) {
	if reader.buff == "\\" {
		reader.state = stateEscape
	} else if stringInSlice(reader.quoteMarks, reader.buff) {
		reader.state = stateOther
	}
	count += writeStringToBuffer(buff, pos, reader.buff)
	reader.buff = next
	return
}

func (reader *reader) processEscape(next string, buff []byte, pos int) (count int) {
	reader.state = stateQuotation
	count += writeStringToBuffer(buff, pos, reader.buff)
	reader.buff = next
	return
}

func (reader *reader) processNextRune(next string, buff []byte, pos int) (count int) {
	switch reader.state {
	case stateOther:
		count = reader.processOther(next, buff, pos)
	case stateMComment:
		count = reader.processMComment(next, buff, pos)
	case stateSComment:
		count = reader.processSComment(next, buff, pos)
	case stateQuotation:
		count = reader.processQuotation(next, buff, pos)
	case stateEscape:
		count = reader.processEscape(next, buff, pos)
	}
	return
}

func (reader *reader) Read(buff []byte) (count int, err error) {
	count, err = reader.checkForPreviousError(buff)

	if err != nil {
		return
	}

	var next string
	buffLen := len(buff)
	for count < buffLen {
		reader.err = reader.fillBuff()
		if reader.err != nil {
			return
		}

		// check buffer free space
		if buffLen-count < len(reader.buff) {
			if count == 0 {
				err = io.ErrShortBuffer
			}
			return
		}

		next, reader.err = reader.readRuneAsString()
		if reader.err != nil {
			return
		}

		count += reader.processNextRune(next, buff, count)
	}

	return
}

// NewReader creates new Reader instance
func NewReader(input io.Reader) *reader {
	return &reader{
		input:             bufio.NewReader(input),
		state:             stateOther,
		commentSingle:     "//",
		commentMultiStart: "/*",
		commentMultiEnd:   "*/",
		quoteMarks:        []string{"'", "\"", "`"},
	}
}
