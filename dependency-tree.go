package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var gopath = ""

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	gopath = os.Getenv("GOPATH")

	modFile := path.Join(cwd, "go.mod")
	if _, err := os.Stat(modFile); os.IsNotExist(err) {
		println("ERROR: go.mod is not present in this directory, please only run this tool in the root of your go project")
		os.Exit(1)
	}

	getModuleList(getModuleName(cwd), "")

}

func getNameAndVersion(module string) (string, string) {
	if strings.Contains(module, "@") {
		s := strings.Split(module, "@")
		return s[0], s[1]
	}
	s := strings.Split(module, " ")
	if len(s) == 1 {
		return s[0], ""
	}
	return s[0], getSemVer(s[1])
}

func constructFilePath(dep string) (string, bool) {
	module, version := getNameAndVersion(dep)
	pkgPath := path.Join(gopath, "pkg", "mod", module+"@"+getSemVer(version))
	fullVersionPkgPath := path.Join(gopath, "pkg", "mod", module+"@"+version)
	srcPath := path.Join(gopath, "src", module)

	if _, err := os.Stat(srcPath); err == nil || !os.IsNotExist(err) {
		return srcPath, true
	}

	if _, err := os.Stat(pkgPath); err == nil || !os.IsNotExist(err) {
		return pkgPath, true
	}

	if _, err := os.Stat(fullVersionPkgPath); err == nil || !os.IsNotExist(err) {
		return fullVersionPkgPath, true
	}

	return "", false
}

func getModuleList(modPath, indent string) {
	rawPath, modFound := constructFilePath(escapeCapitalsInModuleName(modPath))
	if !modFound {
		fmt.Println(indent + strings.Split(modPath, " //")[0])
		return
	}
	modFilePath := path.Join(rawPath, "go.mod")
	fileBytes, err := ioutil.ReadFile(modFilePath)

	if err != nil {
		fmt.Println(indent + strings.Split(modPath, " //")[0])
		return
	}

	found := false

	lines := strings.Split(string(fileBytes), "\n")
	fmt.Println(indent + strings.Split(modPath, " //")[0] + ":")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !found {
			if line == "require (" {
				found = true
			}
		} else {
			if line == ")" {
				return
			} else {
				getModuleList(line, indent+"  ")
			}
		}
	}
}

func getSemVer(version string) string {
	re := regexp.MustCompile("(v\\d+\\.\\d+\\.\\d+)(-.*)*")
	match := re.FindStringSubmatch(version)
	if len(match) == 0 {
		return ""
	} else if len(match) == 1 {
		return match[0]
	}
	return match[1]
}

func getModuleName(cwd string) string {

	modFilePath := path.Join(cwd, "go.mod")
	fileBytes, err := ioutil.ReadFile(modFilePath)

	if err != nil {
		fmt.Println("Error reading go.mod: ", err)
		os.Exit(1)
	}

	lines := strings.Split(string(fileBytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modAddress := strings.Split(line, "module ")[1]
			var modName string
			if strings.HasSuffix(cwd, modAddress) {
				modName = modAddress
			} else {
				modName = modAddress + strings.Split(cwd, modAddress)[1]
			}
			return modName
		}
	}

	fmt.Println("Invalid go.mod, not module name")
	os.Exit(1)
	return ""
}

func escapeCapitalsInModuleName(name string) string {
	letters := strings.Split(name, "")
	newName := ""
	for _, letter := range letters {
		if strings.ToLower(letter) != letter {
			newName += "!" + strings.ToLower(letter)
		} else {
			newName += letter
		}
	}
	return newName
}
