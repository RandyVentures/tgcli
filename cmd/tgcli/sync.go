package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd(flags *rootFlags) *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync messages (requires prior auth)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			mode := "once"
			if follow {
				mode = "continuous"
			}
			fmt.Printf("🚧 Sync command (%s) - coming soon\n", mode)
			return nil
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "continuous sync (stay connected)")
	return cmd
}
