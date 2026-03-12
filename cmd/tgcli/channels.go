package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newChannelsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "Channel operations (list stored channels)",
	}

	cmd.AddCommand(newChannelsListCmd(flags))

	return cmd
}

func newChannelsListCmd(flags *rootFlags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List channels from local database",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			chats, err := a.Store().ListChats(ctx, limit)
			if err != nil {
				return fmt.Errorf("list chats: %w", err)
			}

			// Filter to channels only
			var channels []interface{}
			for _, chat := range chats {
				if chat.Type == "channel" {
					channels = append(channels, chat)
				}
			}

			if len(channels) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Println("No channels found. Channels are stored after receiving messages via 'tgcli sync'.")
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, channels)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE")
			for _, chat := range chats {
				if chat.Type == "channel" {
					fmt.Fprintf(w, "%d\t%s\n", chat.ID, chat.Title)
				}
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "max channels to show")
	return cmd
}
