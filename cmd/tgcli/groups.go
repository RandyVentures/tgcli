package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newGroupsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Group operations (list stored groups)",
	}

	cmd.AddCommand(newGroupsListCmd(flags))

	return cmd
}

func newGroupsListCmd(flags *rootFlags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List groups from local database",
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

			// Filter to groups only
			var groups []interface{}
			for _, chat := range chats {
				if chat.Type == "group" || chat.Type == "supergroup" {
					groups = append(groups, chat)
				}
			}

			if len(groups) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Println("No groups found. Groups are stored after receiving messages via 'tgcli sync'.")
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, groups)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTYPE\tTITLE")
			for _, chat := range chats {
				if chat.Type == "group" || chat.Type == "supergroup" {
					fmt.Fprintf(w, "%d\t%s\t%s\n", chat.ID, chat.Type, chat.Title)
				}
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "max groups to show")
	return cmd
}
