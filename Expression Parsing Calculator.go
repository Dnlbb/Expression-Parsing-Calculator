package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
    "unicode"
)

type Tag int

const (
    ERROR Tag = 1 << iota
    NUMBER
    VAR
    PLUS
    MINUS
    MUL
    DIV
    LPAREN
    RPAREN
)

type Lexem struct {
    Tag   Tag
    Image string
}
var error_when_using_lexer bool = true

func main() {
    if len(os.Args) < 2 {
        return
    }
    expr := os.Args[1]
    tokensChan = make(chan Lexem, 100)
    varValues = make(map[string]int)
    go lexer(expr, tokensChan)
    getNextToken()
   
    if result, err := parseExpr(); err == nil && error_when_using_lexer {
    fmt.Println(result)
} else {
    fmt.Println("error")
}
}

var currentToken Lexem
var tokensChan chan Lexem
var varValues map[string]int

func parseExpr() (int, error) {
    if value, err := parseTerm(); err == nil {
        return parseExprPrime(value)
    } else {
        return 0, err
    }
}

func parseExprPrime(inheritedValue int) (int, error) {
    if currentToken.Tag == PLUS {
        getNextToken()
        if value, err := parseTerm(); err == nil {
            return parseExprPrime(inheritedValue + value)
        } else {
            return 0, err
        }
    } else if currentToken.Tag == MINUS {
        getNextToken()
        if value, err := parseTerm(); err == nil {
            return parseExprPrime(inheritedValue - value)
        } else {
            return 0, err
        }
    }
    return inheritedValue, nil
}

func parseTerm() (int, error) {
    if value, err := parseFactor(); err == nil {
        return parseTermPrime(value)
    } else {
        return 0, err
    }
}

func parseTermPrime(inheritedValue int) (int, error) {
    if currentToken.Tag == MUL {
        getNextToken()
        if value, err := parseFactor(); err == nil {
            return parseTermPrime(inheritedValue * value)
        } else {
            return 0, err
        }
    } else if currentToken.Tag == DIV {
        getNextToken()
        if value, err := parseFactor(); err == nil {
            if value == 0 {
                return 0, nil
            }
            return parseTermPrime(inheritedValue / value)
        } else {
            return 0, err
        }
    }
    return inheritedValue, nil
}

func parseFactor() (int, error) {
    switch currentToken.Tag {
    case NUMBER:
        value, _ := strconv.Atoi(currentToken.Image)
        getNextToken()
        return value, nil
    case VAR:
        name := currentToken.Image
        if value, exists := varValues[name]; !exists {
            reader := bufio.NewReader(os.Stdin)
            input, _ := reader.ReadString('\n')
            if val, err := strconv.Atoi(strings.TrimSpace(input)); err == nil {
                varValues[name] = val
                getNextToken()
                return val, nil
            } else {
                return 0, fmt.Errorf("error")
            }
        } else {
            getNextToken()
            return value, nil
        }
    case LPAREN:
        getNextToken()
        if value, err := parseExpr(); err == nil {
            if currentToken.Tag == RPAREN {
                getNextToken()
                return value, nil
            }
            return 0, fmt.Errorf("error")
        } else {
            return 0, err
        }
    case MINUS:
        getNextToken()
        if value, err := parseFactor(); err == nil {
            return -value, nil
        } else {
            return 0, err
        }
    default:
        return 0, fmt.Errorf("error")
    }
}


func getNextToken() {
    currentToken = <-tokensChan
}


func lexer( expr string, tokens chan<- Lexem)  {
    n := len(expr)
    var lastTokenType Tag = ERROR 
    leftParenCount, rightParenCount := 0, 0
    for i := 0; i < n; {
        switch c := expr[i]; {
        case unicode.IsSpace(rune(c)):
            i++
            continue
        case unicode.IsDigit(rune(c)):
            start := i
            for i < n && unicode.IsDigit(rune(expr[i])) {
                i++
            }
            if lastTokenType == NUMBER || lastTokenType == VAR{
                error_when_using_lexer = false
            }
            tokens <- Lexem{Tag: NUMBER, Image: expr[start:i]}
            lastTokenType = NUMBER
        case unicode.IsLetter(rune(c)):
            start := i
            for i < n && (unicode.IsLetter(rune(expr[i])) || unicode.IsDigit(rune(expr[i]))) {
                i++
            }
            if lastTokenType == VAR || lastTokenType == NUMBER{
                error_when_using_lexer = false
            }
            lastTokenType = VAR
            tokens <- Lexem{Tag: VAR, Image: expr[start:i]}
        case c == '+':
            tokens <- Lexem{Tag: PLUS, Image: "+"}
            i++
            lastTokenType = PLUS
        case c == '-':
            tokens <- Lexem{Tag: MINUS, Image: "-"}
            i++
            lastTokenType = MINUS
        case c == '*':
            tokens <- Lexem{Tag: MUL, Image: "*"}
            i++
            lastTokenType = MUL
        case c == '/':
            tokens <- Lexem{Tag: DIV, Image: "/"}
            i++
            lastTokenType = DIV
        case c == '(':
            leftParenCount++
            tokens <- Lexem{Tag: LPAREN, Image: "("}
            i++
            lastTokenType =LPAREN
        case c == ')':
            rightParenCount++
            tokens <- Lexem{Tag: RPAREN, Image: ")"}
            i++
            lastTokenType = RPAREN
        default:
            tokens <- Lexem{Tag: ERROR, Image: string(c)}
            i++
            lastTokenType = ERROR
            error_when_using_lexer = false
        }
    }
    if leftParenCount != rightParenCount {
        error_when_using_lexer = false
    }
    close(tokens)
}
