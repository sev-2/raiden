package generate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate application resource",
		Long:  "Generate deployment manifest and main function backend application",
		Run:   generateCmd,
	}
}

func generateCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Must be implementation soon :)")
}
