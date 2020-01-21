package json

import (
	"fmt"
	"io"
	"unicode"

	"github.com/eyenih/go-log"
	"github.com/eyenih/go-moc"
)

type TextIterator struct {
	buf  [1]byte
	r    io.Reader
	done bool
}

func NewTextIterator(r io.Reader) *TextIterator {
	return &TextIterator{r: r}
}

func (it TextIterator) Done() bool {
	return it.done
}

func (it *TextIterator) Next() (interface{}, error) {
	_, err := it.r.Read(it.buf[:])
	if err == io.EOF {
		it.done = true

		return nil, nil
	}

	return it.buf[0], err
}

const (
	Nil State = iota
	Object
	ObjectKey
	ObjectValue
)

type State uint8

type Mapper interface {
	Key(string)
}

type GrammarStateMachine struct {
	buf string
	m   Mapper
	s   State
}

func NewGrammarStateMachine(m Mapper) *GrammarStateMachine {
	return &GrammarStateMachine{"", m, Nil}
}

func inputType(i interface{}) Input {
	switch i.(byte) {
	case ' ', '\r', '\n', '\t':
		return Whitespace
	case '{':
		return CurlyBracketOpen
	case '"':
		return DoubleQuotationMark
	case ':':
		return Colon
	default:
		if unicode.IsNumber(rune(i.(byte))) {
			return Number
		}
		return Character
	}

	panic("no input type found")
}

type NoTransitionFunc struct {
	s State
	i Input
}

func (e NoTransitionFunc) Error() string {
	return fmt.Sprintf("No transition func for %T(%d) and %T(%d).", e.s, e.s, e.i, e.i)
}

func (sm *GrammarStateMachine) Transition(i interface{}) error {
	if f, ok := StateInputTable[sm.s][inputType(i)]; !ok {
		return NoTransitionFunc{sm.s, inputType(i)}
	} else {
		f(i, sm)
	}

	return nil
}

const (
	Whitespace Input = iota
	CurlyBracketOpen
	DoubleQuotationMark
	Number
	Character
	Colon
)

type Input uint8

type TransitionFunc func(interface{}, *GrammarStateMachine)

var StateInputTable = map[State]map[Input]TransitionFunc{
	Nil: map[Input]TransitionFunc{
		CurlyBracketOpen: func(i interface{}, sm *GrammarStateMachine) { sm.s = Object },
		Whitespace:       func(i interface{}, sm *GrammarStateMachine) { sm.s = Nil },
	},
	Object: map[Input]TransitionFunc{
		Whitespace:          func(i interface{}, sm *GrammarStateMachine) { sm.s = Object },
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) { sm.s = ObjectKey },
		Colon:               func(i interface{}, sm *GrammarStateMachine) { sm.s = ObjectValue },
	},
	ObjectKey: map[Input]TransitionFunc{
		Character: func(i interface{}, sm *GrammarStateMachine) {
			sm.s = ObjectKey
			sm.buf += string(i.(byte))
		},
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) {
			sm.s = Object
			sm.m.Key(sm.buf)
		},
	},
}

type Parser struct {
	logger   log.Logger
	executor *moc.Executor
}

func NewParser(l log.Logger) *Parser {
	return &Parser{l, moc.NewExecutor(l)}
}

func (p *Parser) Parse(it *TextIterator, sm *GrammarStateMachine) error {
	return p.executor.Execute(it, sm)
}

func Parse(it *TextIterator, sm *GrammarStateMachine) error {
	return NewParser(nil).Parse(it, sm)
}
