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

func main()  {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	gopath = os.Getenv("GOPATH")

	modFile := path.Join(cwd, "go.mod")
	if _, err := os.Stat(modFile); os.IsNotExist(err) {
		println("ERROR: go.mod is not present in this directory, please only run this tool in the root of your go project")
	}

	fmt.Println(cwd + ":")
	getModuleList(cwd, "  ")

}

func getNameAndVersion(module string) (string, string) {

	s := strings.Split(module, " ")
	if len(s) == 1 {
		return s[0], ""
	}
	return s[0], getSemVer(s[1])
}

func constructFilePath(dep string) string {
	module, version := getNameAndVersion(dep)
	pkgPath := path.Join(gopath, "pkg", "mod", module + "@" + getSemVer(version))
	srcPath := path.Join(gopath, "src", module)

	if _, err := os.Stat(srcPath); err == nil || !os.IsNotExist(err) {
		return srcPath
	}

	if _, err := os.Stat(pkgPath); err == nil || !os.IsNotExist(err) {
		return pkgPath
	}

	return ""
}

func getModuleList(modPath, indent string) {
	fmt.Println("Looing at:" + modPath)
	modFilePath := path.Join(constructFilePath(modPath), "go.mod")
	fileBytes, err := ioutil.ReadFile(modFilePath)

	if err != nil {
		return
	}

	found := false

	lines := strings.Split(string(fileBytes), "\n")
	fmt.Println(lines)
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
				name, _ := getNameAndVersion(line)
				fmt.Println(name + ":")
				getModuleList(line, indent + "  ")
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

