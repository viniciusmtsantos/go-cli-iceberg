// Package vetdemo contains intentional vet bugs for demonstration purposes.
// This package is excluded from go test ./... to prevent build failures,
// but go vet ./... will still report the intentional bugs.
//
// To run tests normally (and skip this package):
//
//	go test ./...
//
// To see the vet issues:
//
//	go vet ./...
package vetdemo
