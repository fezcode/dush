package ast

import (
	"bytes"
	"dush/internal/parser/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// LetStatement: let @x = 5
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *VarExpression
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " @")
	out.WriteString(ls.Name.Name)
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ConstStatement: const @x = 5
type ConstStatement struct {
	Token token.Token
	Name  *VarExpression
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	var out bytes.Buffer
	out.WriteString("const @")
	out.WriteString(cs.Name.Name)
	out.WriteString(" = ")
	if cs.Value != nil {
		out.WriteString(cs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// PubStatement: pub @x = 5, pub const @x = 5, or pub @x
type PubStatement struct {
	Token   token.Token
	Name    *VarExpression
	Value   Expression // nil if just marking existing var as pub
	IsConst bool
}

func (ps *PubStatement) statementNode()       {}
func (ps *PubStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PubStatement) String() string {
	var out bytes.Buffer
	out.WriteString("pub ")
	if ps.IsConst {
		out.WriteString("const ")
	}
	out.WriteString("@")
	out.WriteString(ps.Name.Name)
	if ps.Value != nil {
		out.WriteString(" = ")
		out.WriteString(ps.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ReturnStatement: return 5;
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement: x + 5; (wrapper)
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement: { ... }
type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Identifier: x (bare word — always a command name or literal string in the new system)
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// VarExpression: @x (variable access)
type VarExpression struct {
	Token token.Token // the AT token
	Name  string      // variable name without @
}

func (ve *VarExpression) expressionNode()      {}
func (ve *VarExpression) TokenLiteral() string { return ve.Token.Literal }
func (ve *VarExpression) String() string       { return "@" + ve.Name }

// MethodCallExpression: @x.upper() or @x.slice(0, 5)
type MethodCallExpression struct {
	Token     token.Token  // the DOT token
	Object    Expression   // the expression before the dot
	Method    string       // method name
	Arguments []Expression // method arguments
}

func (mc *MethodCallExpression) expressionNode()      {}
func (mc *MethodCallExpression) TokenLiteral() string { return mc.Token.Literal }
func (mc *MethodCallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Method)
	out.WriteString("(")
	args := []string{}
	for _, a := range mc.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// InterpolatedStringExpression: "hello @name" → parts: [StringLiteral("hello "), VarExpression("name")]
type InterpolatedStringExpression struct {
	Token token.Token   // the STRING token
	Parts []Expression  // mix of StringLiteral and VarExpression/MethodCallExpression
}

func (is *InterpolatedStringExpression) expressionNode()      {}
func (is *InterpolatedStringExpression) TokenLiteral() string { return is.Token.Literal }
func (is *InterpolatedStringExpression) String() string {
	var out bytes.Buffer
	out.WriteString("\"")
	for _, p := range is.Parts {
		out.WriteString(p.String())
	}
	out.WriteString("\"")
	return out.String()
}

// IntegerLiteral: 5
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral: 3.14
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral: "hello"
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Value }

// BooleanLiteral: true / false
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// PrefixExpression: -5, !true
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression: 5 + 5
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// IfExpression: if (x) { ... } else { ... }
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// CallExpression: add(1, 2)
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// ProcStatement: proc name(@x, @y) { ... }
type ProcStatement struct {
	Token      token.Token // The 'proc' token
	Name       *Identifier
	Parameters []*VarExpression
	Body       *BlockStatement
}

func (ps *ProcStatement) statementNode()       {}
func (ps *ProcStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *ProcStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ps.TokenLiteral() + " ")
	out.WriteString(ps.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range ps.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(ps.Body.String())
	return out.String()
}

// ProcLiteral: proc(@x, @y) { ... } (anonymous function)
type ProcLiteral struct {
	Token      token.Token // The 'proc' token
	Parameters []*VarExpression
	Body       *BlockStatement
}

func (pl *ProcLiteral) expressionNode()      {}
func (pl *ProcLiteral) TokenLiteral() string { return pl.Token.Literal }
func (pl *ProcLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("proc(")
	params := []string{}
	for _, p := range pl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(pl.Body.String())
	return out.String()
}

// LoopStatement: loop (@x < 10) OR loop (@x : items)
type LoopStatement struct {
	Token     token.Token    // LOOP
	Condition Expression     // Not nil for conditional loops
	Iterator  *VarExpression // Not nil for iterator loops
	Source    Expression     // Not nil for iterator loops
	Body      *BlockStatement
}

func (ls *LoopStatement) statementNode()       {}
func (ls *LoopStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LoopStatement) String() string {
	var out bytes.Buffer
	out.WriteString("loop (")
	if ls.Condition != nil {
		out.WriteString(ls.Condition.String())
	} else {
		out.WriteString(ls.Iterator.String() + " : " + ls.Source.String())
	}
	out.WriteString(") ")
	out.WriteString(ls.Body.String())
	return out.String()
}

// MatchExpression: match (@x) { case 1 { ... } case _ { ... } }
type MatchExpression struct {
	Token   token.Token // The 'match' token
	Subject Expression
	Cases   []*MatchCase
}

type MatchCase struct {
	Token      token.Token // The 'case' token
	Value      Expression  // nil means wildcard (_)
	IsDefault  bool        // true for case _
	Body       *BlockStatement
}

func (me *MatchExpression) expressionNode()      {}
func (me *MatchExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MatchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("match (")
	out.WriteString(me.Subject.String())
	out.WriteString(") { ")
	for _, c := range me.Cases {
		out.WriteString(c.String())
		out.WriteString(" ")
	}
	out.WriteString("}")
	return out.String()
}

func (mc *MatchCase) String() string {
	var out bytes.Buffer
	out.WriteString("case ")
	if mc.IsDefault {
		out.WriteString("_")
	} else {
		out.WriteString(mc.Value.String())
	}
	out.WriteString(" ")
	out.WriteString(mc.Body.String())
	return out.String()
}

// CommandExpression: git status -m "fix"
type CommandExpression struct {
	Token token.Token  // The first word (command name)
	Name  string       // "git"
	Args  []Expression // Parsed arguments
}

func (ce *CommandExpression) expressionNode()      {}
func (ce *CommandExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CommandExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(CMD: ")
	out.WriteString(ce.Name)
	for _, arg := range ce.Args {
		out.WriteString(" " + arg.String())
	}
	out.WriteString(")")
	return out.String()
}

// WithExpression: with (@ENV="val") { ... }
type WithExpression struct {
	Token        token.Token // The 'with' token
	EnvOverrides map[string]Expression
	Body         *BlockStatement
}

func (we *WithExpression) expressionNode()      {}
func (we *WithExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WithExpression) String() string {
	var out bytes.Buffer
	out.WriteString("with (")

	pairs := []string{}
	for k, v := range we.EnvOverrides {
		pairs = append(pairs, "@"+k+"="+v.String())
	}
	out.WriteString(strings.Join(pairs, ", "))

	out.WriteString(") ")
	out.WriteString(we.Body.String())
	return out.String()
}

// BackgroundExpression: command &
type BackgroundExpression struct {
	Token      token.Token // The '&' token
	Expression Expression  // The command to run in background
}

func (be *BackgroundExpression) expressionNode()      {}
func (be *BackgroundExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BackgroundExpression) String() string {
	return be.Expression.String() + " &"
}
