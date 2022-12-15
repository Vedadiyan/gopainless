package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	INVALID_COMMAND_LINE_ARGUMENT ParserError = ParserError("invalid command line argument")
	COMMAND_GROUP_NOT_FOUND       ParserError = ParserError("command group not found")
)

type ParserError string

func (parserError ParserError) Error() string {
	return string(parserError)
}

type Token struct {
	must     map[string]*string
	optional map[string]string
	flags    map[string]bool
	help     map[string]string
}

type Command struct {
	commands map[string]Token
	help     map[string]string
}

func New() *Command {
	command := &Command{}
	command.commands = make(map[string]Token)
	command.help = make(map[string]string)
	return command
}

func (token Token) GetMust(command string) *string {
	return token.must[command]
}

func (token Token) GetOptional(command string) string {
	return token.optional[command]
}

func (token Token) GetFlag(command string) bool {
	return token.flags[command]
}

func (token Token) Help() string {
	longest := 0
	sortedKeys := make([]string, 0)
	for key, value := range token.help {
		_ = value
		len := len(key)
		if len > longest {
			longest = len
		}
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i] < sortedKeys[j]
	})
	buffer := bytes.NewBufferString("")
	for _, key := range sortedKeys {
		_, isMust := token.must[key]
		_, isOptional := token.must[key]
		if isMust || isOptional {
			buffer.WriteString("-")
			buffer.WriteString(key)
			for i := 0; i < (longest-len(key)-1)+10; i++ {
				buffer.WriteString(" ")
			}
		} else if _, ok := token.flags[key]; ok {
			buffer.WriteString("--")
			buffer.WriteString(key)
			for i := 0; i < (longest-len(key)-2)+10; i++ {
				buffer.WriteString(" ")
			}
		} else {
			panic("Unknown Case")
		}
		buffer.WriteString(token.help[key])
		buffer.WriteString("\r\n")
	}
	return buffer.String()
}

func (token *Token) RegisterCommand(cmd string, help string, defaultValue *string) *Token {
	if defaultValue != nil {
		token.optional[cmd] = *defaultValue
	} else {
		token.must[cmd] = nil
	}
	token.help[cmd] = help
	return token
}
func (token *Token) RegisterFlag(cmd string, help string) *Token {
	token.flags[cmd] = false
	token.help[cmd] = help
	return token
}
func (command *Command) RegisterGroup(group string, help string) *Token {
	token := Token{}
	token.must = make(map[string]*string)
	token.optional = make(map[string]string)
	token.flags = make(map[string]bool)
	token.help = make(map[string]string)
	command.commands[group] = token
	command.help[group] = help
	return &token
}
func (command *Command) Parse() (string, *Token, error) {
	commands := make(map[string]string)
	flags := make(map[string]bool)
	var group string
	var prev *string
	for i := 1; i < len(os.Args); i++ {
		val := os.Args[i]
		if strings.HasPrefix(val, "--") {
			_val := strings.TrimPrefix(val, "--")
			flags[_val] = true
			continue
		}
		if strings.HasPrefix(val, "-") {
			_val := strings.TrimPrefix(val, "-")
			prev = &_val
			continue
		}
		if i == 1 {
			group = val
			continue
		}
		if prev == nil {
			return "", nil, INVALID_COMMAND_LINE_ARGUMENT
		}
		commands[*prev] = val
		prev = nil
	}
	value, ok := command.commands[group]
	if !ok {
		return "", nil, COMMAND_GROUP_NOT_FOUND
	}
	errs := make([]string, 0)
	for key, val := range value.must {
		_ = val
		_value, ok := commands[key]
		if !ok {
			errs = append(errs, fmt.Sprintf("-%s is missing", key))
			continue
		}
		value.must[key] = &_value
	}
	for key, val := range value.optional {
		_ = val
		value.optional[key] = commands[key]
	}
	for key, val := range flags {
		_ = val
		_, ok := value.flags[key]
		if ok {
			value.flags[key] = true
		}
	}
	if len(errs) != 0 {
		return group, &value, errors.New(strings.Join(errs, "\r\n"))
	}
	return group, &value, nil
}

func (command Command) Help() string {
	longest := 0
	sortedKeys := make([]string, 0)
	for key, value := range command.help {
		_ = value
		len := len(key)
		if len > longest {
			longest = len
		}
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i] < sortedKeys[j]
	})
	buffer := bytes.NewBufferString("")
	for _, key := range sortedKeys {
		buffer.WriteString(key)
		for i := 0; i < (longest-len(key))+10; i++ {
			buffer.WriteString(" ")
		}
		buffer.WriteString(command.help[key])
		buffer.WriteString("\r\n")
	}
	return buffer.String()
}
