package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// ============================================================================
// Value Types
// ============================================================================

type ValueType int

const (
	TypeNil ValueType = iota
	TypeSymbol
	TypeNumber
	TypeString
	TypeList
	TypeFunc
	TypeBuiltin
	TypeStack
	TypeQueue
	TypeBool
	TypeTailCall
	TypeBlocked
)

type Value struct {
	Type    ValueType
	Symbol  string
	Number  float64
	Str     string
	List    []Value
	Func    *Function
	Builtin func(*Evaluator, []Value, *Env) Value
	Stack   *BoundedStack
	Queue   *BoundedQueue
	Bool    bool
	Tail    *TailCall
	Blocked *BlockedOp
}

type Function struct {
	Params []string
	Body   Value
	Env    *Env
	IsTail bool
}

type TailCall struct {
	Func Value
	Args []Value
}

type BlockReason int

const (
	BlockNone BlockReason = iota
	BlockStackFull
	BlockStackEmpty
	BlockQueueFull
	BlockQueueEmpty
	BlockCallStackFull
)

type BlockedOp struct {
	Reason   BlockReason
	Resource interface{}
}

// Value constructors
func Nil() Value                     { return Value{Type: TypeNil} }
func Sym(s string) Value             { return Value{Type: TypeSymbol, Symbol: s} }
func Num(n float64) Value            { return Value{Type: TypeNumber, Number: n} }
func Str(s string) Value             { return Value{Type: TypeString, Str: s} }
func Lst(items ...Value) Value       { return Value{Type: TypeList, List: items} }
func Bool(b bool) Value              { return Value{Type: TypeBool, Bool: b} }
func Blocked(r BlockReason) Value    { return Value{Type: TypeBlocked, Blocked: &BlockedOp{Reason: r}} }

func (v Value) IsNil() bool    { return v.Type == TypeNil }
func (v Value) IsList() bool   { return v.Type == TypeList }
func (v Value) IsSymbol() bool { return v.Type == TypeSymbol }
func (v Value) IsTruthy() bool {
	switch v.Type {
	case TypeNil:
		return false
	case TypeBool:
		return v.Bool
	case TypeList:
		return len(v.List) > 0
	case TypeNumber:
		return v.Number != 0
	case TypeString:
		return v.Str != ""
	default:
		return true
	}
}

func (v Value) String() string {
	switch v.Type {
	case TypeNil:
		return "nil"
	case TypeSymbol:
		return v.Symbol
	case TypeNumber:
		if v.Number == float64(int64(v.Number)) {
			return fmt.Sprintf("%d", int64(v.Number))
		}
		return fmt.Sprintf("%g", v.Number)
	case TypeString:
		return fmt.Sprintf("%q", v.Str)
	case TypeBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case TypeList:
		parts := make([]string, len(v.List))
		for i, item := range v.List {
			parts[i] = item.String()
		}
		return "(" + strings.Join(parts, " ") + ")"
	case TypeFunc:
		return "<function>"
	case TypeBuiltin:
		return "<builtin>"
	case TypeStack:
		return fmt.Sprintf("<stack %d/%d>", len(v.Stack.Data), v.Stack.Capacity)
	case TypeQueue:
		return fmt.Sprintf("<queue %d/%d>", len(v.Queue.Data), v.Queue.Capacity)
	case TypeBlocked:
		return fmt.Sprintf("<blocked: %d>", v.Blocked.Reason)
	default:
		return "<unknown>"
	}
}

// ============================================================================
// Bounded Data Structures
// ============================================================================

type BoundedStack struct {
	Capacity int
	Data     []Value
}

func NewStack(capacity int) *BoundedStack {
	return &BoundedStack{
		Capacity: capacity,
		Data:     make([]Value, 0, capacity),
	}
}

func (s *BoundedStack) IsFull() bool  { return len(s.Data) >= s.Capacity }
func (s *BoundedStack) IsEmpty() bool { return len(s.Data) == 0 }

func (s *BoundedStack) PushNow(v Value) bool {
	if s.IsFull() {
		return false
	}
	s.Data = append(s.Data, v)
	return true
}

func (s *BoundedStack) PopNow() (Value, bool) {
	if s.IsEmpty() {
		return Nil(), false
	}
	v := s.Data[len(s.Data)-1]
	s.Data = s.Data[:len(s.Data)-1]
	return v, true
}

func (s *BoundedStack) PeekNow() (Value, bool) {
	if s.IsEmpty() {
		return Nil(), false
	}
	return s.Data[len(s.Data)-1], true
}

func (s *BoundedStack) Read(index int) (Value, bool) {
	if index >= 0 && index < len(s.Data) {
		return s.Data[index], true
	}
	return Nil(), false
}

func (s *BoundedStack) Write(index int, v Value) bool {
	if index >= 0 && index < len(s.Data) {
		s.Data[index] = v
		return true
	}
	return false
}

type BoundedQueue struct {
	Capacity int
	Data     []Value
}

func NewQueue(capacity int) *BoundedQueue {
	return &BoundedQueue{
		Capacity: capacity,
		Data:     make([]Value, 0, capacity),
	}
}

func (q *BoundedQueue) IsFull() bool  { return len(q.Data) >= q.Capacity }
func (q *BoundedQueue) IsEmpty() bool { return len(q.Data) == 0 }

func (q *BoundedQueue) SendNow(v Value) bool {
	if q.IsFull() {
		return false
	}
	q.Data = append(q.Data, v)
	return true
}

func (q *BoundedQueue) RecvNow() (Value, bool) {
	if q.IsEmpty() {
		return Nil(), false
	}
	v := q.Data[0]
	q.Data = q.Data[1:]
	return v, true
}

func (q *BoundedQueue) PeekNow() (Value, bool) {
	if q.IsEmpty() {
		return Nil(), false
	}
	return q.Data[0], true
}

// ============================================================================
// Tokenizer
// ============================================================================

type TokenType int

const (
	TokLParen TokenType = iota
	TokRParen
	TokQuote
	TokSymbol
	TokNumber
	TokString
	TokEOF
)

type Token struct {
	Type   TokenType
	Text   string
	Number float64
}

type Tokenizer struct {
	input []rune
	pos   int
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{input: []rune(input), pos: 0}
}

func (t *Tokenizer) peek() rune {
	if t.pos >= len(t.input) {
		return 0
	}
	return t.input[t.pos]
}

func (t *Tokenizer) advance() rune {
	if t.pos >= len(t.input) {
		return 0
	}
	r := t.input[t.pos]
	t.pos++
	return r
}

func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.input) {
		c := t.peek()
		if c == ';' {
			for t.pos < len(t.input) && t.peek() != '\n' {
				t.advance()
			}
		} else if unicode.IsSpace(c) {
			t.advance()
		} else {
			break
		}
	}
}

func (t *Tokenizer) Next() Token {
	t.skipWhitespace()

	if t.pos >= len(t.input) {
		return Token{Type: TokEOF}
	}

	c := t.peek()

	switch c {
	case '(':
		t.advance()
		return Token{Type: TokLParen}
	case ')':
		t.advance()
		return Token{Type: TokRParen}
	case '\'':
		t.advance()
		return Token{Type: TokQuote}
	case '"':
		t.advance()
		var sb strings.Builder
		for t.pos < len(t.input) && t.peek() != '"' {
			if t.peek() == '\\' {
				t.advance()
				switch t.peek() {
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case '"':
					sb.WriteRune('"')
				case '\\':
					sb.WriteRune('\\')
				default:
					sb.WriteRune(t.peek())
				}
				t.advance()
			} else {
				sb.WriteRune(t.advance())
			}
		}
		t.advance() // closing quote
		return Token{Type: TokString, Text: sb.String()}
	default:
		var sb strings.Builder
		for t.pos < len(t.input) {
			c := t.peek()
			if unicode.IsSpace(c) || c == '(' || c == ')' || c == '\'' || c == '"' {
				break
			}
			sb.WriteRune(t.advance())
		}
		text := sb.String()

		// Try parsing as number
		if n, err := strconv.ParseFloat(text, 64); err == nil {
			return Token{Type: TokNumber, Number: n, Text: text}
		}

		return Token{Type: TokSymbol, Text: text}
	}
}

// ============================================================================
// Parser
// ============================================================================

type Parser struct {
	tokenizer *Tokenizer
	current   Token
}

func NewParser(input string) *Parser {
	p := &Parser{tokenizer: NewTokenizer(input)}
	p.current = p.tokenizer.Next()
	return p
}

func (p *Parser) advance() Token {
	tok := p.current
	p.current = p.tokenizer.Next()
	return tok
}

func (p *Parser) Parse() []Value {
	var exprs []Value
	for p.current.Type != TokEOF {
		exprs = append(exprs, p.parseExpr())
	}
	return exprs
}

func (p *Parser) parseExpr() Value {
	switch p.current.Type {
	case TokLParen:
		p.advance()
		var items []Value
		for p.current.Type != TokRParen && p.current.Type != TokEOF {
			items = append(items, p.parseExpr())
		}
		p.advance() // consume ')'
		return Lst(items...)

	case TokQuote:
		// Just return quote as a symbol, let evaluator handle (' ...) forms
		p.advance()
		return Sym("'")

	case TokNumber:
		tok := p.advance()
		return Num(tok.Number)

	case TokString:
		tok := p.advance()
		return Str(tok.Text)

	case TokSymbol:
		tok := p.advance()
		switch tok.Text {
		case "true":
			return Bool(true)
		case "false":
			return Bool(false)
		case "nil":
			return Nil()
		default:
			return Sym(tok.Text)
		}

	default:
		p.advance()
		return Nil()
	}
}

// ============================================================================
// Environment
// ============================================================================

type Env struct {
	bindings map[string]Value
	parent   *Env
}

func NewEnv(parent *Env) *Env {
	return &Env{
		bindings: make(map[string]Value),
		parent:   parent,
	}
}

func (e *Env) Get(name string) (Value, bool) {
	if v, ok := e.bindings[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return Nil(), false
}

func (e *Env) Set(name string, val Value) {
	e.bindings[name] = val
}

func (e *Env) SetLocal(name string, val Value) {
	if _, ok := e.bindings[name]; ok {
		e.bindings[name] = val
		return
	}
	if e.parent != nil {
		if _, ok := e.parent.Get(name); ok {
			e.parent.SetLocal(name, val)
			return
		}
	}
	e.bindings[name] = val
}

// ============================================================================
// Evaluator
// ============================================================================

type Evaluator struct {
	CallStack *BoundedStack
	GlobalEnv *Env
}

func NewEvaluator(callStackDepth int) *Evaluator {
	ev := &Evaluator{
		CallStack: NewStack(callStackDepth),
		GlobalEnv: NewEnv(nil),
	}
	ev.setupBuiltins()
	return ev
}

func (ev *Evaluator) setupBuiltins() {
	env := ev.GlobalEnv

	// Arithmetic
	env.Set("+", Value{Type: TypeBuiltin, Builtin: builtinAdd})
	env.Set("-", Value{Type: TypeBuiltin, Builtin: builtinSub})
	env.Set("*", Value{Type: TypeBuiltin, Builtin: builtinMul})
	env.Set("/", Value{Type: TypeBuiltin, Builtin: builtinDiv})
	env.Set("mod", Value{Type: TypeBuiltin, Builtin: builtinMod})

	// Comparison
	env.Set("=", Value{Type: TypeBuiltin, Builtin: builtinEq})
	env.Set("!=", Value{Type: TypeBuiltin, Builtin: builtinNeq})
	env.Set("<", Value{Type: TypeBuiltin, Builtin: builtinLt})
	env.Set("<=", Value{Type: TypeBuiltin, Builtin: builtinLte})
	env.Set(">", Value{Type: TypeBuiltin, Builtin: builtinGt})
	env.Set(">=", Value{Type: TypeBuiltin, Builtin: builtinGte})

	// Logic
	env.Set("and", Value{Type: TypeBuiltin, Builtin: builtinAnd})
	env.Set("or", Value{Type: TypeBuiltin, Builtin: builtinOr})
	env.Set("not", Value{Type: TypeBuiltin, Builtin: builtinNot})

	// List operations
	env.Set("first", Value{Type: TypeBuiltin, Builtin: builtinFirst})
	env.Set("rest", Value{Type: TypeBuiltin, Builtin: builtinRest})
	env.Set("cons", Value{Type: TypeBuiltin, Builtin: builtinCons})
	env.Set("append", Value{Type: TypeBuiltin, Builtin: builtinAppend})
	env.Set("list", Value{Type: TypeBuiltin, Builtin: builtinList})
	env.Set("empty?", Value{Type: TypeBuiltin, Builtin: builtinEmpty})
	env.Set("length", Value{Type: TypeBuiltin, Builtin: builtinLength})
	env.Set("nth", Value{Type: TypeBuiltin, Builtin: builtinNth})

	// Type checks
	env.Set("list?", Value{Type: TypeBuiltin, Builtin: builtinIsList})
	env.Set("number?", Value{Type: TypeBuiltin, Builtin: builtinIsNumber})
	env.Set("symbol?", Value{Type: TypeBuiltin, Builtin: builtinIsSymbol})
	env.Set("string?", Value{Type: TypeBuiltin, Builtin: builtinIsString})
	env.Set("nil?", Value{Type: TypeBuiltin, Builtin: builtinIsNil})

	// Bounded structures
	env.Set("make-stack", Value{Type: TypeBuiltin, Builtin: builtinMakeStack})
	env.Set("make-queue", Value{Type: TypeBuiltin, Builtin: builtinMakeQueue})

	// Stack operations (blocking and non-blocking)
	env.Set("push!", Value{Type: TypeBuiltin, Builtin: builtinPush})
	env.Set("pop!", Value{Type: TypeBuiltin, Builtin: builtinPop})
	env.Set("push-now!", Value{Type: TypeBuiltin, Builtin: builtinPushNow})
	env.Set("pop-now!", Value{Type: TypeBuiltin, Builtin: builtinPopNow})
	env.Set("stack-peek", Value{Type: TypeBuiltin, Builtin: builtinStackPeek})
	env.Set("stack-peek-now", Value{Type: TypeBuiltin, Builtin: builtinStackPeekNow})
	env.Set("stack-read", Value{Type: TypeBuiltin, Builtin: builtinStackRead})
	env.Set("stack-write!", Value{Type: TypeBuiltin, Builtin: builtinStackWrite})
	env.Set("stack-full?", Value{Type: TypeBuiltin, Builtin: builtinStackFull})
	env.Set("stack-empty?", Value{Type: TypeBuiltin, Builtin: builtinStackEmpty})

	// Queue operations (blocking and non-blocking)
	env.Set("send!", Value{Type: TypeBuiltin, Builtin: builtinSend})
	env.Set("recv!", Value{Type: TypeBuiltin, Builtin: builtinRecv})
	env.Set("send-now!", Value{Type: TypeBuiltin, Builtin: builtinSendNow})
	env.Set("recv-now!", Value{Type: TypeBuiltin, Builtin: builtinRecvNow})
	env.Set("queue-peek", Value{Type: TypeBuiltin, Builtin: builtinQueuePeek})
	env.Set("queue-peek-now", Value{Type: TypeBuiltin, Builtin: builtinQueuePeekNow})
	env.Set("queue-full?", Value{Type: TypeBuiltin, Builtin: builtinQueueFull})
	env.Set("queue-empty?", Value{Type: TypeBuiltin, Builtin: builtinQueueEmpty})

	// I/O
	env.Set("print", Value{Type: TypeBuiltin, Builtin: builtinPrint})
	env.Set("println", Value{Type: TypeBuiltin, Builtin: builtinPrintln})
	env.Set("repr", Value{Type: TypeBuiltin, Builtin: builtinRepr})
}

func (ev *Evaluator) Eval(expr Value, env *Env) Value {
	if env == nil {
		env = ev.GlobalEnv
	}

	// Trampoline loop for tail calls
	for {
		result := ev.evalStep(expr, env)

		if result.Type == TypeTailCall {
			tc := result.Tail
			if tc.Func.Type == TypeFunc {
				fn := tc.Func.Func
				env = NewEnv(fn.Env)
				for i, param := range fn.Params {
					if i < len(tc.Args) {
						env.Set(param, tc.Args[i])
					}
				}
				expr = fn.Body
			} else {
				// Not a function, just call normally
				args := make([]Value, len(tc.Args))
				for i, arg := range tc.Args {
					args[i] = ev.Eval(arg, env)
				}
				return ev.apply(tc.Func, args, env)
			}
		} else {
			return result
		}
	}
}

func (ev *Evaluator) evalStep(expr Value, env *Env) Value {
	switch expr.Type {
	case TypeNil, TypeNumber, TypeString, TypeBool, TypeFunc, TypeBuiltin, TypeStack, TypeQueue:
		return expr

	case TypeSymbol:
		if v, ok := env.Get(expr.Symbol); ok {
			return v
		}
		fmt.Fprintf(os.Stderr, "Undefined symbol: %s\n", expr.Symbol)
		return Nil()

	case TypeList:
		if len(expr.List) == 0 {
			return expr
		}

		head := expr.List[0]

		// Special forms
		if head.IsSymbol() {
			switch head.Symbol {
			case "'": // Quote
				if len(expr.List) == 2 {
					// (' x) - return x unevaluated
					return expr.List[1]
				} else if len(expr.List) > 2 {
					// (' a b c) - return (a b c) as a list
					return Lst(expr.List[1:]...)
				}
				return Nil()

			case "if":
				if len(expr.List) < 3 {
					return Nil()
				}
				cond := ev.Eval(expr.List[1], env)
				if cond.IsTruthy() {
					return ev.Eval(expr.List[2], env)
				} else if len(expr.List) > 3 {
					return ev.Eval(expr.List[3], env)
				}
				return Nil()

			case "cond":
				for i := 1; i < len(expr.List); i++ {
					clause := expr.List[i]
					if !clause.IsList() || len(clause.List) < 2 {
						continue
					}
					test := clause.List[0]
					if test.IsSymbol() && test.Symbol == "else" {
						return ev.Eval(clause.List[1], env)
					}
					if ev.Eval(test, env).IsTruthy() {
						return ev.Eval(clause.List[1], env)
					}
				}
				return Nil()

			case "let":
				if len(expr.List) < 3 {
					return Nil()
				}
				name := expr.List[1]
				val := ev.Eval(expr.List[2], env)
				newEnv := NewEnv(env)
				newEnv.Set(name.Symbol, val)
				if len(expr.List) > 3 {
					return ev.Eval(expr.List[3], newEnv)
				}
				return val

			case "let*":
				if len(expr.List) < 3 {
					return Nil()
				}
				bindings := expr.List[1]
				newEnv := NewEnv(env)
				if bindings.IsList() {
					for _, binding := range bindings.List {
						if binding.IsList() && len(binding.List) >= 2 {
							name := binding.List[0].Symbol
							val := ev.Eval(binding.List[1], newEnv)
							newEnv.Set(name, val)
						}
					}
				}
				return ev.Eval(expr.List[2], newEnv)

			case "set!":
				if len(expr.List) < 3 {
					return Nil()
				}
				name := expr.List[1].Symbol
				val := ev.Eval(expr.List[2], env)
				env.SetLocal(name, val)
				return val

			case "define":
				if len(expr.List) < 3 {
					return Nil()
				}
				// (define name value) or (define (name args...) body)
				if expr.List[1].IsList() {
					// Function shorthand
					sig := expr.List[1].List
					name := sig[0].Symbol
					params := make([]string, len(sig)-1)
					for i, p := range sig[1:] {
						params[i] = p.Symbol
					}
					fn := &Function{
						Params: params,
						Body:   expr.List[2],
						Env:    env,
					}
					val := Value{Type: TypeFunc, Func: fn}
					ev.GlobalEnv.Set(name, val)
					return val
				} else {
					name := expr.List[1].Symbol
					val := ev.Eval(expr.List[2], env)
					ev.GlobalEnv.Set(name, val)
					return val
				}

			case "lambda", "fn":
				if len(expr.List) < 3 {
					return Nil()
				}
				params := make([]string, 0)
				if expr.List[1].IsList() {
					for _, p := range expr.List[1].List {
						params = append(params, p.Symbol)
					}
				}
				return Value{
					Type: TypeFunc,
					Func: &Function{
						Params: params,
						Body:   expr.List[2],
						Env:    env,
					},
				}

			case "tail":
				// Tail call - evaluate args but return TailCall marker
				if len(expr.List) < 2 {
					return Nil()
				}
				fn := ev.Eval(expr.List[1], env)
				args := make([]Value, len(expr.List)-2)
				for i, arg := range expr.List[2:] {
					args[i] = ev.Eval(arg, env)
				}
				return Value{
					Type: TypeTailCall,
					Tail: &TailCall{Func: fn, Args: args},
				}

			case "do", "begin":
				var result Value = Nil()
				for _, e := range expr.List[1:] {
					result = ev.Eval(e, env)
				}
				return result

			case "match":
				if len(expr.List) < 2 {
					return Nil()
				}
				target := ev.Eval(expr.List[1], env)
				for i := 2; i < len(expr.List); i++ {
					clause := expr.List[i]
					if !clause.IsList() || len(clause.List) < 2 {
						continue
					}
					pattern := clause.List[0]
					body := clause.List[1]
					if bindings, ok := ev.match(pattern, target, env); ok {
						newEnv := NewEnv(env)
						for k, v := range bindings {
							newEnv.Set(k, v)
						}
						return ev.Eval(body, newEnv)
					}
				}
				return Nil()
			}
		}

		// Function call
		fn := ev.Eval(head, env)
		args := make([]Value, len(expr.List)-1)
		for i, arg := range expr.List[1:] {
			args[i] = ev.Eval(arg, env)
		}
		return ev.apply(fn, args, env)
	}

	return Nil()
}

func (ev *Evaluator) apply(fn Value, args []Value, env *Env) Value {
	switch fn.Type {
	case TypeBuiltin:
		return fn.Builtin(ev, args, env)

	case TypeFunc:
		f := fn.Func
		newEnv := NewEnv(f.Env)
		for i, param := range f.Params {
			if i < len(args) {
				newEnv.Set(param, args[i])
			}
		}

		// Check call stack bounds
		if !ev.CallStack.PushNow(Lst(args...)) {
			return Blocked(BlockCallStackFull)
		}

		result := ev.Eval(f.Body, newEnv)
		ev.CallStack.PopNow()
		return result
	}

	return Nil()
}

func (ev *Evaluator) match(pattern, target Value, env *Env) (map[string]Value, bool) {
	bindings := make(map[string]Value)

	// Wildcard
	if pattern.IsSymbol() && pattern.Symbol == "_" {
		return bindings, true
	}

	// Pattern variable ?name
	if pattern.IsSymbol() && len(pattern.Symbol) > 0 && pattern.Symbol[0] == '?' {
		bindings[pattern.Symbol[1:]] = target
		return bindings, true
	}

	// Quoted symbol matches symbol
	if pattern.IsList() && len(pattern.List) == 2 &&
		pattern.List[0].IsSymbol() && pattern.List[0].Symbol == "'" {
		if target.IsSymbol() && target.Symbol == pattern.List[1].Symbol {
			return bindings, true
		}
		return nil, false
	}

	// Literal match
	if pattern.Type == target.Type {
		switch pattern.Type {
		case TypeNil:
			return bindings, true
		case TypeNumber:
			if pattern.Number == target.Number {
				return bindings, true
			}
		case TypeString:
			if pattern.Str == target.Str {
				return bindings, true
			}
		case TypeSymbol:
			if pattern.Symbol == target.Symbol {
				return bindings, true
			}
		case TypeBool:
			if pattern.Bool == target.Bool {
				return bindings, true
			}
		case TypeList:
			if len(pattern.List) != len(target.List) {
				return nil, false
			}
			for i := range pattern.List {
				sub, ok := ev.match(pattern.List[i], target.List[i], env)
				if !ok {
					return nil, false
				}
				for k, v := range sub {
					bindings[k] = v
				}
			}
			return bindings, true
		}
	}

	return nil, false
}

// ============================================================================
// Builtins
// ============================================================================

func builtinAdd(ev *Evaluator, args []Value, env *Env) Value {
	sum := 0.0
	for _, a := range args {
		sum += a.Number
	}
	return Num(sum)
}

func builtinSub(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Num(0)
	}
	if len(args) == 1 {
		return Num(-args[0].Number)
	}
	result := args[0].Number
	for _, a := range args[1:] {
		result -= a.Number
	}
	return Num(result)
}

func builtinMul(ev *Evaluator, args []Value, env *Env) Value {
	product := 1.0
	for _, a := range args {
		product *= a.Number
	}
	return Num(product)
}

func builtinDiv(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Num(0)
	}
	return Num(args[0].Number / args[1].Number)
}

func builtinMod(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Num(0)
	}
	return Num(float64(int64(args[0].Number) % int64(args[1].Number)))
}

func builtinEq(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Bool(true)
	}
	a, b := args[0], args[1]
	if a.Type != b.Type {
		return Bool(false)
	}
	switch a.Type {
	case TypeNumber:
		return Bool(a.Number == b.Number)
	case TypeString:
		return Bool(a.Str == b.Str)
	case TypeSymbol:
		return Bool(a.Symbol == b.Symbol)
	case TypeBool:
		return Bool(a.Bool == b.Bool)
	case TypeNil:
		return Bool(true)
	}
	return Bool(false)
}

func builtinNeq(ev *Evaluator, args []Value, env *Env) Value {
	return Bool(!builtinEq(ev, args, env).Bool)
}

func builtinLt(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Bool(false)
	}
	return Bool(args[0].Number < args[1].Number)
}

func builtinLte(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Bool(false)
	}
	return Bool(args[0].Number <= args[1].Number)
}

func builtinGt(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Bool(false)
	}
	return Bool(args[0].Number > args[1].Number)
}

func builtinGte(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Bool(false)
	}
	return Bool(args[0].Number >= args[1].Number)
}

func builtinAnd(ev *Evaluator, args []Value, env *Env) Value {
	for _, a := range args {
		if !a.IsTruthy() {
			return Bool(false)
		}
	}
	return Bool(true)
}

func builtinOr(ev *Evaluator, args []Value, env *Env) Value {
	for _, a := range args {
		if a.IsTruthy() {
			return Bool(true)
		}
	}
	return Bool(false)
}

func builtinNot(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(true)
	}
	return Bool(!args[0].IsTruthy())
}

func builtinFirst(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 || !args[0].IsList() || len(args[0].List) == 0 {
		return Nil()
	}
	return args[0].List[0]
}

func builtinRest(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 || !args[0].IsList() || len(args[0].List) == 0 {
		return Lst()
	}
	return Lst(args[0].List[1:]...)
}

func builtinCons(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Lst()
	}
	if args[1].IsList() {
		return Lst(append([]Value{args[0]}, args[1].List...)...)
	}
	return Lst(args[0], args[1])
}

func builtinAppend(ev *Evaluator, args []Value, env *Env) Value {
	var result []Value
	for _, a := range args {
		if a.IsList() {
			result = append(result, a.List...)
		} else {
			result = append(result, a)
		}
	}
	return Lst(result...)
}

func builtinList(ev *Evaluator, args []Value, env *Env) Value {
	return Lst(args...)
}

func builtinEmpty(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(true)
	}
	if args[0].IsList() {
		return Bool(len(args[0].List) == 0)
	}
	return Bool(true)
}

func builtinLength(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 || !args[0].IsList() {
		return Num(0)
	}
	return Num(float64(len(args[0].List)))
}

func builtinNth(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || !args[0].IsList() {
		return Nil()
	}
	idx := int(args[1].Number)
	if idx >= 0 && idx < len(args[0].List) {
		return args[0].List[idx]
	}
	return Nil()
}

func builtinIsList(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(false)
	}
	return Bool(args[0].Type == TypeList)
}

func builtinIsNumber(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(false)
	}
	return Bool(args[0].Type == TypeNumber)
}

func builtinIsSymbol(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(false)
	}
	return Bool(args[0].Type == TypeSymbol)
}

func builtinIsString(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(false)
	}
	return Bool(args[0].Type == TypeString)
}

func builtinIsNil(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Bool(true)
	}
	return Bool(args[0].Type == TypeNil)
}

func builtinMakeStack(ev *Evaluator, args []Value, env *Env) Value {
	capacity := 16
	if len(args) > 0 {
		capacity = int(args[0].Number)
	}
	return Value{Type: TypeStack, Stack: NewStack(capacity)}
}

func builtinMakeQueue(ev *Evaluator, args []Value, env *Env) Value {
	capacity := 16
	if len(args) > 0 {
		capacity = int(args[0].Number)
	}
	return Value{Type: TypeQueue, Queue: NewQueue(capacity)}
}

// Stack operations
func builtinPush(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || args[0].Type != TypeStack {
		return Nil()
	}
	stack := args[0].Stack
	if stack.IsFull() {
		return Blocked(BlockStackFull)
	}
	stack.PushNow(args[1])
	return Sym("ok")
}

func builtinPop(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeStack {
		return Nil()
	}
	stack := args[0].Stack
	if stack.IsEmpty() {
		return Blocked(BlockStackEmpty)
	}
	v, _ := stack.PopNow()
	return v
}

func builtinPushNow(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || args[0].Type != TypeStack {
		return Nil()
	}
	if args[0].Stack.PushNow(args[1]) {
		return Sym("ok")
	}
	return Sym("full")
}

func builtinPopNow(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeStack {
		return Nil()
	}
	v, ok := args[0].Stack.PopNow()
	if ok {
		return v
	}
	return Sym("empty")
}

func builtinStackPeek(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeStack {
		return Nil()
	}
	stack := args[0].Stack
	if stack.IsEmpty() {
		return Blocked(BlockStackEmpty)
	}
	v, _ := stack.PeekNow()
	return v
}

func builtinStackPeekNow(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeStack {
		return Nil()
	}
	v, ok := args[0].Stack.PeekNow()
	if ok {
		return v
	}
	return Sym("empty")
}

func builtinStackRead(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || args[0].Type != TypeStack {
		return Nil()
	}
	v, ok := args[0].Stack.Read(int(args[1].Number))
	if ok {
		return v
	}
	return Nil()
}

func builtinStackWrite(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 3 || args[0].Type != TypeStack {
		return Nil()
	}
	if args[0].Stack.Write(int(args[1].Number), args[2]) {
		return Sym("ok")
	}
	return Sym("error")
}

func builtinStackFull(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeStack {
		return Bool(false)
	}
	return Bool(args[0].Stack.IsFull())
}

func builtinStackEmpty(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeStack {
		return Bool(true)
	}
	return Bool(args[0].Stack.IsEmpty())
}

// Queue operations
func builtinSend(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || args[0].Type != TypeQueue {
		return Nil()
	}
	queue := args[0].Queue
	if queue.IsFull() {
		return Blocked(BlockQueueFull)
	}
	queue.SendNow(args[1])
	return Sym("ok")
}

func builtinRecv(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeQueue {
		return Nil()
	}
	queue := args[0].Queue
	if queue.IsEmpty() {
		return Blocked(BlockQueueEmpty)
	}
	v, _ := queue.RecvNow()
	return v
}

func builtinSendNow(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || args[0].Type != TypeQueue {
		return Nil()
	}
	if args[0].Queue.SendNow(args[1]) {
		return Sym("ok")
	}
	return Sym("full")
}

func builtinRecvNow(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeQueue {
		return Nil()
	}
	v, ok := args[0].Queue.RecvNow()
	if ok {
		return v
	}
	return Sym("empty")
}

func builtinQueuePeek(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeQueue {
		return Nil()
	}
	queue := args[0].Queue
	if queue.IsEmpty() {
		return Blocked(BlockQueueEmpty)
	}
	v, _ := queue.PeekNow()
	return v
}

func builtinQueuePeekNow(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeQueue {
		return Nil()
	}
	v, ok := args[0].Queue.PeekNow()
	if ok {
		return v
	}
	return Sym("empty")
}

func builtinQueueFull(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeQueue {
		return Bool(false)
	}
	return Bool(args[0].Queue.IsFull())
}

func builtinQueueEmpty(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeQueue {
		return Bool(true)
	}
	return Bool(args[0].Queue.IsEmpty())
}

// I/O
func builtinPrint(ev *Evaluator, args []Value, env *Env) Value {
	parts := make([]string, len(args))
	for i, a := range args {
		if a.Type == TypeString {
			parts[i] = a.Str
		} else {
			parts[i] = a.String()
		}
	}
	fmt.Print(strings.Join(parts, " "))
	return Nil()
}

func builtinPrintln(ev *Evaluator, args []Value, env *Env) Value {
	builtinPrint(ev, args, env)
	fmt.Println()
	return Nil()
}

func builtinRepr(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Str("")
	}
	return Str(args[0].String())
}

// ============================================================================
// REPL and File Execution
// ============================================================================

func runREPL(ev *Evaluator) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("BoundedLISP - Type (exit) to quit")
	fmt.Print("> ")

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "(exit)" {
			break
		}

		parser := NewParser(line)
		exprs := parser.Parse()

		for _, expr := range exprs {
			result := ev.Eval(expr, nil)
			if result.Type != TypeNil {
				fmt.Println(result.String())
			}
		}

		fmt.Print("> ")
	}
}

func runFile(ev *Evaluator, filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	parser := NewParser(string(content))
	exprs := parser.Parse()

	for _, expr := range exprs {
		result := ev.Eval(expr, nil)
		if result.Type == TypeBlocked {
			fmt.Fprintf(os.Stderr, "Blocked: %v\n", result.Blocked.Reason)
		}
	}
}

func main() {
	ev := NewEvaluator(64) // 64 frame call stack limit

	if len(os.Args) > 1 {
		runFile(ev, os.Args[1])
	} else {
		runREPL(ev)
	}
}
