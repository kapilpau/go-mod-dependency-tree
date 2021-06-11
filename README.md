# Go Mod Dependency Tree

This package will print out the dependency tree for your go project

## Usage

To use this tool, make sure the binary is in your PATH. Call the CLI from the root of your go project:
```
go-tree
```
and it will recursively output the list of dependencies for each dependency module. Example output:
```
golang.org/x/crypto@v0.0.0-20200221231518-2aa609cf4a9d:
  golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
  golang.org/x/sys v0.0.0-20190412213103-97732733099d
```

## License

This tool is published under the MIT License