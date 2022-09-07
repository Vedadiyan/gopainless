package main

import (
	gopainless "gopainless/internal"
	"os"
	"strings"
)

func main() {
	group, commands, options := parse()
	switch group {
	case "setup":
		{
			gopainless.Setup()
			break
		}
	case "init":
		{
			name := (*commands)["-N"]
			version := (*commands)["-V"]
			gopainless.PkgFileCreate(name[0], version[0])
			gopainless.ModFileCreate(name[0], "")
			break
		}
	case "install":
		{
			url := (*commands)["-U"]
			name := (*commands)["-N"]
			private := (*options)["--private"]
			recursive := (*options)["--recursive"]
			update := (*options)["--update"]
			gopainless.PkgFileLoad()
			gopainless.PkgAdd(url[0], name[0], private, update, recursive)
			gopainless.Write()
			break
		}
	case "remove":
		{
			name := (*commands)["-N"]
			gopainless.PkgDelete(name[0])
			gopainless.Write()
			break
		}
	case "restore":
		{
			tidy := (*options)["--tidy"]
			update := (*options)["--update"]
			gopainless.Clean()
			gopainless.PkgFileLoad()
			gopainless.PkgRestore(true, update)
			gopainless.Write()
			if tidy {
				gopainless.Tidy()
			}
			break
		}
	case "clean":
		{
			gopainless.Clean()
			break
		}
	case "build":
		{
			goos := (*commands)["-R"]
			goarch := (*commands)["-A"]
			output := (*commands)["-O"]
			target := (*commands)["-T"]
			gopainless.Build(goos[0], goarch[0], output[0], target[0])
			break
		}
	case "tidy":
		{
			gopainless.Tidy()
			break
		}
	}
}

func parse() (string, *map[string][]string, *map[string]bool) {
	commands := make(map[string][]string)
	options := make(map[string]bool)
	var group string
	var prev *string
	for i := 1; i < len(os.Args); i++ {
		val := os.Args[i]
		if strings.HasPrefix(val, "--") {
			options[val] = true
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
		commands[*prev] = append(commands[*prev], val)
	}
	return group, &commands, &options
}
