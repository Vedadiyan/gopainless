package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/gookit/color"
	gopainless "github.com/vedadiyan/gopainless/internal"
)

func main() {
	commands := New()
	commands.
		RegisterGroup("create", "Used for creating a new project based on an existing template").
		RegisterCommand("T", "Specifies the template url", nil).
		RegisterCommand("N", "Specifies the project name", nil)
	commands.
		RegisterGroup("setup", "Setups go-painless in the system")
	commands.
		RegisterGroup("init", "Initializes a new project").
		RegisterCommand("N", "Specifies project name", nil).
		RegisterCommand("V", "Specifies project version", nil)
	commands.
		RegisterGroup("add", "Adds a dependency").
		RegisterCommand("U", "Specifies dependency URL", nil).
		RegisterCommand("N", "Specifies dependency name", nil).
		RegisterFlag("private", "Used for installing from private repositories").
		RegisterFlag("recursive", "Used for recursively installing depdendencies").
		RegisterFlag("update", "Used for updating previously downloaded dependencies")
	commands.
		RegisterGroup("remove", "Removes an existing dependency").
		RegisterCommand("N", "Specifies dependency name", nil)
	commands.
		RegisterGroup("restore", "Restores dependencies in an existing project").
		RegisterFlag("tidy", "Runs go mod tidy after restoring the project").
		RegisterFlag("update", "Used for updating previously downloaded dependencies")
	commands.
		RegisterGroup("clear", "Removes go.mod and go.sum files")
	commands.
		RegisterGroup("publish", "Builds the project").
		RegisterCommand("R", "Specifies the runtime", nil).
		RegisterCommand("A", "Specifies build architecture", nil).
		RegisterCommand("O", "Specifies the output", nil).
		RegisterCommand("T", "Specifies the target build path", nil)
	commands.
		RegisterGroup("help", "Shows go-painless help")

	group, token, err := commands.Parse()
	if err != nil {
		if err == COMMAND_GROUP_NOT_FOUND {
			group = "go"
		} else {
			color.Hex(gopainless.RED).Println("Error:")
			fmt.Println()
			color.Hex(gopainless.RED).Println(err.Error())
			fmt.Println()
			color.Hex(gopainless.GREEN).Println("Help:")
			fmt.Println()
			color.Hex(gopainless.GREEN).Println(token.Help())
			return
		}
	}
	switch group {
	case "create":
		{
			templateName := token.GetMust("T")
			projectName := token.GetMust("N")
			gopainless.CreateFromTemplate(*templateName, *projectName)
			break
		}
	case "setup":
		{
			gopainless.Setup()
			break
		}
	case "init":
		{
			name := token.GetMust("N")
			version := token.GetMust("V")
			gopainless.PkgFileCreate(*name, *version)
			gopainless.ModFileCreate(*name, "")
			break
		}
	case "add":
		{
			url := token.GetMust("U")
			name := token.GetMust("N")
			private := token.GetFlag("private")
			recursive := token.GetFlag("recursive")
			update := token.GetFlag("update")
			gopainless.PkgFileLoad()
			gopainless.PkgAdd(*url, *name, private, update, recursive)
			gopainless.Write()
			break
		}
	case "remove":
		{
			name := token.GetMust("N")
			gopainless.PkgDelete(*name)
			gopainless.Write()
			break
		}
	case "restore":
		{
			tidy := token.GetFlag("tidy")
			update := token.GetFlag("update")
			gopainless.Clean()
			gopainless.PkgFileLoad()
			gopainless.PkgRestore(true, update)
			gopainless.Write()
			if tidy {
				gopainless.Tidy()
			}
			break
		}
	case "clear":
		{
			gopainless.Clean()
			break
		}
	case "publish":
		{
			goos := token.GetMust("R")
			goarch := token.GetMust("A")
			output := token.GetMust("O")
			target := token.GetMust("T")
			gopainless.Build(*goos, *goarch, *output, *target)
			break
		}
	case "tidy":
		{
			gopainless.Tidy()
			break
		}
	case "help":
		{
			fmt.Println("Go Painless Commands:")
			fmt.Println()
			color.Hex(gopainless.GREEN).Println(commands.Help())
			fmt.Println("Would you like to view original Go command help? (y/N)")
			reader := bufio.NewReader(os.Stdin)
			c, _, err := reader.ReadRune()
			if err != nil {
				panic(err)
			}
			if unicode.ToLower(c) == 'y' {
				fmt.Println("Original Go Commands")
				fmt.Println()
				gopainless.Run("go", "--help", nil)
			}

		}
	case "go":
		{
			gopainless.Run("go", strings.Join(os.Args[1:], " "), nil)
		}
	default:
		{
			panic("Invalid Command")
		}
	}

}
