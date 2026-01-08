//go:build !solution

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Evaluator is a stack-based evaluator for the Forth language.
type Evaluator struct {
	stack            []int
	customOperations map[string][]string
	basicOperations  map[string]bool
}

// NewEvaluator creates a new evaluator for processing Forth code.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		customOperations: make(map[string][]string),
		basicOperations: map[string]bool{
			"dup":  true,
			"over": true,
			"drop": true,
			"swap": true,
			"+":    true,
			"-":    true,
			"*":    true,
			"/":    true,
		},
	}
}

func BinaryEvaluation(operation string, one int, two int) (int, error) {
	switch operation {
	case "+":
		return one + two, nil
	case "-":
		return one - two, nil
	case "*":
		return one * two, nil
	case "/":
		if two == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return one / two, nil
	default:
		panic("unsupported binary operation")
	}
}

func (e *Evaluator) ProcessOperation(operation string) error {
	switch operation {
	case "+", "-", "*", "/":
		if len(e.stack) < 2 {
			return fmt.Errorf("not enough values on stack for %s operation", operation)
		}
		val, err := BinaryEvaluation(operation, e.stack[len(e.stack)-2], e.stack[len(e.stack)-1])
		if err != nil {
			return err
		}
		e.stack = e.stack[:len(e.stack)-2]
		e.stack = append(e.stack, val)
	case "dup":
		if len(e.stack) == 0 {
			return fmt.Errorf("stack is empty for dup operation")
		}
		e.stack = append(e.stack, e.stack[len(e.stack)-1])
	case "over":
		if len(e.stack) < 2 {
			return fmt.Errorf("not enough values on stack for over operation")
		}
		e.stack = append(e.stack, e.stack[len(e.stack)-2])
	case "drop":
		if len(e.stack) == 0 {
			return fmt.Errorf("stack is empty for drop operation")
		}
		e.stack = e.stack[:len(e.stack)-1]
	case "swap":
		if len(e.stack) < 2 {
			return fmt.Errorf("not enough values on stack for swap operation")
		}
		e.stack[len(e.stack)-1], e.stack[len(e.stack)-2] = e.stack[len(e.stack)-2], e.stack[len(e.stack)-1]
	default:
		val, err := strconv.Atoi(operation)
		if err != nil {
			return fmt.Errorf("unsupported stack operation: %s", operation)
		}
		e.stack = append(e.stack, val)
	}
	return nil
}

func (e *Evaluator) AddOperation(name string, commands []string) {
	for i := 0; i < len(commands); i++ {
		commands[i] = strings.ToLower(commands[i])
	}
	for i := 0; i < len(commands); i++ {
		if _, ok := e.customOperations[commands[i]]; ok {
			commandsSuffix := commands[i+1:]
			length := len(e.customOperations[commands[i]])
			commands = append(commands[:i], e.customOperations[commands[i]]...)
			commands = append(commands, commandsSuffix...)
			i += length - 1
		}
	}
	e.customOperations[strings.ToLower(name)] = commands
}

func (e *Evaluator) Process(row string) ([]int, error) {
	parts := strings.Fields(row)
	var i int
	for ; i < len(parts); i++ {
		part := strings.ToLower(parts[i])

		// If it's a new word definition (starts with ':')
		if part == ":" {
			wordName := strings.ToLower(parts[i+1])

			if _, err := strconv.Atoi(wordName); err == nil {
				return nil, fmt.Errorf("invalid word definition, missing word name")
			}

			var commands []string
			i += 2

			for i < len(parts) && parts[i] != ";" {
				commands = append(commands, parts[i])
				i++
			}
			i++

			e.AddOperation(wordName, commands)
			continue
		}

		// Check if the part is an operation (either basic or custom)
		if commands, ok := e.customOperations[part]; ok {
			// Execute the custom operation
			for _, cmd := range commands {
				err := e.ProcessOperation(cmd)
				if err != nil {
					return nil, err
				}
			}
		} else {
			err := e.ProcessOperation(part)
			if err != nil {
				return nil, err
			}
		}
	}
	return e.stack, nil
}
