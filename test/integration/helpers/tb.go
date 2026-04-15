//go:build integration

package helpers

// TB is a subset of testing.TB compatible with both *testing.T and Ginkgo's GinkgoTInterface.
// Use this instead of testing.TB in all helper signatures so helpers work from Ginkgo specs.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
	Logf(format string, args ...any)
}
