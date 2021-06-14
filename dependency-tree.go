package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var gopath = ""
var maxDepth = flag.Int("maxDepth", -1, "Maximum recursion level to scan, -1 for no limit, otherwise must be an integer greater than 0, ignored if -find specified. Defaults to -1.")
var modulePath = flag.String("modulePath", ".", "Path to module to scan, can be relative or absolute. Defaults to current working directory.")
var searchText = flag.String("find", "", "Search for a specific module. Useful for if you're looking for the dependency chain for a specific module. If not set, the program will print out the entire tree.")

type dependencyChain struct {
	module   string
	children []dependencyChain
}

func main() {
	flag.Parse()

	cwd := *modulePath

	if cwd == "." {
		dir, err := os.Getwd()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		cwd = dir
	} else {
		if !path.IsAbs(*modulePath) {

			dir, err := os.Getwd()
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			cwd = path.Join(dir, *modulePath)
		}
	}

	gopath = os.Getenv("GOPATH")

	modFile := path.Join(cwd, "go.mod")
	if _, err := os.Stat(modFile); os.IsNotExist(err) {
		println("ERROR: go.mod is not present in this directory, please only run this tool in the root of your go project or specify a path to the root directory of a go project")
		os.Exit(1)
	}

	if *searchText != "" {
		chain := search(cwd)
		if len(chain.children) != 0 {
			printChain(chain, "")
		} else {
			fmt.Println("Unable to find module '" + *searchText + "' in dependency tree.")
		}
	} else {
		getModuleList(getModuleName(cwd), "", *maxDepth)
	}

	os.Exit(0)
}

func printChain(chain dependencyChain, indent string) {
	if len(chain.children) == 0 {
		fmt.Println(indent + chain.module)
	} else {
		fmt.Println(indent + chain.module + ":")
		for _, child := range chain.children {
			printChain(child, indent+"  ")
		}
	}
}

func search(cwd string) dependencyChain {
	fmt.Println("Searching for " + *searchText)

	return dependencyChain{
		module:   getModuleName(cwd),
		children: rescursiveFind(getModuleName(cwd)),
	}
}

func rescursiveFind(module string) []dependencyChain {
	children := make([]dependencyChain, 0)
	rawPath, modFound := constructFilePath(escapeCapitalsInModuleName(module))

	if !modFound {
		return children
	}

	modFilePath := path.Join(rawPath, "go.mod")
	fileBytes, err := ioutil.ReadFile(modFilePath)

	if err != nil {
		return children
	}

	found := false

	lines := strings.Split(string(fileBytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !found {
			if line == "require (" {
				found = true
			}
		} else {
			if line == "" {
				// skip
			} else if strings.Split(line, " ")[0] == *searchText {
				children = append(children, dependencyChain{
					module:   line,
					children: make([]dependencyChain, 0),
				})
			} else if line == ")" {
				return children
			} else {
				chain := rescursiveFind(line)
				if len(chain) > 0 {
					children = append(children, dependencyChain{
						module:   line,
						children: chain,
					})
				}
			}
		}
	}
	return children
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

func getModuleList(modPath, indent string, depth int) {
	if depth == 0 {
		fmt.Println(indent + strings.Split(modPath, " //")[0])
		return
	}
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
			if line == "" {
				// skip
			} else if line == ")" {
				return
			} else {
				getModuleList(line, indent+"  ", depth-1)
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
