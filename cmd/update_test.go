package cmd

// import (
// 	"testing"

// 	"github.com/metal-stack/metalctl/cmd/completion"
// 	"github.com/spf13/afero"
// )

// func TestUpdateDoCommand(t *testing.T) {
// 	c := &config{fs: afero.NewMemMapFs(), out: nil, driverURL: "", comp: &completion.Completion{}, client: nil, log: nil, describePrinter: nil, listPrinter: nil} // Initialize your config here as needed
// 	test := newRootCmd(c)
// 	//updateCmd := newUpdateCmd(c)

// 	// Simulate "update do --version v0.15.3"
// 	test.SetArgs([]string{"update", "do", "--version", "v0.15.3"})

// 	// Execute the command
// 	err := test.Execute()
// 	if err != nil {
// 		t.Fatalf("updateCmd.Execute() failed: %v", err)
// 	}

// 	// Assert that desired version is set to "v0.15.3"
// 	// This step assumes you have a way to verify the version was set correctly,
// 	// such as checking a variable in your application that should be set based on the flag.
// 	// Since your current implementation directly uses Viper and updater.New(),
// 	// you might need to refactor to make it testable or mock certain parts.
// }
