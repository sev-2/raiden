package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Deploy and run application",
		Long:  "Deploy resource , run backend application",
		Run:   runCmd,
	}
}

func runCmd(cmd *cobra.Command, args []string) {

	// prerequisite
	// - check configuration
	// - validate config file is exist
	// - marshall configuration

	// 1. compare resource
	// 2. re-generate main function app
	// 3. build binary
	// 4. run deployment
	// 5. run app
	fmt.Println("Must be implementation soon :)")
}
