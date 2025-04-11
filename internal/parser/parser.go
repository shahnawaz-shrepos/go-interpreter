package parser

import (
	"fmt"

	"github.com/go-interpreter/internal/ast"
	"github.com/go-interpreter/internal/errors"
	"github.com/go-interpreter/internal/token"
)

// Parser is responsible for processing a sequence of tokens and
// converting them into a meaningful structure, typically an
// Abstract Syntax Tree (AST). It keeps track of the tokens to
// be parsed and the current position within the token stream.
type Parser struct {
	Tokens  []token.Token
	Current int
}

// NewParser creates a new instance of the Parser struct with the provided
// slice of tokens. The tokens are used as the input for the parser to
// process and generate the corresponding syntax tree or perform other
// parsing operations.
func NewParser(tokens []token.Token) Parser {
	return Parser{
		Tokens: tokens,
	}
}

// Parse parses the input source code into a slice of abstract syntax tree (AST) statements.
// It continues parsing until the end of the input is reached or an error occurs.
// Returns the parsed statements or an error if parsing fails.
func (parser *Parser) Parse() ([]ast.Stmt, error) {
	var statements []ast.Stmt
	for !parser.isAtEnd() {
		statement, err := parser.statement()
		if err != nil {
			fmt.Print(fmt.Errorf("%v", err))
		}
		statements = append(statements, statement)
	}

	return statements, nil
}

// statement parses a statement from the input tokens and returns it as an
// abstract syntax tree (AST) node. It first checks if the statement is a
// "print" statement and delegates parsing to the printStatement method if so.
// If not, it assumes the statement is an expression statement and parses it
// accordingly. Returns an error if parsing fails at any stage.
func (parser *Parser) statement() (ast.Stmt, error) {
	if parser.match(token.PRINT) {
		printStatment, err := parser.printStatement()
		if err != nil {
			return nil, err
		}
		return printStatment, nil
	}
	// It must be an expression statement
	expressionStmt, err := parser.expression()
	if err != nil {
		return nil, err
	}
	return ast.ExpressionStmt{Expression: expressionStmt}, nil
}

// PrintStatement parses a print statement in the source code.
// It expects an expression followed by a semicolon (';').
// Returns an abstract syntax tree (AST) node representing the print statement
// or an error if parsing fails.
func (parser *Parser) printStatement() (ast.Stmt, error) {
	value, err := parser.expression()
	if err != nil {
		return nil, err
	}
	_, err = parser.consume(token.SEMICOLON, "Expect ';' after value.")
	if err != nil {
		return nil, err
	}
	return ast.PrintStmt{Expression: value}, nil

}

// expression parses and returns an expression from the input source.
// It delegates the parsing to the equality method and returns the resulting
// abstract syntax tree (AST) expression or an error if parsing fails.
func (parser *Parser) expression() (ast.Expr, error) {
	eq, err := parser.equality()
	if err != nil {
		return nil, err
	}
	return eq, nil
}

// equality parses and constructs an equality expression in the abstract syntax tree (AST).
// It first parses a comparison expression and then checks for equality operators
// (!= or ==). If an equality operator is found, it creates a binary expression
// node with the operator and the right-hand side expression.
// Returns the constructed expression or an error if parsing fails.
func (parser *Parser) equality() (ast.Expr, error) {
	expr, err := parser.comparison()
	if err != nil {
		return nil, err
	}
	for parser.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := parser.previous()
		right, err := parser.comparison()
		if err != nil {
			return nil, err
		}
		expr = ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// comparison parses and constructs a comparison expression in the form of a binary
// operation. It first parses a term expression and then checks for comparison
// operators such as GREATER, GREATER_EQUAL, LESS, and LESS_EQUAL. If a comparison
// operator is found, it continues parsing the right-hand side term and constructs
// a binary expression node. The process repeats for chained comparisons.
// Returns the constructed expression or an error if parsing fails.
func (parser *Parser) comparison() (ast.Expr, error) {
	expr, err := parser.term()
	if err != nil {
		return nil, err
	}
	for parser.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := parser.previous()
		right, err := parser.term()
		if err != nil {
			return nil, err
		}
		expr = ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// term parses and returns an expression representing a term in the grammar.
// A term is defined as a sequence of factors combined using addition or subtraction
// operators. The method first parses a factor and then checks for any subsequent
// addition or subtraction operators, combining them into a binary expression tree.
func (parser *Parser) term() (ast.Expr, error) {
	expr, err := parser.factor()
	if err != nil {
		return nil, err
	}
	for parser.match(token.MINUS, token.PLUS) {
		operator := parser.previous()
		right, err := parser.factor()
		if err != nil {
			return nil, err
		}
		expr = ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// factor parses and returns an expression representing a binary operation
// involving multiplication (*) or division (/). It first parses a unary
// expression and then checks for subsequent binary operations with the
// specified operators. If such operations are found, it constructs a
// Binary AST node with the left operand, operator, and right operand.
// Returns the resulting expression or an error if parsing fails.
func (parser *Parser) factor() (ast.Expr, error) {
	expr, err := parser.unary()
	if err != nil {
		return nil, err
	}
	for parser.match(token.SLASH, token.STAR) {
		operator := parser.previous()
		right, err := parser.unary()
		if err != nil {
			return nil, err
		}
		expr = ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// unary parses a unary expression in the source code. A unary expression
// consists of an operator (e.g., '!' or '-') followed by a single operand.
// If the current token matches a unary operator, this function recursively
// parses the operand and constructs an abstract syntax tree (AST) node
// representing the unary expression. If no unary operator is matched, it
// delegates parsing to the primary expression parser.
//
// Returns an AST expression node representing the unary expression or
// primary expression, along with any error encountered during parsing.
func (parser *Parser) unary() (ast.Expr, error) {
	if parser.match(token.BANG, token.MINUS) {
		operator := parser.previous()
		right, err := parser.unary()
		if err != nil {
			return nil, err
		}
		return ast.Unary{Operator: operator, Right: right}, nil
	}
	primary, err := parser.primary()
	if err != nil {
		return nil, err
	}
	return primary, nil
}

// primary parses a primary expression in the source code and returns an
// abstract syntax tree (AST) representation of the expression or an error
// if parsing fails. A primary expression can be a literal value (e.g., true,
// false, nil, numbers, or strings), a grouped expression enclosed in
// parentheses, or an unexpected token.
//
// The function uses a switch statement to match the current token against
// various cases, such as boolean literals, nil, numeric or string literals,
// and grouped expressions. If a grouped expression is encountered, it
// recursively parses the inner expression and ensures that it is properly
// closed with a right parenthesis.
//
// If an unexpected token is encountered, the function returns a parser error
// with details about the token and its location in the source code.
func (parser *Parser) primary() (ast.Expr, error) {
	switch {
	case parser.match(token.FALSE):
		return ast.Literal{Value: false}, nil
	case parser.match(token.TRUE):
		return ast.Literal{Value: true}, nil
	case parser.match(token.NIL):
		return ast.Literal{Value: nil}, nil
	case parser.match(token.NUMBER, token.STRING):
		return ast.Literal{Value: parser.previous().Literal}, nil
	case parser.match(token.LEFT_PAREN):
		expr, e := parser.expression()
		if e != nil {
			fmt.Println(fmt.Errorf("%v", e))
		}
		_, err := parser.consume(token.RIGHT_PAREN, "Expect ')' after expression.")
		if err != nil {
			return nil, err
		}
		return ast.Grouping{Expression: expr}, nil
	default:
		// We probaby don't want to panic here because we are syncing the parser
		// We will catch it in parser.match(token.LEFT_PAREN) and report it back to
		// the stdout
		peek := parser.peek()
		return nil, errors.ExecutionError{
			Type:    errors.PARSER_ERROR,
			Line:    peek.Line,
			Where:   peek.Char,
			Message: fmt.Sprintf("Unexpected token '%v'", peek.Lexeme),
		}

	}
}

// Comparison parses a comparison expression from the list of tokens.
// It returns the root node of the abstract syntax tree.
func (parser *Parser) match(types ...token.TokenType) bool {
	for _, tokenType := range types {
		if parser.check(tokenType) {
			parser.advance()
			return true
		}
	}
	return false
}

// This is a helper function to check if the current token is of the given type.
func (parser *Parser) check(type_ token.TokenType) bool {
	if parser.isAtEnd() {
		return false
	}
	return parser.peek().Type == type_
}

// This is a helper function to advance the parser to the next token.
func (parser *Parser) advance() token.Token {
	if !parser.isAtEnd() {
		parser.Current++
	}
	return parser.previous()
}

// This is a helper function to match the type at the end of the list of tokens.
// If it is at the end, we return the null character.
func (parser *Parser) isAtEnd() bool {
	return parser.peek().Type == token.EOF
}

// This is a helper function to peek at the end of the string and return it.
func (parser *Parser) peek() token.Token {
	return parser.Tokens[parser.Current]
}

// This is a helper function to return the previous token.
func (parser *Parser) previous() token.Token {
	return parser.Tokens[parser.Current-1]
}

// This function consumer or otherwise it throws an error
// It consumes the token if it is of the given type.
// If it is not, it throws an error with the given message.
// The caller can handle the error and decide what to do with it.
func (parser *Parser) consume(type_ token.TokenType, message string) (token.Token, error) {
	if parser.check(type_) {
		return parser.advance(), nil
	}
	return token.Token{}, errors.ExecutionError{
		Type:    errors.PARSER_ERROR,
		Line:    parser.peek().Line,
		Where:   parser.peek().Char,
		Message: message,
	}
}
