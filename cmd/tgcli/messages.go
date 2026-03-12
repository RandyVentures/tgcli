package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/RandyVentures/tgcli/internal/store"
	"github.com/spf13/cobra"
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
	var beforeStr string
	var afterStr string
	var mediaType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages in a chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			// Parse time filters
			params := store.ListMessagesParams{
				ChatID:    chatID,
				Limit:     limit,
				MediaType: mediaType,
			}

			if beforeStr != "" {
				t, err := parseTimeFlag(beforeStr)
				if err != nil {
					return fmt.Errorf("invalid --before: %w", err)
				}
				params.Before = &t
			}
			if afterStr != "" {
				t, err := parseTimeFlag(afterStr)
				if err != nil {
					return fmt.Errorf("invalid --after: %w", err)
				}
				params.After = &t
			}

			messages, err := a.Store().ListMessages(ctx, params)
			if err != nil {
				return fmt.Errorf("list messages: %w", err)
			}

			if len(messages) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Println("No messages found for this chat.")
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, messages)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFROM\tDATE\tTEXT")
			for _, msg := range messages {
				text := msg.Text
				if len(text) > 50 {
					text = text[:47] + "..."
				}
				date := time.Unix(msg.Date, 0).Format("01/02 15:04")
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\n", msg.ID, msg.FromUserID, date, text)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max messages to show")
	cmd.Flags().StringVar(&beforeStr, "before", "", "messages before this time (RFC3339 or Unix timestamp)")
	cmd.Flags().StringVar(&afterStr, "after", "", "messages after this time (RFC3339 or Unix timestamp)")
	cmd.Flags().StringVar(&mediaType, "media-type", "", "filter by media type")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newMessagesSearchCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var limit int
	var beforeStr string
	var afterStr string
	var mediaType string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search messages (FTS if available, otherwise LIKE)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			// Parse time filters
			params := store.SearchMessagesParams{
				Query:     query,
				ChatID:    chatID,
				Limit:     limit,
				MediaType: mediaType,
			}

			if beforeStr != "" {
				t, err := parseTimeFlag(beforeStr)
				if err != nil {
					return fmt.Errorf("invalid --before: %w", err)
				}
				params.Before = &t
			}
			if afterStr != "" {
				t, err := parseTimeFlag(afterStr)
				if err != nil {
					return fmt.Errorf("invalid --after: %w", err)
				}
				params.After = &t
			}

			messages, err := a.Store().SearchMessages(ctx, params)
			if err != nil {
				return fmt.Errorf("search messages: %w", err)
			}

			if len(messages) == 0 {
				if flags.asJSON {
					return writeJSON(os.Stdout, []interface{}{})
				}
				fmt.Printf("No messages found matching '%s'\n", query)
				return nil
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, messages)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tCHAT\tDATE\tMATCH")
			for _, msg := range messages {
				// Use snippet if available (from FTS), otherwise truncate text
				displayText := msg.Text
				if msg.Snippet != "" {
					displayText = msg.Snippet
				}
				if len(displayText) > 60 {
					displayText = displayText[:57] + "..."
				}
				date := time.Unix(msg.Date, 0).Format("01/02 15:04")
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\n", msg.ID, msg.ChatID, date, displayText)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "search within specific chat (optional)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max results")
	cmd.Flags().StringVar(&beforeStr, "before", "", "messages before this time (RFC3339 or Unix timestamp)")
	cmd.Flags().StringVar(&afterStr, "after", "", "messages after this time (RFC3339 or Unix timestamp)")
	cmd.Flags().StringVar(&mediaType, "media-type", "", "filter by media type")

	return cmd
}
