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

func newMessagesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Message operations (list, search)",
	}

	cmd.AddCommand(newMessagesListCmd(flags))
	cmd.AddCommand(newMessagesSearchCmd(flags))

	return cmd
}

func newMessagesListCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages in a chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			messages, err := a.Store().ListMessages(chatID, limit)
			if err != nil {
				return wrapErr(err, "list messages")
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]interface{}{
					"messages": messages,
					"count":    len(messages),
					"chat_id":  chatID,
				})
			}

			// Human-readable output
			if len(messages) == 0 {
				fmt.Printf("No messages found in chat %d. Run 'tgcli sync' to fetch messages.\n", chatID)
				return nil
			}

			// Get chat info for title
			chat, err := a.Store().GetChat(chatID)
			if err == nil {
				fmt.Printf("Messages in %s (ID: %d)\n\n", chat.Title, chatID)
			} else {
				fmt.Printf("Messages in chat %d\n\n", chatID)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFrom\tDate\tText")
			fmt.Fprintln(w, "--\t----\t----\t----")

			for _, msg := range messages {
				from := fmt.Sprintf("%d", msg.FromUserID)
				if msg.FromUserID == 0 {
					from = "-"
				}

				date := time.Unix(msg.Date, 0).Format("Jan 2 15:04")

				text := msg.Text
				if len(text) > 60 {
					text = text[:57] + "..."
				}
				if text == "" && msg.MediaType != "" {
					text = fmt.Sprintf("[%s]", msg.MediaType)
				}

				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
					msg.ID,
					from,
					date,
					text,
				)
			}

			w.Flush()
			fmt.Printf("\nTotal: %d messages\n", len(messages))

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max messages to show")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newMessagesSearchCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search messages (FTS) - not implemented in Phase 1",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("message search not implemented in Phase 1")
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "search within specific chat (optional)")

	return cmd
}
