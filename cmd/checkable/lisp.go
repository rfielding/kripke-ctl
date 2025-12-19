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
	TypeTagged
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
	Tagged  *TaggedValue
}

type TaggedValue struct {
	Tag   string
	Value Value
}

type Function struct {
	Params    []string
	RestParam string
	Body      Value
	Env       *Env
	IsTail    bool
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
	case TypeTagged:
		return fmt.Sprintf("#%s{%s}", v.Tagged.Tag, v.Tagged.Value.String())
	case TypeActor:
		return fmt.Sprintf("<actor:%s>", v.Symbol)
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
		
		// Normal list
		var items []Value
		for p.current.Type != TokRParen && p.current.Type != TokEOF {
			items = append(items, p.parseExpr())
		}
		p.advance() // consume ')'
		return Lst(items...)

	case TokQuote:
		p.advance()
		// Quote wraps next expression: 'x -> (quote x)
		expr := p.parseExpr()
		return Lst(Sym("quote"), expr)

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
	CallStack    *BoundedStack
	GlobalEnv    *Env
	Registry     map[string]Value
	GensymCount  int64
	Scheduler    *Scheduler
}

// ============================================================================
// Scheduler and Actors
// ============================================================================

type ActorState int

const (
	ActorRunnable ActorState = iota
	ActorBlocked
	ActorDone
)

type Actor struct {
	Name      string
	Mailbox   *BoundedQueue
	State     ActorState
	BlockedOn string         // Description of what we're blocked on
	Env       *Env           // Actor's local environment
	Code      Value          // Current code to execute (continuation)
	Result    Value          // Last result
}

type Scheduler struct {
	Actors       map[string]*Actor
	RunQueue     []string      // Names of runnable actors
	CurrentActor string        // Currently executing actor
	StepCount    int64
	MaxSteps     int64         // 0 = unlimited
	Trace        bool          // Print execution trace
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Actors:   make(map[string]*Actor),
		RunQueue: make([]string, 0),
		MaxSteps: 0,
		Trace:    false,
	}
}

func (s *Scheduler) AddActor(name string, mailboxSize int, env *Env, code Value) *Actor {
	actor := &Actor{
		Name:    name,
		Mailbox: NewQueue(mailboxSize),
		State:   ActorRunnable,
		Env:     env,
		Code:    code,
		Result:  Nil(),
	}
	s.Actors[name] = actor
	s.RunQueue = append(s.RunQueue, name)
	return actor
}

func (s *Scheduler) GetActor(name string) *Actor {
	return s.Actors[name]
}

func (s *Scheduler) BlockActor(name string, reason string) {
	if actor, ok := s.Actors[name]; ok {
		actor.State = ActorBlocked
		actor.BlockedOn = reason
		// Remove from run queue
		newQueue := make([]string, 0, len(s.RunQueue))
		for _, n := range s.RunQueue {
			if n != name {
				newQueue = append(newQueue, n)
			}
		}
		s.RunQueue = newQueue
	}
}

func (s *Scheduler) UnblockActor(name string) {
	if actor, ok := s.Actors[name]; ok {
		if actor.State == ActorBlocked {
			actor.State = ActorRunnable
			actor.BlockedOn = ""
			s.RunQueue = append(s.RunQueue, name)
		}
	}
}

func (s *Scheduler) MarkDone(name string) {
	if actor, ok := s.Actors[name]; ok {
		actor.State = ActorDone
		// Remove from run queue
		newQueue := make([]string, 0, len(s.RunQueue))
		for _, n := range s.RunQueue {
			if n != name {
				newQueue = append(newQueue, n)
			}
		}
		s.RunQueue = newQueue
	}
}

func (s *Scheduler) IsDeadlocked() bool {
	// Deadlock if no actors are runnable and at least one is blocked
	if len(s.RunQueue) > 0 {
		return false
	}
	for _, actor := range s.Actors {
		if actor.State == ActorBlocked {
			return true
		}
	}
	return false
}

func (s *Scheduler) AllDone() bool {
	for _, actor := range s.Actors {
		if actor.State != ActorDone {
			return false
		}
	}
	return len(s.Actors) > 0
}

func (s *Scheduler) NextActor() *Actor {
	if len(s.RunQueue) == 0 {
		return nil
	}
	name := s.RunQueue[0]
	// Rotate queue (round-robin)
	s.RunQueue = append(s.RunQueue[1:], name)
	s.CurrentActor = name
	return s.Actors[name]
}

func (s *Scheduler) Status() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Step %d:\n", s.StepCount))
	for name, actor := range s.Actors {
		state := "runnable"
		extra := ""
		switch actor.State {
		case ActorBlocked:
			state = "blocked"
			extra = fmt.Sprintf(" on %s", actor.BlockedOn)
		case ActorDone:
			state = "done"
		}
		sb.WriteString(fmt.Sprintf("  %s: %s%s (mailbox: %d/%d)\n", 
			name, state, extra, len(actor.Mailbox.Data), actor.Mailbox.Capacity))
	}
	return sb.String()
}

func NewEvaluator(callStackDepth int) *Evaluator {
	ev := &Evaluator{
		CallStack:   NewStack(callStackDepth),
		GlobalEnv:   NewEnv(nil),
		Registry:    make(map[string]Value),
		GensymCount: 0,
		Scheduler:   NewScheduler(),
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

	// Evaluation
	env.Set("eval", Value{Type: TypeBuiltin, Builtin: builtinEval})

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

	// String operations
	env.Set("string-append", Value{Type: TypeBuiltin, Builtin: builtinStringAppend})
	env.Set("symbol->string", Value{Type: TypeBuiltin, Builtin: builtinSymbolToString})
	env.Set("string->symbol", Value{Type: TypeBuiltin, Builtin: builtinStringToSymbol})
	env.Set("number->string", Value{Type: TypeBuiltin, Builtin: builtinNumberToString})

	// Registry
	env.Set("registry-set!", Value{Type: TypeBuiltin, Builtin: builtinRegistrySet})
	env.Set("registry-get", Value{Type: TypeBuiltin, Builtin: builtinRegistryGet})
	env.Set("registry-keys", Value{Type: TypeBuiltin, Builtin: builtinRegistryKeys})
	env.Set("registry-has?", Value{Type: TypeBuiltin, Builtin: builtinRegistryHas})
	env.Set("registry-delete!", Value{Type: TypeBuiltin, Builtin: builtinRegistryDelete})

	// Type tagging
	env.Set("tag", Value{Type: TypeBuiltin, Builtin: builtinTag})
	env.Set("tag-type", Value{Type: TypeBuiltin, Builtin: builtinTagType})
	env.Set("tag-value", Value{Type: TypeBuiltin, Builtin: builtinTagValue})
	env.Set("tagged?", Value{Type: TypeBuiltin, Builtin: builtinIsTagged})
	env.Set("tag-is?", Value{Type: TypeBuiltin, Builtin: builtinTagIs})

	// Symbol generation
	env.Set("gensym", Value{Type: TypeBuiltin, Builtin: builtinGensym})

	// Scheduler and actor management
	env.Set("spawn-actor", Value{Type: TypeBuiltin, Builtin: builtinSpawnActor})
	env.Set("self", Value{Type: TypeBuiltin, Builtin: builtinSelf})
	env.Set("send-to!", Value{Type: TypeBuiltin, Builtin: builtinSendTo})
	env.Set("receive!", Value{Type: TypeBuiltin, Builtin: builtinReceive})
	env.Set("receive-now!", Value{Type: TypeBuiltin, Builtin: builtinReceiveNow})
	env.Set("mailbox-empty?", Value{Type: TypeBuiltin, Builtin: builtinMailboxEmpty})
	env.Set("mailbox-full?", Value{Type: TypeBuiltin, Builtin: builtinMailboxFull})
	env.Set("yield!", Value{Type: TypeBuiltin, Builtin: builtinYield})
	env.Set("done!", Value{Type: TypeBuiltin, Builtin: builtinDone})
	env.Set("run-scheduler", Value{Type: TypeBuiltin, Builtin: builtinRunScheduler})
	env.Set("scheduler-status", Value{Type: TypeBuiltin, Builtin: builtinSchedulerStatus})
	env.Set("set-trace!", Value{Type: TypeBuiltin, Builtin: builtinSetTrace})
	env.Set("actor-state", Value{Type: TypeBuiltin, Builtin: builtinActorState})
	env.Set("list-actors-sched", Value{Type: TypeBuiltin, Builtin: builtinListActorsSched})
	env.Set("reset-scheduler", Value{Type: TypeBuiltin, Builtin: builtinResetScheduler})
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
				
				// Bind regular parameters
				for i, param := range fn.Params {
					if i < len(tc.Args) {
						env.Set(param, tc.Args[i])
					} else {
						env.Set(param, Nil())
					}
				}
				
				// Bind rest parameter if present
				if fn.RestParam != "" {
					restArgs := make([]Value, 0)
					if len(tc.Args) > len(fn.Params) {
						restArgs = tc.Args[len(fn.Params):]
					}
					env.Set(fn.RestParam, Lst(restArgs...))
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
			case "quote": // Quote - return argument unevaluated
				if len(expr.List) > 1 {
					return expr.List[1]
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
				// Propagate blocked status
				if val.Type == TypeBlocked {
					return val
				}
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
				// Try to set in existing scope, fall back to global
				if _, found := env.Get(name); found {
					env.SetLocal(name, val)
				} else {
					ev.GlobalEnv.Set(name, val)
				}
				return val

			case "define":
				if len(expr.List) < 3 {
					return Nil()
				}
				// (define name value) or (define (name args...) body...)
				if expr.List[1].IsList() {
					// Function shorthand
					sig := expr.List[1].List
					name := sig[0].Symbol
					params := make([]string, 0)
					restParam := ""
					sigParams := sig[1:] // Parameters part of signature
					for i := 0; i < len(sigParams); i++ {
						p := sigParams[i]
						if p.IsSymbol() && p.Symbol == "." {
							// Rest parameter: next symbol is the rest param name
							if i+1 < len(sigParams) && sigParams[i+1].IsSymbol() {
								restParam = sigParams[i+1].Symbol
							}
							break
						}
						if p.IsSymbol() {
							params = append(params, p.Symbol)
						}
					}
					// Handle multi-expression body: wrap in implicit begin
					var body Value
					if len(expr.List) == 3 {
						body = expr.List[2]
					} else {
						// Multiple body expressions - wrap in begin
						bodyExprs := make([]Value, len(expr.List)-2+1)
						bodyExprs[0] = Sym("begin")
						copy(bodyExprs[1:], expr.List[2:])
						body = Lst(bodyExprs...)
					}
					fn := &Function{
						Params:    params,
						RestParam: restParam,
						Body:      body,
						Env:       env,
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
				restParam := ""
				if expr.List[1].IsList() {
					paramList := expr.List[1].List
					for i := 0; i < len(paramList); i++ {
						p := paramList[i]
						if p.IsSymbol() && p.Symbol == "." {
							// Rest parameter: next symbol is the rest param name
							if i+1 < len(paramList) && paramList[i+1].IsSymbol() {
								restParam = paramList[i+1].Symbol
							}
							break
						}
						if p.IsSymbol() {
							params = append(params, p.Symbol)
						}
					}
				}
				// Handle multi-expression body
				var body Value
				if len(expr.List) == 3 {
					body = expr.List[2]
				} else {
					bodyExprs := make([]Value, len(expr.List)-2+1)
					bodyExprs[0] = Sym("begin")
					copy(bodyExprs[1:], expr.List[2:])
					body = Lst(bodyExprs...)
				}
				return Value{
					Type: TypeFunc,
					Func: &Function{
						Params:    params,
						RestParam: restParam,
						Body:      body,
						Env:       env,
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
					// Propagate blocked status
					if result.Type == TypeBlocked {
						return result
					}
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
		
		// Bind regular parameters
		for i, param := range f.Params {
			if i < len(args) {
				newEnv.Set(param, args[i])
			} else {
				newEnv.Set(param, Nil())
			}
		}
		
		// Bind rest parameter if present
		if f.RestParam != "" {
			restArgs := make([]Value, 0)
			if len(args) > len(f.Params) {
				restArgs = args[len(f.Params):]
			}
			newEnv.Set(f.RestParam, Lst(restArgs...))
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

// eval - evaluate a data structure as code
func builtinEval(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Nil()
	}
	// Evaluate the argument in the global environment
	return ev.Eval(args[0], ev.GlobalEnv)
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
// String Operations
// ============================================================================

func builtinStringAppend(ev *Evaluator, args []Value, env *Env) Value {
	var sb strings.Builder
	for _, arg := range args {
		switch arg.Type {
		case TypeString:
			sb.WriteString(arg.Str)
		case TypeSymbol:
			sb.WriteString(arg.Symbol)
		case TypeNumber:
			if arg.Number == float64(int64(arg.Number)) {
				sb.WriteString(fmt.Sprintf("%d", int64(arg.Number)))
			} else {
				sb.WriteString(fmt.Sprintf("%g", arg.Number))
			}
		default:
			sb.WriteString(arg.String())
		}
	}
	return Str(sb.String())
}

func builtinSymbolToString(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Str("")
	}
	if args[0].Type == TypeSymbol {
		return Str(args[0].Symbol)
	}
	return Str(args[0].String())
}

func builtinStringToSymbol(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Sym("")
	}
	if args[0].Type == TypeString {
		return Sym(args[0].Str)
	}
	if args[0].Type == TypeSymbol {
		return args[0]
	}
	return Sym(args[0].String())
}

func builtinNumberToString(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Str("0")
	}
	if args[0].Type == TypeNumber {
		if args[0].Number == float64(int64(args[0].Number)) {
			return Str(fmt.Sprintf("%d", int64(args[0].Number)))
		}
		return Str(fmt.Sprintf("%g", args[0].Number))
	}
	return Str(args[0].String())
}

// ============================================================================
// Registry Builtins
// ============================================================================

func builtinRegistrySet(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Nil()
	}
	var name string
	if args[0].Type == TypeSymbol {
		name = args[0].Symbol
	} else if args[0].Type == TypeString {
		name = args[0].Str
	} else {
		return Nil()
	}
	ev.Registry[name] = args[1]
	return args[1]
}

func builtinRegistryGet(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 {
		return Nil()
	}
	var name string
	if args[0].Type == TypeSymbol {
		name = args[0].Symbol
	} else if args[0].Type == TypeString {
		name = args[0].Str
	} else {
		return Nil()
	}
	if v, ok := ev.Registry[name]; ok {
		return v
	}
	return Nil()
}

func builtinRegistryKeys(ev *Evaluator, args []Value, env *Env) Value {
	keys := make([]Value, 0, len(ev.Registry))
	for k := range ev.Registry {
		keys = append(keys, Sym(k))
	}
	return Lst(keys...)
}

func builtinRegistryHas(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 {
		return Bool(false)
	}
	var name string
	if args[0].Type == TypeSymbol {
		name = args[0].Symbol
	} else if args[0].Type == TypeString {
		name = args[0].Str
	} else {
		return Bool(false)
	}
	_, ok := ev.Registry[name]
	return Bool(ok)
}

func builtinRegistryDelete(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 {
		return Bool(false)
	}
	var name string
	if args[0].Type == TypeSymbol {
		name = args[0].Symbol
	} else if args[0].Type == TypeString {
		name = args[0].Str
	} else {
		return Bool(false)
	}
	if _, ok := ev.Registry[name]; ok {
		delete(ev.Registry, name)
		return Bool(true)
	}
	return Bool(false)
}

// ============================================================================
// Type Tagging Builtins
// ============================================================================

func builtinTag(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		return Nil()
	}
	var tagName string
	if args[0].Type == TypeSymbol {
		tagName = args[0].Symbol
	} else if args[0].Type == TypeString {
		tagName = args[0].Str
	} else {
		return Nil()
	}
	return Value{
		Type: TypeTagged,
		Tagged: &TaggedValue{
			Tag:   tagName,
			Value: args[1],
		},
	}
}

func builtinTagType(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeTagged {
		return Nil()
	}
	return Sym(args[0].Tagged.Tag)
}

func builtinTagValue(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 || args[0].Type != TypeTagged {
		return Nil()
	}
	return args[0].Tagged.Value
}

func builtinIsTagged(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 1 {
		return Bool(false)
	}
	return Bool(args[0].Type == TypeTagged)
}

func builtinTagIs(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 || args[0].Type != TypeTagged {
		return Bool(false)
	}
	var tagName string
	if args[1].Type == TypeSymbol {
		tagName = args[1].Symbol
	} else if args[1].Type == TypeString {
		tagName = args[1].Str
	} else {
		return Bool(false)
	}
	return Bool(args[0].Tagged.Tag == tagName)
}

// ============================================================================
// Symbol Generation
// ============================================================================

func builtinGensym(ev *Evaluator, args []Value, env *Env) Value {
	prefix := "g"
	if len(args) > 0 {
		if args[0].Type == TypeSymbol {
			prefix = args[0].Symbol
		} else if args[0].Type == TypeString {
			prefix = args[0].Str
		}
	}
	ev.GensymCount++
	return Sym(fmt.Sprintf("%s-%d", prefix, ev.GensymCount))
}

// ============================================================================
// Scheduler Builtins
// ============================================================================

// TypeActor for actor references
const TypeActor ValueType = 100

type ActorRef struct {
	Name string
}

func ActorVal(name string) Value {
	return Value{Type: TypeActor, Symbol: name}
}

func (v Value) IsActor() bool {
	return v.Type == TypeActor
}

// (spawn-actor name mailbox-size body)
// Creates a new actor with the given name, mailbox size, and initial code
func builtinSpawnActor(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "spawn-actor: need name, mailbox-size, body")
		return Nil()
	}
	
	var name string
	if args[0].Type == TypeSymbol {
		name = args[0].Symbol
	} else if args[0].Type == TypeString {
		name = args[0].Str
	} else {
		fmt.Fprintln(os.Stderr, "spawn-actor: name must be symbol or string")
		return Nil()
	}
	
	mailboxSize := 16
	if args[1].Type == TypeNumber {
		mailboxSize = int(args[1].Number)
	}
	
	// Create actor's own environment (inherits from global)
	actorEnv := NewEnv(ev.GlobalEnv)
	
	// The body is a thunk (code to execute)
	body := args[2]
	
	ev.Scheduler.AddActor(name, mailboxSize, actorEnv, body)
	
	return ActorVal(name)
}

// (self) - returns current actor's name
func builtinSelf(ev *Evaluator, args []Value, env *Env) Value {
	if ev.Scheduler.CurrentActor == "" {
		return Nil()
	}
	return Sym(ev.Scheduler.CurrentActor)
}

// (send-to! actor-name message)
// Sends a message to the named actor's mailbox
// Blocks if mailbox is full
func builtinSendTo(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "send-to!: need actor-name and message")
		return Nil()
	}
	
	var targetName string
	if args[0].Type == TypeSymbol {
		targetName = args[0].Symbol
	} else if args[0].Type == TypeString {
		targetName = args[0].Str
	} else if args[0].Type == TypeActor {
		targetName = args[0].Symbol
	} else {
		fmt.Fprintln(os.Stderr, "send-to!: target must be symbol, string, or actor ref")
		return Nil()
	}
	
	target := ev.Scheduler.GetActor(targetName)
	if target == nil {
		fmt.Fprintf(os.Stderr, "send-to!: unknown actor %s\n", targetName)
		return Nil()
	}
	
	message := args[1]
	
	if target.Mailbox.SendNow(message) {
		// Message sent successfully
		// If target was blocked on receive, unblock it
		if target.State == ActorBlocked && strings.HasPrefix(target.BlockedOn, "recv") {
			ev.Scheduler.UnblockActor(targetName)
		}
		return Sym("ok")
	} else {
		// Mailbox full, block sender
		if ev.Scheduler.CurrentActor != "" {
			ev.Scheduler.BlockActor(ev.Scheduler.CurrentActor, 
				fmt.Sprintf("send-to %s (full)", targetName))
		}
		return Blocked(BlockQueueFull)
	}
}

// (receive!) - receive from own mailbox, blocks if empty
func builtinReceive(ev *Evaluator, args []Value, env *Env) Value {
	if ev.Scheduler.CurrentActor == "" {
		fmt.Fprintln(os.Stderr, "receive!: no current actor")
		return Nil()
	}
	
	actor := ev.Scheduler.GetActor(ev.Scheduler.CurrentActor)
	if actor == nil {
		return Nil()
	}
	
	if msg, ok := actor.Mailbox.RecvNow(); ok {
		return msg
	} else {
		// Mailbox empty, block
		ev.Scheduler.BlockActor(ev.Scheduler.CurrentActor, "recv (empty)")
		return Blocked(BlockQueueEmpty)
	}
}

// (receive-now!) - non-blocking receive, returns 'empty if nothing
func builtinReceiveNow(ev *Evaluator, args []Value, env *Env) Value {
	if ev.Scheduler.CurrentActor == "" {
		fmt.Fprintln(os.Stderr, "receive-now!: no current actor")
		return Sym("empty")
	}
	
	actor := ev.Scheduler.GetActor(ev.Scheduler.CurrentActor)
	if actor == nil {
		return Sym("empty")
	}
	
	if msg, ok := actor.Mailbox.RecvNow(); ok {
		return msg
	}
	return Sym("empty")
}

// (mailbox-empty?) - check if own mailbox is empty
func builtinMailboxEmpty(ev *Evaluator, args []Value, env *Env) Value {
	if ev.Scheduler.CurrentActor == "" {
		return Bool(true)
	}
	actor := ev.Scheduler.GetActor(ev.Scheduler.CurrentActor)
	if actor == nil {
		return Bool(true)
	}
	return Bool(actor.Mailbox.IsEmpty())
}

// (mailbox-full? actor-name) - check if actor's mailbox is full
func builtinMailboxFull(ev *Evaluator, args []Value, env *Env) Value {
	var targetName string
	if len(args) > 0 {
		if args[0].Type == TypeSymbol {
			targetName = args[0].Symbol
		} else if args[0].Type == TypeString {
			targetName = args[0].Str
		}
	} else if ev.Scheduler.CurrentActor != "" {
		targetName = ev.Scheduler.CurrentActor
	} else {
		return Bool(false)
	}
	
	actor := ev.Scheduler.GetActor(targetName)
	if actor == nil {
		return Bool(false)
	}
	return Bool(actor.Mailbox.IsFull())
}

// (yield!) - voluntarily give up execution
func builtinYield(ev *Evaluator, args []Value, env *Env) Value {
	// This is a marker - the scheduler will handle it
	return Sym("yield")
}

// (done!) - mark current actor as finished
func builtinDone(ev *Evaluator, args []Value, env *Env) Value {
	if ev.Scheduler.CurrentActor != "" {
		ev.Scheduler.MarkDone(ev.Scheduler.CurrentActor)
	}
	return Sym("done")
}

// (run-scheduler max-steps) - run the scheduler
func builtinRunScheduler(ev *Evaluator, args []Value, env *Env) Value {
	maxSteps := int64(10000)
	if len(args) > 0 && args[0].Type == TypeNumber {
		maxSteps = int64(args[0].Number)
	}
	
	ev.Scheduler.MaxSteps = maxSteps
	ev.Scheduler.StepCount = 0
	
	for ev.Scheduler.StepCount < maxSteps {
		// Check termination conditions
		if ev.Scheduler.AllDone() {
			return Lst(Sym("completed"), Num(float64(ev.Scheduler.StepCount)))
		}
		if ev.Scheduler.IsDeadlocked() {
			// Return deadlock info
			blocked := make([]Value, 0)
			for name, actor := range ev.Scheduler.Actors {
				if actor.State == ActorBlocked {
					blocked = append(blocked, Lst(Sym(name), Str(actor.BlockedOn)))
				}
			}
			return Lst(Sym("deadlock"), Num(float64(ev.Scheduler.StepCount)), Lst(blocked...))
		}
		
		// Get next actor
		actor := ev.Scheduler.NextActor()
		if actor == nil {
			// No runnable actors but not deadlocked - all must be done
			return Lst(Sym("completed"), Num(float64(ev.Scheduler.StepCount)))
		}
		
		if ev.Scheduler.Trace {
			fmt.Printf("[%d] Running %s\n", ev.Scheduler.StepCount, actor.Name)
		}
		
		// Execute one step of actor's code
		if ev.Scheduler.Trace {
			fmt.Printf("    code: %s\n", actor.Code.String())
		}
		result := ev.Eval(actor.Code, actor.Env)
		actor.Result = result
		ev.Scheduler.StepCount++
		
		if ev.Scheduler.Trace {
			fmt.Printf("    result: %s\n", result.String())
		}
		
		// Check result
		if result.Type == TypeBlocked {
			// Already blocked by the operation
			if ev.Scheduler.Trace {
				fmt.Printf("    %s blocked: %s\n", actor.Name, actor.BlockedOn)
			}
		} else if result.Type == TypeSymbol && result.Symbol == "yield" {
			// Yielded voluntarily - stays runnable, re-run same code
			if ev.Scheduler.Trace {
				fmt.Printf("    %s yielded\n", actor.Name)
			}
		} else if result.Type == TypeSymbol && result.Symbol == "done" {
			// Actor finished
			ev.Scheduler.MarkDone(actor.Name)
			if ev.Scheduler.Trace {
				fmt.Printf("    %s done\n", actor.Name)
			}
		} else if result.IsList() && len(result.List) >= 2 {
			// Check for (next-state new-code) or (become new-code)
			if result.List[0].IsSymbol() && result.List[0].Symbol == "become" {
				// Change actor's code
				actor.Code = result.List[1]
				if ev.Scheduler.Trace {
					fmt.Printf("    %s become %s\n", actor.Name, result.List[1].String())
				}
			} else if result.List[0].IsSymbol() && result.List[0].Symbol == "continue" {
				// Update code and keep running
				actor.Code = result.List[1]
			}
		}
		
		// Try to unblock actors whose conditions may have changed
		ev.tryUnblockActors()
	}
	
	return Lst(Sym("max-steps"), Num(float64(ev.Scheduler.StepCount)))
}

// Try to unblock actors that can now proceed
func (ev *Evaluator) tryUnblockActors() {
	for name, actor := range ev.Scheduler.Actors {
		if actor.State != ActorBlocked {
			continue
		}
		
		if strings.HasPrefix(actor.BlockedOn, "recv") {
			// Blocked on receive - check if mailbox now has messages
			if !actor.Mailbox.IsEmpty() {
				ev.Scheduler.UnblockActor(name)
			}
		} else if strings.HasPrefix(actor.BlockedOn, "send-to ") {
			// Blocked on send - check if target mailbox has space
			parts := strings.Split(actor.BlockedOn, " ")
			if len(parts) >= 2 {
				targetName := parts[1]
				target := ev.Scheduler.GetActor(targetName)
				if target != nil && !target.Mailbox.IsFull() {
					ev.Scheduler.UnblockActor(name)
				}
			}
		}
	}
}

// (scheduler-status) - print scheduler state
func builtinSchedulerStatus(ev *Evaluator, args []Value, env *Env) Value {
	fmt.Print(ev.Scheduler.Status())
	return Nil()
}

// (set-trace! bool) - enable/disable execution tracing
func builtinSetTrace(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) > 0 {
		ev.Scheduler.Trace = args[0].IsTruthy()
	}
	return Bool(ev.Scheduler.Trace)
}

// (actor-state name) - get actor's current state
func builtinActorState(ev *Evaluator, args []Value, env *Env) Value {
	if len(args) == 0 {
		return Nil()
	}
	var name string
	if args[0].Type == TypeSymbol {
		name = args[0].Symbol
	} else if args[0].Type == TypeString {
		name = args[0].Str
	} else {
		return Nil()
	}
	
	actor := ev.Scheduler.GetActor(name)
	if actor == nil {
		return Nil()
	}
	
	state := "unknown"
	switch actor.State {
	case ActorRunnable:
		state = "runnable"
	case ActorBlocked:
		state = "blocked"
	case ActorDone:
		state = "done"
	}
	
	return Lst(
		Sym(state),
		Str(actor.BlockedOn),
		Num(float64(len(actor.Mailbox.Data))),
		Num(float64(actor.Mailbox.Capacity)),
	)
}

// (list-actors-sched) - list all actors in scheduler
func builtinListActorsSched(ev *Evaluator, args []Value, env *Env) Value {
	names := make([]Value, 0, len(ev.Scheduler.Actors))
	for name := range ev.Scheduler.Actors {
		names = append(names, Sym(name))
	}
	return Lst(names...)
}

// (reset-scheduler) - clear all actors and reset scheduler state
func builtinResetScheduler(ev *Evaluator, args []Value, env *Env) Value {
	ev.Scheduler = NewScheduler()
	return Sym("ok")
}

// ============================================================================
// REPL and File Execution
// ============================================================================

func countParens(s string) (int, int) {
	open := 0
	close := 0
	inString := false
	escaped := false
	for _, c := range s {
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == '(' {
			open++
		} else if c == ')' {
			close++
		}
	}
	return open, close
}

func runREPL(ev *Evaluator) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("BoundedLISP - Type (exit) to quit")
	fmt.Print("> ")

	var accum strings.Builder
	openCount := 0
	closeCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.TrimSpace(line) == "(exit)" && openCount == closeCount {
			break
		}

		accum.WriteString(line)
		accum.WriteString("\n")
		
		o, c := countParens(line)
		openCount += o
		closeCount += c

		// If parens are balanced and we have something, evaluate
		if openCount > 0 && openCount == closeCount {
			input := accum.String()
			accum.Reset()
			openCount = 0
			closeCount = 0

			parser := NewParser(input)
			exprs := parser.Parse()

			for _, expr := range exprs {
				result := ev.Eval(expr, nil)
				if result.Type != TypeNil {
					fmt.Println(result.String())
				}
			}
			fmt.Print("> ")
		} else if openCount > closeCount {
			// Need more input
			fmt.Print("  ")
		} else {
			// Unbalanced or empty line
			fmt.Print("> ")
		}
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
