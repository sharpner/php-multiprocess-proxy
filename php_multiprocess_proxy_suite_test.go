package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPhpMultiprocessProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PhpMultiprocessProxy Suite")
}
