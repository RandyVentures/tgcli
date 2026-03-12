package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/out"
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

func newChatsListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all chats (DMs, groups, channels)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			chats, err := a.Store().ListChats()
			if err != nil {
				return wrapErr(err, "list chats")
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]interface{}{
					"chats": chats,
					"count": len(chats),
				})
			}

			// Human-readable output
			if len(chats) == 0 {
				fmt.Println("No chats found. Run 'tgcli sync' to fetch chats.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tType\tTitle\tUsername\tLast Message")
			fmt.Fprintln(w, "--\t----\t-----\t--------\t------------")

			for _, chat := range chats {
				username := chat.Username
				if username == "" {
					username = "-"
				}

				lastMsg := "-"
				if chat.LastMessageTs > 0 {
					lastMsg = formatTimestamp(chat.LastMessageTs)
				}

				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
					chat.ID,
					chat.Type,
					chat.Title,
					username,
					lastMsg,
				)
			}

			w.Flush()
			fmt.Printf("\nTotal: %d chats\n", len(chats))

			return nil
		},
	}
}

func newChatsInfoCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show chat details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			chat, err := a.Store().GetChat(chatID)
			if err != nil {
				return wrapErr(err, "get chat")
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, chat)
			}

			// Human-readable output
			fmt.Printf("Chat ID: %d\n", chat.ID)
			fmt.Printf("Type: %s\n", chat.Type)
			fmt.Printf("Title: %s\n", chat.Title)
			if chat.Username != "" {
				fmt.Printf("Username: @%s\n", chat.Username)
			}
			if chat.LastMessageTs > 0 {
				fmt.Printf("Last Message: %s\n", formatTimestamp(chat.LastMessageTs))
			}
			if chat.UnreadCount > 0 {
				fmt.Printf("Unread: %d\n", chat.UnreadCount)
			}

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func formatTimestamp(ts int64) string {
	t := time.Unix(ts, 0)
	now := time.Now()

	// If today, show time
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}

	// If this year, show date without year
	if t.Year() == now.Year() {
		return t.Format("Jan 2 15:04")
	}

	// Otherwise show full date
	return t.Format("2006-01-02 15:04")
}
