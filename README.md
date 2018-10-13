# monkey [![Build Status](https://travis-ci.org/bradleyjkemp/monkey.svg?branch=master)](https://travis-ci.org/bradleyjkemp/monkey) [![Coverage Status](https://coveralls.io/repos/github/bradleyjkemp/monkey/badge.svg)](https://coveralls.io/github/bradleyjkemp/monkey?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/bradleyjkemp/monkey)](https://goreportcard.com/report/github.com/bradleyjkemp/monkey) [![GoDoc](https://godoc.org/github.com/bradleyjkemp/monkey?status.svg)](https://godoc.org/github.com/bradleyjkemp/monkey)

Ever wanted to tamper with the value of an unexported field?
This library lets you create a small patch value which can be applied to modify unexported fields regardless of how deeply nested they are.

## Usage

Say you have a struct like so:
```go
type CoolLibrary struct {
	usefulInternalLogs io.Writer
	...
	...
}
```
But you want access to the internal logs which can't be configured to write to a location of your choice.

You could fork the library and add this feature or just use Monkey to patch this field:
```go
lib := CoolLibraryConstructor()

coolLibraryPatch := struct {
	usefulInternalLogs io.Writer
}{
	os.Stdout
}

monkey.Patch(&lib, &coolLibraryPatch)
```

Now your useful internal logs will be written to os.Stdout!

