package fakeclock_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFakeClock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FakeClock Suite")
}
