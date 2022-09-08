package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Instruction struct {
	must     map[string]*string
	optional map[string]string
	flags    map[string]bool
	help     map[string]string
}

type Command struct {
	commands map[string]Instruction
	help     map[string]string
}

func New() *Command {
	command := &Command{}
	command.commands = make(map[string]Instruction)
	command.help = make(map[string]string)
	return command
}

func (instructions Instruction) GetMust(command string) *string {
	return instructions.must[command]
}

func (instructions Instruction) GetOptional(command string) string {
	return instructions.optional[command]
}

func (instructions Instruction) GetFlag(command string) bool {
	return instructions.flags[command]
}

func (instruction Instruction) PrintHelp() {
	longest := 0
	sortedKeys := make([]string, 0)
	for key, value := range instruction.help {
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
	for _, key := range sortedKeys {
		_, isMust := instruction.must[key]
		_, isOptional := instruction.must[key]
		if isMust || isOptional {
			fmt.Printf("-")
			fmt.Print(key)
			for i := 0; i < (longest-len(key)-1)+10; i++ {
				fmt.Print(" ")
			}
		} else if _, ok := instruction.flags[key]; ok {
			fmt.Printf("--")
			fmt.Print(key)
			for i := 0; i < (longest-len(key)-2)+10; i++ {
				fmt.Print(" ")
			}
		} else {
			panic("Unknown Case")
		}
		fmt.Println(instruction.help[key])
	}
}

func (instructions *Instruction) RegisterCommand(cmd string, help string, defaultValue *string) *Instruction {
	if defaultValue != nil {
		instructions.optional[cmd] = *defaultValue
	} else {
		instructions.must[cmd] = nil
	}
	instructions.help[cmd] = help
	return instructions
}
func (instructions *Instruction) RegisterFlag(cmd string, help string) *Instruction {
	instructions.flags[cmd] = false
	instructions.help[cmd] = help
	return instructions
}
func (command *Command) RegisterGroup(group string, help string) *Instruction {
	instructions := Instruction{}
	instructions.must = make(map[string]*string)
	instructions.optional = make(map[string]string)
	instructions.flags = make(map[string]bool)
	instructions.help = make(map[string]string)
	command.commands[group] = instructions
	command.help[group] = help
	return &instructions
}
func (command *Command) Parse() (string, *Instruction, error) {
	commands := make(map[string]string)
	flags := make(map[string]bool)
	var group string
	var prev *string
	for i := 1; i < len(os.Args); i++ {
		val := os.Args[i]
		if strings.HasPrefix(val, "--") {
			flags[val] = true
			continue
		}
		if strings.HasPrefix(val, "-") {
			prev = &val
			continue
		}
		if i == 1 {
			group = val
			continue
		}
		if prev == nil {
			panic("Invalid Command Line Argument")
		}
		commands[*prev] = val
		prev = nil
	}
	value, ok := command.commands[group]
	if !ok {
		command.PrintHelp()
		return "", nil, errors.New("Command group not found")
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
	for key, val := range value.flags {
		_ = val
		_ = val
		_, ok := commands[key]
		if ok {
			value.flags[key] = true
		}
	}
	if len(errs) != 0 {
		value.PrintHelp()
		return "", nil, errors.New(strings.Join(errs, "\r\n"))
	}
	return group, &value, nil
}

func (command Command) PrintHelp() {
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
	for _, key := range sortedKeys {
		fmt.Print(key)
		for i := 0; i < (longest-len(key))+10; i++ {
			fmt.Print(" ")
		}
		fmt.Println(command.help[key])
	}
}
