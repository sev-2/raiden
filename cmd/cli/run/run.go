package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Deploy and run application",
		Long:  "Deploy resource , run backend application, and frontend application",
		Run:   runCmd,
	}
}

func runCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Must be implementation soon :)")
}
