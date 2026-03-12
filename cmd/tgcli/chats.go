package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func newChatsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chats",
		Short: "Chat operations (list, info)",
	}

	cmd.AddCommand(newChatsListCmd(flags))
	cmd.AddCommand(newChatsInfoCmd(flags))

	return cmd
}

func newChatsInfoCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get detailed chat information",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			chat, err := client.GetChat(struct{ ChatID int64 }{ChatID: chatID})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, chat)
			}

			fmt.Printf("Chat Info\n")
			fmt.Printf("=========\n")
			fmt.Printf("ID:          %d\n", chat.ID)
			fmt.Printf("Type:        %s\n", chat.Type)
			if chat.Title != "" {
				fmt.Printf("Title:       %s\n", chat.Title)
			}
			if chat.UserName != "" {
				fmt.Printf("Username:    @%s\n", chat.UserName)
			}
			if chat.FirstName != "" {
				fmt.Printf("First Name:  %s\n", chat.FirstName)
			}
			if chat.LastName != "" {
				fmt.Printf("Last Name:   %s\n", chat.LastName)
			}
			if chat.Description != "" {
				fmt.Printf("Description: %s\n", chat.Description)
			}

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newChatsListCmd(flags *rootFlags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all chats from local database",
		Long: `Lists chats stored in the local database.

Note: Chats are only stored after receiving messages via 'tgcli sync'.
Run 'tgcli sync --follow' to populate the database with incoming messages.`,
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

			if len(chats) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Println("No chats found. Run 'tgcli sync --follow' to receive messages.")
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, chats)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTYPE\tTITLE\tLAST MESSAGE")
			for _, chat := range chats {
				lastMsg := "never"
				if chat.LastMessageTS > 0 {
					lastMsg = formatTimeAgo(chat.LastMessageTS)
				}
				title := chat.Title
				if title == "" {
					title = "(no title)"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", chat.ID, chat.Type, title, lastMsg)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "max chats to show")
	return cmd
}

func formatTimeAgo(ts int64) string {
	t := time.Unix(ts, 0)
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
