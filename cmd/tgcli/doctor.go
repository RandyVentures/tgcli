package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Show diagnostics (session status, DB stats)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			fmt.Println("🚧 Doctor command - coming soon")
			fmt.Println("Will show session status, DB stats, health checks")
			return nil
		},
	}
}
