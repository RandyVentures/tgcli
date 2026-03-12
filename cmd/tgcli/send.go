package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/out"
)

func newSendCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send messages (text, file, reaction)",
	}

	cmd.AddCommand(newSendTextCmd(flags))
	cmd.AddCommand(newSendFileCmd(flags))
	cmd.AddCommand(newSendReactionCmd(flags))

	return cmd
}

func newSendTextCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var message string
	var replyTo int

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Send text message",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			tgClient := a.TGClient()
			if tgClient == nil {
				return fmt.Errorf("telegram client not initialized")
			}

			// Check if authenticated
			authed, err := tgClient.IsAuthed(ctx)
			if err != nil {
				return wrapErr(err, "check auth status")
			}
			if !authed {
				return fmt.Errorf("not authenticated. Run 'tgcli auth' first")
			}

			// Send message
			msgID, err := tgClient.SendTextMessage(ctx, to, message, replyTo)
			if err != nil {
				return wrapErr(err, "send message failed")
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]interface{}{
					"status":     "sent",
					"message_id": msgID,
					"chat_id":    to,
				})
			}

			fmt.Printf("✅ Message sent (ID: %d)\n", msgID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&message, "message", "", "message text (required)")
	cmd.Flags().IntVar(&replyTo, "reply-to", 0, "reply to message ID (optional)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

func newSendFileCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var file string
	var caption string

	cmd := &cobra.Command{
		Use:   "file",
		Short: "Send file (not implemented in Phase 1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("file sending not implemented in Phase 1")
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "file path (required)")
	cmd.Flags().StringVar(&caption, "caption", "", "file caption (optional)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newSendReactionCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var messageID int
	var emoji string

	cmd := &cobra.Command{
		Use:   "reaction",
		Short: "Send reaction to message (not implemented in Phase 1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("reactions not implemented in Phase 1")
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&messageID, "message-id", 0, "message ID (required)")
	cmd.Flags().StringVar(&emoji, "emoji", "", "emoji reaction (required)")
	_ = cmd.MarkFlagRequired("chat")
	_ = cmd.MarkFlagRequired("message-id")
	_ = cmd.MarkFlagRequired("emoji")

	return cmd
}
