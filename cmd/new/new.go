package new

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Setup new app",
		Long:  "Create project folder and scaffold application",
		Run:   newCmd,
	}
}

func newCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Must be implementation soon :)")
}
