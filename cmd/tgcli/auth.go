package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Telegram (phone + code)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			fmt.Println("🚧 Auth command - coming soon")
			fmt.Println("Will prompt for phone number and auth code")
			return nil
		},
	}

	return cmd
}
