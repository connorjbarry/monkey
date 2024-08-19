package parser

import (
	"fmt"
	"strconv"

	"github.com/connorjbarry/monkey/interpreter/lexer"

	"github.com/connorjbarry/monkey/interpreter/token"

	"github.com/connorjbarry/monkey/interpreter/ast"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // >, <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X, !X
	CALL        // func()
	INDEX       // []
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	currT token.Token
	peekT token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// read two tokens, sets currT and peekT
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix((token.IDENTIFER), p.parseIdentifier)
	p.registerPrefix((token.INT), p.parseIntegerLiteral)
	p.registerPrefix((token.BANG), p.parsePrefixExpression)
	p.registerPrefix((token.MINUS), p.parsePrefixExpression)
	p.registerPrefix((token.TRUE), p.parseBoolean)
	p.registerPrefix((token.FALSE), p.parseBoolean)
	p.registerPrefix((token.LPAREN), p.parseGroupedExpression)
	p.registerPrefix((token.IF), p.parseIfExpression)
	p.registerPrefix((token.FUNCTION), p.parseFunctionLiteral)
	p.registerPrefix((token.STRING), p.parseStringLiteral)
	p.registerPrefix((token.LBRACKET), p.parseArrayLiteral)
	p.registerPrefix((token.LBRACE), p.parseHashLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix((token.PLUS), p.parseInfixExpression)
	p.registerInfix((token.MINUS), p.parseInfixExpression)
	p.registerInfix((token.ASTERISK), p.parseInfixExpression)
	p.registerInfix((token.SLASH), p.parseInfixExpression)
	p.registerInfix((token.EQ), p.parseInfixExpression)
	p.registerInfix((token.NEQ), p.parseInfixExpression)
	p.registerInfix((token.LT), p.parseInfixExpression)
	p.registerInfix((token.GT), p.parseInfixExpression)
	p.registerInfix((token.LPAREN), p.parseCallExpression)
	p.registerInfix((token.LBRACKET), p.parseIndexExpression)

	return p
}

func (p *Parser) nextToken() {
	p.currT = p.peekT
	p.peekT = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currT.Type != token.EOF {
		stmt := p.parseStatment()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseStatment() ast.Statement {
	switch p.currT.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatment()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currT}

	if !p.expectPeek(token.IDENTIFER) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.currT, Value: p.currT.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	for !p.currTIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currT}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for !p.currTIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatment() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currT}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.currT.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currT.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecendence() {
		infix := p.infixParseFns[p.peekT.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currT, Value: p.currT.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.currT}

	val, err := strconv.ParseInt(p.currT.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currT.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = val
	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.currT, Value: p.currTIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{
		Token:    p.currT,
		Operator: p.currT.Literal,
	}
	p.nextToken()

	exp.Right = p.parseExpression(PREFIX)

	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{
		Token:    p.currT,
		Operator: p.currT.Literal,
		Left:     left,
	}

	prec := p.currPrecendence()
	p.nextToken()

	exp.Right = p.parseExpression(prec)

	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.currT}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	exp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		exp.Alternative = p.parseBlockStatement()
	}

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currT}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.currTIs(token.RBRACE) && !p.currTIs(token.EOF) {
		stmt := p.parseStatment()

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.currT}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Params = p.parseFunctionParams()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParams() []*ast.Identifier {
	idents := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return idents
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.currT, Value: p.currT.Literal}
	idents = append(idents, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		ident := &ast.Identifier{Token: p.currT, Value: p.currT.Literal}
		idents = append(idents, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return idents
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.currT, Func: function}
	exp.Args = p.parseExpressionList(token.RPAREN)
	return exp
}

// Deprecated: use parseExpressionList instead
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.currT, Value: p.currT.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.currT}

	arr.Elements = p.parseExpressionList(token.RBRACKET)

	return arr
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.currT}

	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()

		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.currT, Left: left}

	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) currTIs(t token.TokenType) bool {
	return p.currT.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekT.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekPrecendence() int {
	if p, ok := precedences[p.peekT.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) currPrecendence() int {
	if p, ok := precedences[p.currT.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekT.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function found for %s", t)
	p.errors = append(p.errors, msg)
}
