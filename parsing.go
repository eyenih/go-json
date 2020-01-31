package json

import (
	"fmt"
	"io"
	"unicode"

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

type Mapper interface {
	SetString(string, string)
	MapperFor(string) Mapper
}

type GrammarStateMachine struct {
	s            State
	nests        [50]Input
	currentLevel int

	keyBuf   string
	valueBuf string
	mappers  [50]Mapper
}

func NewGrammarStateMachine(m Mapper) *GrammarStateMachine {
	return &GrammarStateMachine{mappers: [50]Mapper{m}}
}

func inputType(i interface{}) Input {
	switch i.(byte) {
	case ' ', '\r', '\n', '\t':
		return Whitespace
	case '{':
		return CurlyBracketOpen
	case '}':
		return CurlyBracketClose
	case '[':
		return SquareBracketOpen
	case ']':
		return SquareBracketClose
	case '"':
		return DoubleQuotationMark
	case ':':
		return Colon
	case ',':
		return Comma
	case '.':
		return Dot
	default:
		if unicode.IsNumber(rune(i.(byte))) {
			return Num
		}
		return Character
	}
}

type NoTransitionFunc struct {
	s State
	i Input
}

func (e NoTransitionFunc) Error() string {
	return fmt.Sprintf("No transition func for %T(%d) and %T(%d).", e.s, e.s, e.i, e.i)
}

func (sm *GrammarStateMachine) Transition(i interface{}) error {
	if i == nil {
		return nil
	}

	if f, ok := StateInputTable[sm.s][inputType(i)]; !ok {
		return NoTransitionFunc{sm.s, inputType(i)}
	} else {
		f(i, sm)
		//fmt.Println("State:", sm.s)
	}

	return nil
}

const (
	Nil State = iota
	InsideObject
	InsideObjectKey
	BetweenKeyAndColon
	BetweenColonAndValue
	InsideStringValue
	InsideNumberValue
	AfterValue
	BetweenMembers
	InsideArray
)

type State uint8

const (
	Whitespace Input = iota
	CurlyBracketOpen
	CurlyBracketClose
	DoubleQuotationMark
	SquareBracketOpen
	SquareBracketClose
	Colon
	Comma
	Dot
	Num
	Character
)

type Input uint8

type TransitionFunc func(interface{}, *GrammarStateMachine)

func (sm *GrammarStateMachine) addNest(i Input) {
	if sm.currentLevel > 0 && sm.nests[sm.currentLevel-1] != SquareBracketOpen {
		//fmt.Println("new submapper for:", sm.keyBuf)
		sm.mappers[sm.currentLevel] = sm.mappers[sm.currentLevel-1].MapperFor(sm.keyBuf)
		if sm.mappers[sm.currentLevel] == nil {
			panic("no mapper for key: " + sm.keyBuf + " at the level: " + string(sm.currentLevel))
		}
		sm.keyBuf = ""
	}

	switch i {
	case CurlyBracketOpen:
		sm.s = InsideObject
	case SquareBracketOpen:
		sm.s = InsideArray
	}
	sm.nests[sm.currentLevel] = i
	sm.currentLevel++
	//	fmt.Println("nest added:", sm.nests, sm.currentLevel)
}

func (sm *GrammarStateMachine) removeNest(i Input) {
	if sm.nests[sm.currentLevel-1] == CurlyBracketOpen && i == CurlyBracketClose {
		sm.currentLevel--
		sm.s = AfterValue
	} else if sm.nests[sm.currentLevel-1] == SquareBracketOpen && i == SquareBracketClose {
		sm.currentLevel--
		sm.s = AfterValue
	} else {
		panic("invalid JSON")
	}

	if sm.currentLevel == 0 {
		sm.s = Nil
	}
	//	fmt.Println("nest removed:", sm.nests, sm.currentLevel)
}

var StateInputTable = map[State]map[Input]TransitionFunc{
	Nil: map[Input]TransitionFunc{
		Whitespace: func(i interface{}, sm *GrammarStateMachine) {},
		CurlyBracketOpen: func(i interface{}, sm *GrammarStateMachine) {
			sm.addNest(CurlyBracketOpen)
		},
	},
	InsideObject: map[Input]TransitionFunc{
		Whitespace:          func(i interface{}, sm *GrammarStateMachine) {},
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) { sm.s = InsideObjectKey },
		CurlyBracketClose: func(i interface{}, sm *GrammarStateMachine) {
			sm.removeNest(CurlyBracketClose)
		},
	},
	InsideObjectKey: map[Input]TransitionFunc{
		Character: func(i interface{}, sm *GrammarStateMachine) {
			sm.keyBuf += string(i.(byte))
		},
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) { sm.s = BetweenKeyAndColon },
	},
	BetweenKeyAndColon: map[Input]TransitionFunc{
		Colon: func(i interface{}, sm *GrammarStateMachine) { sm.s = BetweenColonAndValue },
	},
	BetweenColonAndValue: map[Input]TransitionFunc{
		Whitespace: func(i interface{}, sm *GrammarStateMachine) {},
		Num:        func(i interface{}, sm *GrammarStateMachine) { sm.s = InsideNumberValue },
		CurlyBracketOpen: func(i interface{}, sm *GrammarStateMachine) {
			sm.addNest(CurlyBracketOpen)
		},
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) { sm.s = InsideStringValue },
		SquareBracketOpen: func(i interface{}, sm *GrammarStateMachine) {
			sm.addNest(SquareBracketOpen)
		},
	},
	InsideStringValue: map[Input]TransitionFunc{
		Character: func(i interface{}, sm *GrammarStateMachine) {
			sm.valueBuf += string(i.(byte))
		},
		Whitespace: func(i interface{}, sm *GrammarStateMachine) {
			sm.valueBuf += string(i.(byte))
		},
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) {
			sm.s = AfterValue
			sm.mappers[sm.currentLevel-1].SetString(sm.keyBuf, sm.valueBuf)
			sm.keyBuf = ""
		},
	},
	InsideNumberValue: map[Input]TransitionFunc{
		Num: func(i interface{}, sm *GrammarStateMachine) {},
		Dot: func(i interface{}, sm *GrammarStateMachine) {},
		Whitespace: func(i interface{}, sm *GrammarStateMachine) {
			sm.s = AfterValue
			sm.keyBuf = ""
		},
	},
	AfterValue: map[Input]TransitionFunc{
		Comma:      func(i interface{}, sm *GrammarStateMachine) { sm.s = BetweenMembers },
		Whitespace: func(i interface{}, sm *GrammarStateMachine) {},
		CurlyBracketClose: func(i interface{}, sm *GrammarStateMachine) {
			sm.removeNest(CurlyBracketClose)
		},
		SquareBracketClose: func(i interface{}, sm *GrammarStateMachine) {
			sm.removeNest(SquareBracketClose)
		},
	},
	BetweenMembers: map[Input]TransitionFunc{
		Whitespace:          func(i interface{}, sm *GrammarStateMachine) {},
		DoubleQuotationMark: func(i interface{}, sm *GrammarStateMachine) { sm.s = InsideObjectKey },
	},
	InsideArray: map[Input]TransitionFunc{
		CurlyBracketOpen: func(i interface{}, sm *GrammarStateMachine) {
			sm.addNest(CurlyBracketOpen)
		},
	},
}

func Parse(it *TextIterator, sm *GrammarStateMachine) error {
	return moc.Execute(it, sm)
}
