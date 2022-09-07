package gopainless

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/gookit/color"
)

type Package struct {
	Name     string                  `json:"Name"`
	Version  string                  `json:"Version"`
	Packages map[string]PackageValue `json:"Packages"`
}

type PackageValue struct {
	URI     string `json:"Uri"`
	Private bool   `json:"Private"`
}

type OperatingSystems string

const (
	LINUX   OperatingSystems = "linux"
	WINDOWS OperatingSystems = "windows"
	MAC     OperatingSystems = "mac"
)

const packageManagementFileName string = "package.json"

var (
	gopackage          Package
	opearingSystem     OperatingSystems
	goPainlessFileName string
	homeDirectory      string
	packageDirectory   string
)

func init() {
	os := runtime.GOOS
	switch os {
	case "linux":
		{
			opearingSystem = LINUX
			goPainlessFileName = "go-painless"
			break
		}
	case "windows":
		{
			opearingSystem = WINDOWS
			goPainlessFileName = "go-painless.exe"
			break
		}
	case "darwin":
		{
			opearingSystem = MAC
			goPainlessFileName = "go-painless.dmg"
			break
		}
	default:
		{
			panic("go-painless does not support the current platform")
		}
	}
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	homeDirectory = usr.HomeDir
	packageDirectory = fmt.Sprintf("%s/%s/%s", homeDirectory, "go-painless", "packages")
}

func Setup() {
	path := fmt.Sprintf("%s/%s/%s", homeDirectory, "go-painless", "bin")
	if os.Args[0] == fmt.Sprintf("%s/%s", path, goPainlessFileName) {
		color.Hex("#00ff5f").Println("This version of go-painless has already been setup")
		return
	}
	exists, err := Exists(path)
	if err != nil {
		panic(err)
	}
	if *exists == false {
		os.MkdirAll(path, 777)
	} else {
		os.Remove(fmt.Sprintf("%s/%s", path, goPainlessFileName))
	}
	src, err := os.Open(os.Args[0])
	if err != nil {
		panic(err)
	}
	dest, err := os.Create(fmt.Sprintf("%s/%s", path, goPainlessFileName))
	if err != nil {
		panic(err)
	}
	io.Copy(dest, src)
	src.Close()
	dest.Close()
	color.Hex("#00ff5f").Println("go-painless setup successfully")
}

func PkgFileNew(name string, version string) {
	if len(name) == 0 {
		panic("name")
	}
	if len(version) == 0 {
		panic("version")
	}
	gopackage = Package{}
	gopackage.Name = name
	gopackage.Version = version
}

func PkgFileLoad() {
	goPackageJson, err := os.ReadFile(packageManagementFileName)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(goPackageJson, &gopackage)
	if err != nil {
		panic(err)
	}
}

func ModFileCreate(name string, workingDirectory string) {
	err := Run("go", fmt.Sprintf("mod init %s", name), workingDirectory)
	if err != nil {
		panic("Could not create mod file")
	}
}

func Exists(filePath string) (*bool, error) {
	_, err := os.Stat(filePath)
	var output bool
	if err == nil {
		output = true
		return &output, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		output = false
		return &output, nil
	}
	return nil, err
}

func PkgFileCreate(name string, version string) {
	exists, err := Exists(packageManagementFileName)
	if err != nil {
		panic(err)
	}
	if *exists == true {
		panic("package.json file already exists")
	}
	if *exists == false {
		gopackage = Package{}
		gopackage.Name = name
		gopackage.Version = version
		gopackage.Packages = map[string]PackageValue{}
		goPackageJson, err := json.Marshal(&gopackage)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(packageManagementFileName, goPackageJson, 777)
		if err != nil {
			panic(err)
		}
	}
}

func PkgAdd(uri string, name string, private bool, update bool, recursive bool) {
	goPackage, ok := gopackage.Packages[name]
	if ok {
		panic("Another package with the same name already exists")
	}
	if private {
		err := getPackage(uri)
		if err != nil {
			panic(err)
		}
		goPackage = PackageValue{}
		goPackage.URI = uri
		goPackage.Private = private
		gopackage.Packages[name] = goPackage
		return
	}
	packagePath := fmt.Sprintf("%s/%s", packageDirectory, name)
	exists, err := Exists(packagePath)
	if err != nil {
		panic(err)
	}
	if update && *exists == true {
		color.Hex("#00ff5f").Println("Deleting", packagePath)
		err = _delete(packagePath)
		if err != nil {
			panic(err)
		}
		*exists = false
	}
	if *exists == false {
		err := getPrivatePackage(uri, name, recursive)
		if err != nil {
			panic(err)
		}
		goPackage = PackageValue{}
		goPackage.URI = uri
		goPackage.Private = private
		gopackage.Packages[name] = goPackage
		err = Run("go", "mod tidy", packagePath)
		if err != nil {
			panic(err)
		}
		return
	}
	goPackage = PackageValue{}
	goPackage.URI = uri
	goPackage.Private = private
	gopackage.Packages[name] = goPackage
}

func PkgDelete(name string) bool {
	_, ok := gopackage.Packages[name]
	if ok {
		delete(gopackage.Packages, name)
		return true
	}
	return false
}

func PkgRestore(recursive bool, update bool) {
	ModFileCreate(gopackage.Name, "")
	for key, value := range gopackage.Packages {
		if !value.Private {
			err := getPackage(value.URI)
			if err != nil {
				panic(err)
			}
			continue
		}
		packagePath := fmt.Sprintf("%s/%s", packageDirectory, key)
		exists, err := Exists(packagePath)
		if err != nil {
			panic(err)
		}
		if *exists == true {
			if !update {
				continue
			}
			color.HEX("#00ff5f").Println("Deleting", packagePath)
			err = _delete(packagePath)
			if err != nil {
				panic(err)
			}
			*exists = false
		}
		if *exists == false {
			err := getPrivatePackage(value.URI, key, false)
			exists, err := Exists(fmt.Sprintf("%s/%s", packagePath, "package.json"))
			if err != nil {
				panic(err)
			}
			if *exists == true {
				ModFileCreate(key, packagePath)
				Run(fmt.Sprintf("%s/go-painless/bin/%s", homeDirectory, goPainlessFileName), "restore", fmt.Sprintf("%s/%s", packagePath, key))
			}
		}
		Run("go", "mod tidy", packagePath)
	}
}

func Write() {
	modFile, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}
	buffer := make([]*string, 0)
	output := bytes.NewBufferString("")
	for _, line := range strings.Split(string(modFile), "\n") {
		tmp := strings.TrimPrefix(line, " ")
		if strings.HasPrefix(tmp, "replace") {
			split := strings.Split(tmp, "=>")
			if len(split) != 2 {
				panic("Malformed go.mod file")
			}
			buffer = append(buffer, &strings.Split(split[0], " ")[1])
			continue
		}
		var _index *int
		for index, value := range buffer {
			if strings.Contains(line, *value) {
				_index = &index
				break
			}
		}
		if _index != nil {
			buffer[*_index] = nil
			continue
		}
		output.WriteString(line)
		output.WriteString("\n")
	}
	for key, value := range gopackage.Packages {
		if !value.Private {
			continue
		}
		path := fmt.Sprintf("%s/%s", packageDirectory, key)
		output.WriteString(fmt.Sprintf("replace %s => \"%s\"", key, strings.ReplaceAll(path, "\\", "\\\\")))
		output.WriteString("\n")
		output.WriteString(fmt.Sprintf("require %s v1.0.0", key))
	}
	err = os.WriteFile("go.mod", output.Bytes(), 777)
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(&gopackage)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(packageManagementFileName, json, 777)
	if err != nil {
		panic(err)
	}
}

func Tidy() {
	Run("go", "mod tidy", "")
}

func Build(goos string, goarch string, output string, target string) {
	os.Setenv("GOOS", goos)
	os.Setenv("GOARCH", goarch)
	Run("go", fmt.Sprintf("build -o %s %s", output, target), "")
}

func Clean() {
	modFileExists, err := Exists("go.mod")
	if err != nil {
		panic(err)
	}
	if *modFileExists == true {
		_delete("go.mod")
	}
	sumFileExists, err := Exists("go.sum")
	if err != nil {
		panic(err)
	}
	if *sumFileExists == true {
		_delete("go.sum")
	}
}

func getPackage(url string) error {
	return Run("go", fmt.Sprintf("get %s", url), "")
}
func getPrivatePackage(url string, name string, recursive bool) error {
	packageDirectoryExists, err := Exists(packageDirectory)
	if err != nil {
		return err
	}
	if *packageDirectoryExists == false {
		os.MkdirAll(packageDirectory, 777)
	}
	err = Run("git", fmt.Sprintf("clone %s %s", url, name), packageDirectory)
	if err != nil {
		return err
	}
	if recursive {
		// Check this line later
		ModFileCreate(name, fmt.Sprintf("%s/%s", packageDirectory, name))
		Run(fmt.Sprintf("%s/go-painless/bin/%s", homeDirectory, goPainlessFileName), "restore", fmt.Sprintf("%s/%s", packageDirectory, name))
	}
	return nil
}
func Run(cmd string, args string, workingDirectory string) error {
	_cmd := exec.Command(cmd, strings.Split(args, " ")...)
	if len(workingDirectory) != 0 {
		_cmd.Dir = workingDirectory
	}
	var outb, errb bytes.Buffer
	_cmd.Stdout = &outb
	_cmd.Stderr = &errb
	err := _cmd.Run()
	if err != nil {
		return err
	}
	if errb.Len() > 0 {
		color.HEX("#ffef00").Println(errb.String())
	}
	if outb.Len() > 0 {
		color.Hex("#00ff5f").Println(outb.String())
	}
	return nil
}

func _delete(packagePath string) error {
	_path, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	files, err := _path.Readdir(-1)
	if err != nil {
		return err
	}
	for _, file := range files {
		name := fmt.Sprintf("%s/%s", packagePath, file.Name())
		err = os.Chmod(name, 777)
		if err != nil {
			return err
		}
		err = os.RemoveAll(name)
		if err != nil {
			return err
		}
	}
	return nil
}